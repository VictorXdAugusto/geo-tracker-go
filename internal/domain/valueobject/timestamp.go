package valueobject

import (
	"errors"
	"fmt"
	"time"
)

// Timestamp representa um momento no tempo com timezone
// Value Object que encapsula comportamentos específicos de tempo para o domínio
type Timestamp struct {
	time time.Time
}

// Constantes de tempo
const (
	TimestampFormat = time.RFC3339 // "2006-01-02T15:04:05Z07:00"
)

// Erros específicos
var (
	ErrInvalidTime = errors.New("invalid time")
	ErrFutureTime  = errors.New("time cannot be in the future")
)

// Now cria um timestamp com o tempo atual
func Now() *Timestamp {
	return &Timestamp{time: time.Now().UTC()}
}

// NewTimestamp cria um timestamp a partir de time.Time
func NewTimestamp(t time.Time) *Timestamp {
	return &Timestamp{time: t.UTC()}
}

// NewTimestampFromString cria um timestamp a partir de string RFC3339
func NewTimestampFromString(timeStr string) (*Timestamp, error) {
	t, err := time.Parse(TimestampFormat, timeStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidTime, err.Error())
	}

	return &Timestamp{time: t.UTC()}, nil
}

// NewTimestampNotInFuture cria timestamp que não pode ser futuro
// Útil para validar timestamps de posições registradas
func NewTimestampNotInFuture(t time.Time) (*Timestamp, error) {
	utcTime := t.UTC()
	now := time.Now().UTC()

	if utcTime.After(now) {
		return nil, fmt.Errorf("%w: got %s, now is %s", ErrFutureTime, utcTime.Format(TimestampFormat), now.Format(TimestampFormat))
	}

	return &Timestamp{time: utcTime}, nil
}

// Time retorna o time.Time interno
func (ts *Timestamp) Time() time.Time {
	return ts.time
}

// Unix retorna timestamp Unix
func (ts *Timestamp) Unix() int64 {
	return ts.time.Unix()
}

// String implementa fmt.Stringer
func (ts *Timestamp) String() string {
	return ts.time.Format(TimestampFormat)
}

// Equals compara dois timestamps
func (ts *Timestamp) Equals(other *Timestamp) bool {
	if other == nil {
		return false
	}
	return ts.time.Equal(other.time)
}

// Before verifica se este timestamp é anterior ao outro
func (ts *Timestamp) Before(other *Timestamp) bool {
	if other == nil {
		return false
	}
	return ts.time.Before(other.time)
}

// After verifica se este timestamp é posterior ao outro
func (ts *Timestamp) After(other *Timestamp) bool {
	if other == nil {
		return true
	}
	return ts.time.After(other.time)
}

// DurationSince calcula duração desde outro timestamp
func (ts *Timestamp) DurationSince(other *Timestamp) time.Duration {
	if other == nil {
		return 0
	}
	return ts.time.Sub(other.time)
}

// IsWithinLast verifica se timestamp está dentro do período especificado
func (ts *Timestamp) IsWithinLast(duration time.Duration) bool {
	now := time.Now().UTC()
	cutoff := now.Add(-duration)
	return ts.time.After(cutoff)
}

// Age retorna a idade do timestamp
func (ts *Timestamp) Age() time.Duration {
	return time.Since(ts.time)
}

// IsExpired verifica se timestamp expirou baseado em TTL
func (ts *Timestamp) IsExpired(ttl time.Duration) bool {
	return ts.Age() > ttl
}

// AddDuration adiciona duração ao timestamp (retorna novo timestamp)
func (ts *Timestamp) AddDuration(duration time.Duration) *Timestamp {
	return &Timestamp{time: ts.time.Add(duration)}
}

// Truncate trunca timestamp para precisão especificada
func (ts *Timestamp) Truncate(precision time.Duration) *Timestamp {
	return &Timestamp{time: ts.time.Truncate(precision)}
}

// ToDate retorna apenas a data (00:00:00 UTC)
func (ts *Timestamp) ToDate() *Timestamp {
	year, month, day := ts.time.Date()
	dateOnly := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &Timestamp{time: dateOnly}
}
