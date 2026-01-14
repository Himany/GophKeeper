package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Himany/GophKeeper/internal/models"
	"github.com/Himany/GophKeeper/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, databaseURL string) (*PostgresStorage, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	storage := &PostgresStorage{
		pool: pool,
	}

	if err := storage.createTables(ctx); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

func (s *PostgresStorage) Close() error {
	s.pool.Close()
	return nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *PostgresStorage) createTables(ctx context.Context) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);`,
		`CREATE TABLE IF NOT EXISTS data_entries (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			type VARCHAR(50) NOT NULL,
			data TEXT NOT NULL,
			metadata TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			version BIGSERIAL,
			UNIQUE(user_id, name)
		);`,
		`CREATE INDEX IF NOT EXISTS idx_data_entries_user_id ON data_entries(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_data_entries_type ON data_entries(type);`,
		`CREATE INDEX IF NOT EXISTS idx_data_entries_version ON data_entries(version);`,
	}

	for _, query := range queries {
		if _, err := s.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

func (s *PostgresStorage) CreateUser(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, username, password_hash, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := s.pool.Exec(ctx, query,
		user.ID, user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `SELECT id, username, password_hash, created_at, updated_at 
			  FROM users WHERE username = $1`

	var user models.User
	err := s.pool.QueryRow(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (s *PostgresStorage) GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	query := `SELECT id, username, password_hash, created_at, updated_at 
			  FROM users WHERE id = $1`

	var user models.User
	err := s.pool.QueryRow(ctx, query, userID).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (s *PostgresStorage) UpdateUser(ctx context.Context, user *models.User) error {
	query := `UPDATE users SET username = $2, password_hash = $3, updated_at = $4 
			  WHERE id = $1`

	user.UpdatedAt = time.Now()
	_, err := s.pool.Exec(ctx, query,
		user.ID, user.Username, user.PasswordHash, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

func (s *PostgresStorage) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (s *PostgresStorage) CreateEntry(ctx context.Context, entry *models.DataEntry) error {
	query := `INSERT INTO data_entries (id, user_id, name, type, data, metadata, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
			  RETURNING version`

	err := s.pool.QueryRow(ctx, query,
		entry.ID, entry.UserID, entry.Name, entry.Type, entry.Data, entry.Metadata,
		entry.CreatedAt, entry.UpdatedAt).Scan(&entry.Version)
	if err != nil {
		return fmt.Errorf("failed to create entry: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetEntry(ctx context.Context, entryID uuid.UUID, userID uuid.UUID) (*models.DataEntry, error) {
	query := `SELECT id, user_id, name, type, data, metadata, created_at, updated_at, version
			  FROM data_entries WHERE id = $1 AND user_id = $2`

	var entry models.DataEntry
	err := s.pool.QueryRow(ctx, query, entryID, userID).Scan(
		&entry.ID, &entry.UserID, &entry.Name, &entry.Type, &entry.Data,
		&entry.Metadata, &entry.CreatedAt, &entry.UpdatedAt, &entry.Version)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrEntryNotFound
		}
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	return &entry, nil
}

func (s *PostgresStorage) GetEntriesByUser(ctx context.Context, userID uuid.UUID) ([]models.DataEntry, error) {
	query := `SELECT id, user_id, name, type, data, metadata, created_at, updated_at, version
			  FROM data_entries WHERE user_id = $1 ORDER BY created_at`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries by user: %w", err)
	}
	defer rows.Close()

	var entries []models.DataEntry
	for rows.Next() {
		var entry models.DataEntry
		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.Name, &entry.Type, &entry.Data,
			&entry.Metadata, &entry.CreatedAt, &entry.UpdatedAt, &entry.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *PostgresStorage) GetEntriesByUserAndType(ctx context.Context, userID uuid.UUID, dataType models.DataType) ([]models.DataEntry, error) {
	query := `SELECT id, user_id, name, type, data, metadata, created_at, updated_at, version
			  FROM data_entries WHERE user_id = $1 AND type = $2 ORDER BY created_at`

	rows, err := s.pool.Query(ctx, query, userID, dataType)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries by user and type: %w", err)
	}
	defer rows.Close()

	var entries []models.DataEntry
	for rows.Next() {
		var entry models.DataEntry
		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.Name, &entry.Type, &entry.Data,
			&entry.Metadata, &entry.CreatedAt, &entry.UpdatedAt, &entry.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *PostgresStorage) GetEntriesAfterVersion(ctx context.Context, userID uuid.UUID, version int64) ([]models.DataEntry, error) {
	query := `SELECT id, user_id, name, type, data, metadata, created_at, updated_at, version
			  FROM data_entries WHERE user_id = $1 AND version > $2 ORDER BY version`

	rows, err := s.pool.Query(ctx, query, userID, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get entries after version: %w", err)
	}
	defer rows.Close()

	var entries []models.DataEntry
	for rows.Next() {
		var entry models.DataEntry
		err := rows.Scan(
			&entry.ID, &entry.UserID, &entry.Name, &entry.Type, &entry.Data,
			&entry.Metadata, &entry.CreatedAt, &entry.UpdatedAt, &entry.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (s *PostgresStorage) UpdateEntry(ctx context.Context, entry *models.DataEntry) error {
	query := `UPDATE data_entries 
			  SET name = $2, type = $3, data = $4, metadata = $5, updated_at = $6
			  WHERE id = $1 AND user_id = $7
			  RETURNING version`

	entry.UpdatedAt = time.Now()
	err := s.pool.QueryRow(ctx, query,
		entry.ID, entry.Name, entry.Type, entry.Data, entry.Metadata, entry.UpdatedAt, entry.UserID).Scan(&entry.Version)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	return nil
}

func (s *PostgresStorage) DeleteEntry(ctx context.Context, entryID uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM data_entries WHERE id = $1 AND user_id = $2`

	result, err := s.pool.Exec(ctx, query, entryID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	if result.RowsAffected() == 0 {
		return storage.ErrEntryNotFound
	}

	return nil
}

func (s *PostgresStorage) GetMaxVersion(ctx context.Context, userID uuid.UUID) (int64, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM data_entries WHERE user_id = $1`

	var maxVersion sql.NullInt64
	err := s.pool.QueryRow(ctx, query, userID).Scan(&maxVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to get max version: %w", err)
	}

	return maxVersion.Int64, nil
}
