// Package mysqlstore provides a mysql store implementation for sagas.
package mysqlstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DoNewsCode/core/dtx"
	"github.com/DoNewsCode/core/dtx/sagas"
	"gorm.io/gorm"
)

// MySQLStore is a Store implementation for sagas.
type MySQLStore struct {
	db              *gorm.DB
	retention       time.Duration
	cleanupInterval time.Duration
}

// Option is the type for MySQLStore options.
type Option func(store *MySQLStore)

// WithRetention is the option that sets the maximum log retention after which the log
// will be cleared.
func WithRetention(duration time.Duration) Option {
	return func(store *MySQLStore) {
		store.retention = duration
	}
}

// WithCleanUpInterval is the option that sets the clean up interval.
func WithCleanUpInterval(duration time.Duration) Option {
	return func(store *MySQLStore) {
		store.cleanupInterval = duration
	}
}

// New returns a pointer to MySQLStore.
func New(db *gorm.DB, opts ...Option) *MySQLStore {
	s := &MySQLStore{db: db, retention: 168 * time.Hour, cleanupInterval: time.Hour}
	for _, f := range opts {
		f(s)
	}
	return s
}

// Log appends the log to mysql store.
func (s *MySQLStore) Log(ctx context.Context, log sagas.Log) error {
	return s.db.WithContext(ctx).Exec(
		"INSERT INTO saga_logs (id, correlation_id, started_at, finished_at, log_type, step_name, step_param, step_error) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		log.ID, ctx.Value(dtx.CorrelationID), log.StartedAt, nil, log.LogType, log.StepName, log.StepParam, "",
	).Error
}

// Ack acknowledges a transaction step is completed.
func (s *MySQLStore) Ack(ctx context.Context, id string, err error) error {
	var estr string
	if err != nil {
		estr = err.Error()
	}
	return s.db.WithContext(ctx).Exec("UPDATE saga_logs SET finished_at = ?, step_error = ? WHERE id = ?", time.Now(), estr, id).Error
}

// UnacknowledgedSteps returns all unacknowledged steps from the store. Those steps are up for rollback.
func (s *MySQLStore) UnacknowledgedSteps(ctx context.Context, correlationID string) ([]sagas.Log, error) {
	return s.unacknowledgedSteps(ctx, correlationID)
}

// UncommittedSagas searches all uncommitted sagas and returns unacknowledged steps from those sagas.
func (s *MySQLStore) UncommittedSagas(ctx context.Context) ([]sagas.Log, error) {
	var (
		logs           []sagas.Log
		correlationIDs []string
	)
	rows, err := s.db.WithContext(ctx).Raw("SELECT correlation_id FROM saga_logs WHERE log_type = ? and finished_at IS NULL", sagas.Session).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var correlationID string
		rows.Scan(&correlationID)
		correlationIDs = append(correlationIDs, correlationID)
	}
	for _, id := range correlationIDs {
		log, err := s.unacknowledgedSteps(ctx, id)
		if err != nil {
			return logs, err
		}
		logs = append(logs, log...)
	}
	return logs, nil
}

// CleanUp periodically removes the logs that exceed their of maximum retention.
func (s *MySQLStore) CleanUp(ctx context.Context) error {
	timer := time.NewTicker(s.cleanupInterval)
	for {
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			err := s.cleanUp(ctx)
			if err != nil {
				timer.Stop()
				return err
			}
		}
	}
}

func (s *MySQLStore) unacknowledgedSteps(ctx context.Context, correlationID string) ([]sagas.Log, error) {
	var logs = make(map[string]sagas.Log)
	rows, err := s.db.WithContext(ctx).Raw(
		`SELECT id, correlation_id, started_at, finished_at, log_type, step_name, step_param, step_error
		FROM saga_logs WHERE correlation_id = ? and log_type = ?`,
		correlationID,
		sagas.Do,
	).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var log sagas.Log
		var nullFinishedAt sql.NullTime
		var errString string
		err := rows.Scan(&log.ID, &log.CorrelationID, &log.StartedAt, &nullFinishedAt, &log.LogType, &log.StepName, &log.StepParam, &errString)
		if err != nil {
			return nil, err
		}
		if nullFinishedAt.Valid {
			log.FinishedAt = nullFinishedAt.Time
		}
		if errString != "" {
			log.StepError = errors.New(errString)
		}
		logs[log.StepName] = log
	}
	excludeRows, err := s.db.WithContext(ctx).Raw(
		"SELECT step_name FROM saga_logs WHERE correlation_id = ? and log_type = ? and finished_at IS NOT NULL and step_error = ?",
		correlationID,
		sagas.Undo,
		"",
	).Rows()
	if err != nil {
		return nil, err
	}
	defer excludeRows.Close()
	for excludeRows.Next() {
		var stepName string
		_ = excludeRows.Scan(&stepName)
		delete(logs, stepName)
	}
	var result []sagas.Log
	for i := range logs {
		result = append(result, logs[i])
	}
	return result, nil
}

// cleanUp removes the logs that exceed their of maximum retention. It can be called periodically to save disk space.
func (s *MySQLStore) cleanUp(ctx context.Context) error {
	return s.db.WithContext(ctx).Exec(
		"DELETE FROM saga_logs WHERE started_at < ?",
		time.Now().Add(-s.retention),
	).Error
}
