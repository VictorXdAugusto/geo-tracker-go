package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/vitao/geolocation-tracker/internal/domain/entity"
	"github.com/vitao/geolocation-tracker/internal/domain/repository"
	"github.com/vitao/geolocation-tracker/internal/domain/valueobject"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// positionRepository implementa repository.PositionRepository usando PostgreSQL + PostGIS
type positionRepository struct {
	db     *DB
	logger logger.Logger
}

// NewPositionRepository cria uma nova instância do repository de posições
func NewPositionRepository(db *DB, logger logger.Logger) repository.PositionRepository {
	return &positionRepository{
		db:     db,
		logger: logger,
	}
}

// Save persiste uma posição
func (r *positionRepository) Save(ctx context.Context, position *entity.Position) error {
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Extrair valores para evitar problemas com métodos
	posID := position.ID()
	userID := position.UserID()

	// 1. Inserir na tabela positions (histórico)
	insertPosition := `
		INSERT INTO positions (id, user_id, location, sector_x, sector_y, created_at)
		VALUES ($1, $2, ST_GeomFromText($3, 4326), $4, $5, $6)
	`

	_, err = tx.ExecContext(ctx, insertPosition,
		posID.Value(),
		userID.Value(),
		position.Coordinate().ToWKT(),
		position.SectorX(),
		position.SectorY(),
		position.RecordedAt().Time(),
	)

	if err != nil {
		r.logger.Error("Failed to insert position",
			"position_id", posID.Value(),
			"user_id", userID.Value(),
			"error", err,
		)
		return fmt.Errorf("failed to insert position: %w", err)
	}

	// 2. Atualizar/inserir posição atual
	if err := r.updateCurrentPosition(ctx, tx, position); err != nil {
		return fmt.Errorf("failed to update current position: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	r.logger.Debug("Position saved successfully",
		"position_id", posID.Value(),
		"user_id", userID.Value(),
	)

	return nil
}

// updateCurrentPosition atualiza a tabela current_positions
func (r *positionRepository) updateCurrentPosition(ctx context.Context, tx *sql.Tx, position *entity.Position) error {
	posID := position.ID()
	userID := position.UserID()

	upsertCurrent := `
		INSERT INTO current_positions (user_id, position_id, location, sector_x, sector_y, updated_at)
		VALUES ($1, $2, ST_GeomFromText($3, 4326), $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET
			position_id = EXCLUDED.position_id,
			location = EXCLUDED.location,
			sector_x = EXCLUDED.sector_x,
			sector_y = EXCLUDED.sector_y,
			updated_at = EXCLUDED.updated_at
	`

	_, err := tx.ExecContext(ctx, upsertCurrent,
		userID.Value(),
		posID.Value(),
		position.Coordinate().ToWKT(),
		position.SectorX(),
		position.SectorY(),
		position.RecordedAt().Time(),
	)

	return err
}

// FindByID busca posição por ID
func (r *positionRepository) FindByID(ctx context.Context, id entity.PositionID) (*entity.Position, error) {
	query := `
		SELECT id, user_id, ST_X(location), ST_Y(location), sector_x, sector_y, created_at
		FROM positions
		WHERE id = $1
	`

	var posID, userID string
	var lat, lng float64
	var sectorX, sectorY int
	var createdAt time.Time

	err := r.db.Connection().QueryRowContext(ctx, query, id.Value()).Scan(
		&posID, &userID, &lng, &lat, &sectorX, &sectorY, &createdAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("position not found: %s", id.Value())
		}
		return nil, fmt.Errorf("failed to find position %s: %w", id.Value(), err)
	}

	return r.scanToPosition(posID, userID, lat, lng, createdAt)
}

// FindCurrentByUserID busca posição atual de um usuário
func (r *positionRepository) FindCurrentByUserID(ctx context.Context, userID entity.UserID) (*entity.Position, error) {
	query := `
		SELECT p.id, p.user_id, ST_X(p.location), ST_Y(p.location), p.sector_x, p.sector_y, p.created_at
		FROM positions p
		INNER JOIN current_positions cp ON p.id = cp.position_id
		WHERE cp.user_id = $1
	`

	var posID, posUserID string
	var lat, lng float64
	var sectorX, sectorY int
	var createdAt time.Time

	err := r.db.Connection().QueryRowContext(ctx, query, userID.Value()).Scan(
		&posID, &posUserID, &lng, &lat, &sectorX, &sectorY, &createdAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("current position not found for user: %s", userID.Value())
		}
		return nil, fmt.Errorf("failed to find current position for user %s: %w", userID.Value(), err)
	}

	return r.scanToPosition(posID, posUserID, lat, lng, createdAt)
}

// FindHistoryByUserID busca histórico de posições de um usuário
func (r *positionRepository) FindHistoryByUserID(ctx context.Context, userID entity.UserID, limit int) ([]*entity.Position, error) {
	query := `
		SELECT id, user_id, ST_X(location), ST_Y(location), sector_x, sector_y, created_at
		FROM positions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Connection().QueryContext(ctx, query, userID.Value(), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find position history for user %s: %w", userID.Value(), err)
	}
	defer rows.Close()

	positions := make([]*entity.Position, 0)

	for rows.Next() {
		var posID, posUserID string
		var lat, lng float64
		var sectorX, sectorY int
		var createdAt time.Time

		if err := rows.Scan(&posID, &posUserID, &lng, &lat, &sectorX, &sectorY, &createdAt); err != nil {
			r.logger.Error("Failed to scan position row", "error", err)
			continue
		}

		position, err := r.scanToPosition(posID, posUserID, lat, lng, createdAt)
		if err != nil {
			r.logger.Error("Failed to reconstruct position", "position_id", posID, "error", err)
			continue
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// FindNearby busca posições próximas usando PostGIS
func (r *positionRepository) FindNearby(ctx context.Context, coord *valueobject.Coordinate, radiusMeters float64, limit int) ([]*entity.Position, error) {
	query := `
		SELECT p.id, p.user_id, ST_X(p.location), ST_Y(p.location), p.sector_x, p.sector_y, p.created_at,
			   ST_Distance(p.location::geography, ST_GeomFromText($1, 4326)::geography) as distance
		FROM positions p
		INNER JOIN current_positions cp ON p.id = cp.position_id
		WHERE ST_DWithin(p.location::geography, ST_GeomFromText($1, 4326)::geography, $2)
		ORDER BY distance
		LIMIT $3
	`

	rows, err := r.db.Connection().QueryContext(ctx, query, coord.ToWKT(), radiusMeters, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find nearby positions: %w", err)
	}
	defer rows.Close()

	positions := make([]*entity.Position, 0)

	for rows.Next() {
		var posID, userID string
		var lat, lng float64
		var sectorX, sectorY int
		var createdAt time.Time
		var distance float64

		if err := rows.Scan(&posID, &userID, &lng, &lat, &sectorX, &sectorY, &createdAt, &distance); err != nil {
			r.logger.Error("Failed to scan nearby position row", "error", err)
			continue
		}

		position, err := r.scanToPosition(posID, userID, lat, lng, createdAt)
		if err != nil {
			r.logger.Error("Failed to reconstruct nearby position", "position_id", posID, "error", err)
			continue
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// FindInSector busca posições em um setor específico
func (r *positionRepository) FindInSector(ctx context.Context, sector *valueobject.Sector) ([]*entity.Position, error) {
	query := `
		SELECT p.id, p.user_id, ST_X(p.location), ST_Y(p.location), p.sector_x, p.sector_y, p.created_at
		FROM positions p
		INNER JOIN current_positions cp ON p.id = cp.position_id
		WHERE p.sector_x = $1 AND p.sector_y = $2
	`

	rows, err := r.db.Connection().QueryContext(ctx, query, sector.X(), sector.Y())
	if err != nil {
		return nil, fmt.Errorf("failed to find positions in sector %s: %w", sector.ID(), err)
	}
	defer rows.Close()

	positions := make([]*entity.Position, 0)

	for rows.Next() {
		var posID, userID string
		var lat, lng float64
		var sectorX, sectorY int
		var createdAt time.Time

		if err := rows.Scan(&posID, &userID, &lng, &lat, &sectorX, &sectorY, &createdAt); err != nil {
			r.logger.Error("Failed to scan sector position row", "error", err)
			continue
		}

		position, err := r.scanToPosition(posID, userID, lat, lng, createdAt)
		if err != nil {
			r.logger.Error("Failed to reconstruct sector position", "position_id", posID, "error", err)
			continue
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// FindInSectors busca posições em múltiplos setores
func (r *positionRepository) FindInSectors(ctx context.Context, sectors []*valueobject.Sector) ([]*entity.Position, error) {
	if len(sectors) == 0 {
		return []*entity.Position{}, nil
	}

	// Construir query dinâmica com placeholders
	query := `
		SELECT p.id, p.user_id, ST_X(p.location), ST_Y(p.location), p.sector_x, p.sector_y, p.created_at
		FROM positions p
		INNER JOIN current_positions cp ON p.id = cp.position_id
		WHERE (p.sector_x, p.sector_y) IN (
	`

	args := make([]interface{}, 0, len(sectors)*2)
	placeholders := make([]string, 0, len(sectors))

	for i, sector := range sectors {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i*2+1, i*2+2))
		args = append(args, sector.X(), sector.Y())
	}

	query += fmt.Sprintf("%s)", fmt.Sprintf("%s", placeholders[0]))
	for _, ph := range placeholders[1:] {
		query += ", " + ph
	}

	rows, err := r.db.Connection().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find positions in sectors: %w", err)
	}
	defer rows.Close()

	positions := make([]*entity.Position, 0)

	for rows.Next() {
		var posID, userID string
		var lat, lng float64
		var sectorX, sectorY int
		var createdAt time.Time

		if err := rows.Scan(&posID, &userID, &lng, &lat, &sectorX, &sectorY, &createdAt); err != nil {
			r.logger.Error("Failed to scan sectors position row", "error", err)
			continue
		}

		position, err := r.scanToPosition(posID, userID, lat, lng, createdAt)
		if err != nil {
			r.logger.Error("Failed to reconstruct sectors position", "position_id", posID, "error", err)
			continue
		}

		positions = append(positions, position)
	}

	return positions, nil
}

// UpdateCurrentPosition atualiza posição atual do usuário
func (r *positionRepository) UpdateCurrentPosition(ctx context.Context, position *entity.Position) error {
	tx, err := r.db.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := r.updateCurrentPosition(ctx, tx, position); err != nil {
		return err
	}

	return tx.Commit()
}

// DeleteOldPositions remove posições antigas
func (r *positionRepository) DeleteOldPositions(ctx context.Context, olderThan *valueobject.Timestamp) (int, error) {
	query := `DELETE FROM positions WHERE created_at < $1`

	result, err := r.db.Connection().ExecContext(ctx, query, olderThan.Time())
	if err != nil {
		return 0, fmt.Errorf("failed to delete old positions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	r.logger.Info("Old positions deleted",
		"count", rowsAffected,
		"older_than", olderThan.String(),
	)

	return int(rowsAffected), nil
}

// scanToPosition converte dados do banco para entidade Position
func (r *positionRepository) scanToPosition(posID, userID string, lat, lng float64, recordedAt time.Time) (*entity.Position, error) {
	// Reconstruir UserID
	uid, err := entity.NewUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	// Criar posição
	position, err := entity.NewPosition(posID, *uid, lat, lng, recordedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create position: %w", err)
	}

	return position, nil
}
