package api

import "github.com/google/uuid"

// Структуры для аутентификации

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=255"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	ExpiresAt int64     `json:"expires_at"`
}

// Структуры для работы с данными

type CreateEntryRequest struct {
	Name     string            `json:"name" validate:"required,min=1,max=255"`
	Type     string            `json:"type" validate:"required,oneof=credentials text binary card"`
	Data     map[string]string `json:"data" validate:"required"`
	Metadata string            `json:"metadata,omitempty"`
}

type UpdateEntryRequest struct {
	Name     string            `json:"name,omitempty"`
	Data     map[string]string `json:"data,omitempty"`
	Metadata string            `json:"metadata,omitempty"`
}

type EntryResponse struct {
	ID        uuid.UUID         `json:"id"`
	Name      string            `json:"name"`
	Type      string            `json:"type"`
	Data      map[string]string `json:"data"`
	Metadata  string            `json:"metadata,omitempty"`
	CreatedAt int64             `json:"created_at"`
	UpdatedAt int64             `json:"updated_at"`
	Version   int64             `json:"version"`
}

type ListEntriesResponse struct {
	Entries []EntryResponse `json:"entries"`
	Total   int             `json:"total"`
}

type SyncRequest struct {
	LastVersion int64 `json:"last_version"`
}

type SyncResponse struct {
	Entries     []EntryResponse `json:"entries"`
	LastVersion int64           `json:"last_version"`
	HasMore     bool            `json:"has_more"`
}

// Общие структуры

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type VersionResponse struct {
	Version   string `json:"version"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}
