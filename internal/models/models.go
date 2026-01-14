package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	PasswordHash string    `json:"-" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type DataType string

const (
	DataTypeCredentials DataType = "credentials"
	DataTypeText        DataType = "text"
	DataTypeBinary      DataType = "binary"
	DataTypeCard        DataType = "card"
)

type DataEntry struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Name      string    `json:"name" db:"name"`
	Type      DataType  `json:"type" db:"type"`
	Data      string    `json:"data" db:"data"`
	Metadata  string    `json:"metadata" db:"metadata"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Version   int64     `json:"version" db:"version"`
}

type Credentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	URL      string `json:"url,omitempty"`
}

type TextData struct {
	Content string `json:"content"`
}

type BinaryData struct {
	Filename string `json:"filename"`
	Content  []byte `json:"content"`
	MimeType string `json:"mime_type,omitempty"`
}

type CardData struct {
	Number     string `json:"number"`
	Holder     string `json:"holder"`
	ExpiryDate string `json:"expiry_date"`
	CVV        string `json:"cvv"`
	Bank       string `json:"bank,omitempty"`
}

type SyncRequest struct {
	LastSyncVersion int64 `json:"last_sync_version"`
}

type SyncResponse struct {
	Entries     []DataEntry `json:"entries"`
	LastVersion int64       `json:"last_version"`
}
