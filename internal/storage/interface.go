package storage

import (
	"context"

	"github.com/Himany/GophKeeper/internal/models"
	"github.com/google/uuid"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, userID uuid.UUID) error
}

type DataStorage interface {
	CreateEntry(ctx context.Context, entry *models.DataEntry) error
	GetEntry(ctx context.Context, entryID uuid.UUID, userID uuid.UUID) (*models.DataEntry, error)
	GetEntriesByUser(ctx context.Context, userID uuid.UUID) ([]models.DataEntry, error)
	GetEntriesByUserAndType(ctx context.Context, userID uuid.UUID, dataType models.DataType) ([]models.DataEntry, error)
	GetEntriesAfterVersion(ctx context.Context, userID uuid.UUID, version int64) ([]models.DataEntry, error)
	UpdateEntry(ctx context.Context, entry *models.DataEntry) error
	DeleteEntry(ctx context.Context, entryID uuid.UUID, userID uuid.UUID) error
	GetMaxVersion(ctx context.Context, userID uuid.UUID) (int64, error)
}

type Storage interface {
	UserStorage
	DataStorage

	Close() error
	Ping(ctx context.Context) error
}
