package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/vitao/geolocation-tracker/pkg/config"
	"github.com/vitao/geolocation-tracker/pkg/logger"
)

// DB representa a conexão com o banco de dados
type DB struct {
	conn   *sql.DB
	logger logger.Logger
}

// New cria uma nova conexão com PostgreSQL
func New(cfg *config.Config, logger logger.Logger) (*DB, error) {
	// Construir string de conexão
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
	)

	// Conectar ao banco
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configurar pool de conexões
	conn.SetMaxOpenConns(25)                 // Máximo de conexões ativas
	conn.SetMaxIdleConns(5)                  // Conexões idle no pool
	conn.SetConnMaxLifetime(5 * time.Minute) // Tempo de vida da conexão

	// Testar conexão
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.DBName,
	)

	return &DB{
		conn:   conn,
		logger: logger,
	}, nil
}

// Connection retorna a conexão SQL
func (db *DB) Connection() *sql.DB {
	return db.conn
}

// Close fecha a conexão com o banco
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// Health verifica saúde da conexão
func (db *DB) Health(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// BeginTx inicia uma transação
func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.conn.BeginTx(ctx, nil)
}

// Stats retorna estatísticas da conexão
func (db *DB) Stats() sql.DBStats {
	return db.conn.Stats()
}

// LogStats loga estatísticas do pool de conexões
func (db *DB) LogStats() {
	stats := db.Stats()
	db.logger.Info("Database connection stats",
		"open_connections", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
		"wait_count", stats.WaitCount,
		"wait_duration", stats.WaitDuration,
		"max_idle_closed", stats.MaxIdleClosed,
		"max_lifetime_closed", stats.MaxLifetimeClosed,
	)
}

// Migration representa uma migração do banco
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// RunMigrations executa migrações do banco
func (db *DB) RunMigrations(ctx context.Context, migrations []Migration) error {
	// Criar tabela de migrações se não existir
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			description TEXT NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`

	if _, err := db.conn.ExecContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Verificar quais migrações já foram aplicadas
	appliedMigrations := make(map[int]bool)
	rows, err := db.conn.QueryContext(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("failed to scan migration version: %w", err)
		}
		appliedMigrations[version] = true
	}

	// Aplicar migrações pendentes
	for _, migration := range migrations {
		if appliedMigrations[migration.Version] {
			db.logger.Debug("Migration already applied",
				"version", migration.Version,
				"description", migration.Description,
			)
			continue
		}

		db.logger.Info("Applying migration",
			"version", migration.Version,
			"description", migration.Description,
		)

		// Executar migração em transação
		tx, err := db.BeginTx(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin migration transaction: %w", err)
		}

		// Executar SQL da migração
		if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
		}

		// Registrar migração como aplicada
		insertMigration := `
			INSERT INTO schema_migrations (version, description) 
			VALUES ($1, $2)
		`
		if _, err := tx.ExecContext(ctx, insertMigration, migration.Version, migration.Description); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Commit da transação
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		db.logger.Info("Migration applied successfully",
			"version", migration.Version,
		)
	}

	return nil
}
