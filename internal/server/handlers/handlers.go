package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Himany/GophKeeper/internal/auth"
	"github.com/Himany/GophKeeper/internal/crypto"
	"github.com/Himany/GophKeeper/internal/models"
	"github.com/Himany/GophKeeper/internal/storage"
	"github.com/Himany/GophKeeper/pkg/api"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Handler struct {
	storage    storage.Storage
	jwtManager *auth.JWTManager
	encryptor  *crypto.Encryptor
	logger     *zap.Logger
}

func NewHandler(storage storage.Storage, jwtManager *auth.JWTManager, encryptor *crypto.Encryptor, logger *zap.Logger) *Handler {
	return &Handler{
		storage:    storage,
		jwtManager: jwtManager,
		encryptor:  encryptor,
		logger:     logger,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode register request", zap.Error(err))
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingUser, _ := h.storage.GetUserByUsername(r.Context(), req.Username)
	if existingUser != nil {
		h.sendError(w, "User already exists", http.StatusConflict)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		h.logger.Error("Failed to hash password", zap.Error(err))
		h.sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.storage.CreateUser(r.Context(), user); err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		h.sendError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := h.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		h.sendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := api.AuthResponse{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	h.sendJSON(w, response, http.StatusCreated)
	h.logger.Info("User registered successfully", zap.String("username", req.Username))
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.storage.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		if err == storage.ErrUserNotFound {
			h.sendError(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		h.logger.Error("Failed to get user", zap.Error(err))
		h.sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		h.sendError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.jwtManager.GenerateToken(user.ID, user.Username)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.Error(err))
		h.sendError(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := api.AuthResponse{
		Token:     token,
		UserID:    user.ID,
		Username:  user.Username,
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	h.sendJSON(w, response, http.StatusOK)
	h.logger.Info("User logged in successfully", zap.String("username", req.Username))
}

func (h *Handler) CreateEntry(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	var req api.CreateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dataJSON, err := json.Marshal(req.Data)
	if err != nil {
		h.sendError(w, "Failed to serialize data", http.StatusBadRequest)
		return
	}

	encryptedData, err := h.encryptor.Encrypt(string(dataJSON))
	if err != nil {
		h.logger.Error("Failed to encrypt data", zap.Error(err))
		h.sendError(w, "Failed to encrypt data", http.StatusInternalServerError)
		return
	}

	entry := &models.DataEntry{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      req.Name,
		Type:      models.DataType(req.Type),
		Data:      encryptedData,
		Metadata:  req.Metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.storage.CreateEntry(r.Context(), entry); err != nil {
		h.logger.Error("Failed to create entry", zap.Error(err))
		h.sendError(w, "Failed to create entry", http.StatusInternalServerError)
		return
	}

	response := h.entryToResponse(entry, req.Data)
	h.sendJSON(w, response, http.StatusCreated)
	h.logger.Info("Entry created successfully", zap.String("name", req.Name), zap.String("type", req.Type))
}

func (h *Handler) GetEntry(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	entryIDStr := chi.URLParam(r, "id")

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		h.sendError(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	entry, err := h.storage.GetEntry(r.Context(), entryID, userID)
	if err != nil {
		if err == storage.ErrEntryNotFound {
			h.sendError(w, "Entry not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get entry", zap.Error(err))
		h.sendError(w, "Failed to get entry", http.StatusInternalServerError)
		return
	}

	data, err := h.decryptEntryData(entry.Data)
	if err != nil {
		h.logger.Error("Failed to decrypt entry data", zap.Error(err))
		h.sendError(w, "Failed to decrypt data", http.StatusInternalServerError)
		return
	}

	response := h.entryToResponse(entry, data)
	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) ListEntries(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	entries, err := h.storage.GetEntriesByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get entries", zap.Error(err))
		h.sendError(w, "Failed to get entries", http.StatusInternalServerError)
		return
	}

	var responses []api.EntryResponse
	for _, entry := range entries {
		data, err := h.decryptEntryData(entry.Data)
		if err != nil {
			h.logger.Error("Failed to decrypt entry data", zap.Error(err), zap.String("entryID", entry.ID.String()))
			continue
		}
		responses = append(responses, h.entryToResponse(&entry, data))
	}

	response := api.ListEntriesResponse{
		Entries: responses,
		Total:   len(responses),
	}

	h.sendJSON(w, response, http.StatusOK)
}

func (h *Handler) UpdateEntry(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	entryIDStr := chi.URLParam(r, "id")

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		h.sendError(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	var req api.UpdateEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entry, err := h.storage.GetEntry(r.Context(), entryID, userID)
	if err != nil {
		if err == storage.ErrEntryNotFound {
			h.sendError(w, "Entry not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to get entry", zap.Error(err))
		h.sendError(w, "Failed to get entry", http.StatusInternalServerError)
		return
	}

	if req.Name != "" {
		entry.Name = req.Name
	}
	if req.Data != nil {
		dataJSON, err := json.Marshal(req.Data)
		if err != nil {
			h.sendError(w, "Failed to serialize data", http.StatusBadRequest)
			return
		}
		encryptedData, err := h.encryptor.Encrypt(string(dataJSON))
		if err != nil {
			h.logger.Error("Failed to encrypt data", zap.Error(err))
			h.sendError(w, "Failed to encrypt data", http.StatusInternalServerError)
			return
		}
		entry.Data = encryptedData
	}
	if req.Metadata != "" {
		entry.Metadata = req.Metadata
	}

	if err := h.storage.UpdateEntry(r.Context(), entry); err != nil {
		h.logger.Error("Failed to update entry", zap.Error(err))
		h.sendError(w, "Failed to update entry", http.StatusInternalServerError)
		return
	}

	var data map[string]string
	if req.Data != nil {
		data = req.Data
	} else {
		data, err = h.decryptEntryData(entry.Data)
		if err != nil {
			h.logger.Error("Failed to decrypt entry data", zap.Error(err))
			h.sendError(w, "Failed to decrypt data", http.StatusInternalServerError)
			return
		}
	}

	response := h.entryToResponse(entry, data)
	h.sendJSON(w, response, http.StatusOK)
	h.logger.Info("Entry updated successfully", zap.String("entryID", entryID.String()))
}

func (h *Handler) DeleteEntry(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	entryIDStr := chi.URLParam(r, "id")

	entryID, err := uuid.Parse(entryIDStr)
	if err != nil {
		h.sendError(w, "Invalid entry ID", http.StatusBadRequest)
		return
	}

	if err := h.storage.DeleteEntry(r.Context(), entryID, userID); err != nil {
		if err == storage.ErrEntryNotFound {
			h.sendError(w, "Entry not found", http.StatusNotFound)
			return
		}
		h.logger.Error("Failed to delete entry", zap.Error(err))
		h.sendError(w, "Failed to delete entry", http.StatusInternalServerError)
		return
	}

	response := api.SuccessResponse{
		Success: true,
		Message: "Entry deleted successfully",
	}

	h.sendJSON(w, response, http.StatusOK)
	h.logger.Info("Entry deleted successfully", zap.String("entryID", entryID.String()))
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	var req api.SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entries, err := h.storage.GetEntriesAfterVersion(r.Context(), userID, req.LastVersion)
	if err != nil {
		h.logger.Error("Failed to get entries for sync", zap.Error(err))
		h.sendError(w, "Failed to sync data", http.StatusInternalServerError)
		return
	}

	var responses []api.EntryResponse
	for _, entry := range entries {
		data, err := h.decryptEntryData(entry.Data)
		if err != nil {
			h.logger.Error("Failed to decrypt entry data", zap.Error(err), zap.String("entryID", entry.ID.String()))
			continue
		}
		responses = append(responses, h.entryToResponse(&entry, data))
	}

	maxVersion, err := h.storage.GetMaxVersion(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get max version", zap.Error(err))
		maxVersion = req.LastVersion
	}

	response := api.SyncResponse{
		Entries:     responses,
		LastVersion: maxVersion,
		HasMore:     false,
	}

	h.sendJSON(w, response, http.StatusOK)
}

// Вспомогательные методы

func (h *Handler) getUserID(r *http.Request) uuid.UUID {
	userID, _ := r.Context().Value("userID").(uuid.UUID)
	return userID
}

func (h *Handler) decryptEntryData(encryptedData string) (map[string]string, error) {
	decryptedData, err := h.encryptor.Decrypt(encryptedData)
	if err != nil {
		return nil, err
	}

	var data map[string]string
	if err := json.Unmarshal([]byte(decryptedData), &data); err != nil {
		return nil, err
	}

	return data, nil
}

func (h *Handler) entryToResponse(entry *models.DataEntry, data map[string]string) api.EntryResponse {
	return api.EntryResponse{
		ID:        entry.ID,
		Name:      entry.Name,
		Type:      string(entry.Type),
		Data:      data,
		Metadata:  entry.Metadata,
		CreatedAt: entry.CreatedAt.Unix(),
		UpdatedAt: entry.UpdatedAt.Unix(),
		Version:   entry.Version,
	}
}

func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (h *Handler) sendError(w http.ResponseWriter, message string, status int) {
	response := api.ErrorResponse{
		Error:   http.StatusText(status),
		Message: message,
		Code:    status,
	}

	h.sendJSON(w, response, status)
}
