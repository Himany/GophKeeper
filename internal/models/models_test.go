package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDataType_Constants(t *testing.T) {
	assert.Equal(t, DataType("credentials"), DataTypeCredentials)
	assert.Equal(t, DataType("text"), DataTypeText)
	assert.Equal(t, DataType("binary"), DataTypeBinary)
	assert.Equal(t, DataType("card"), DataTypeCard)
}

func TestUser_Creation(t *testing.T) {
	user := User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.False(t, user.CreatedAt.IsZero())
	assert.False(t, user.UpdatedAt.IsZero())
}

func TestDataEntry_Creation(t *testing.T) {
	userID := uuid.New()
	entryID := uuid.New()
	now := time.Now()

	entry := DataEntry{
		ID:        entryID,
		UserID:    userID,
		Name:      "Test Entry",
		Type:      DataTypeCredentials,
		Data:      "encrypted_data",
		Metadata:  "test metadata",
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}

	assert.Equal(t, entryID, entry.ID)
	assert.Equal(t, userID, entry.UserID)
	assert.Equal(t, "Test Entry", entry.Name)
	assert.Equal(t, DataTypeCredentials, entry.Type)
	assert.Equal(t, "encrypted_data", entry.Data)
	assert.Equal(t, "test metadata", entry.Metadata)
	assert.Equal(t, now, entry.CreatedAt)
	assert.Equal(t, now, entry.UpdatedAt)
	assert.Equal(t, int64(1), entry.Version)
}

func TestCredentials_Structure(t *testing.T) {
	creds := Credentials{
		Login:    "testuser",
		Password: "testpass",
		URL:      "https://example.com",
	}

	assert.Equal(t, "testuser", creds.Login)
	assert.Equal(t, "testpass", creds.Password)
	assert.Equal(t, "https://example.com", creds.URL)
}

func TestTextData_Structure(t *testing.T) {
	text := TextData{
		Content: "Some important text content",
	}

	assert.Equal(t, "Some important text content", text.Content)
}

func TestBinaryData_Structure(t *testing.T) {
	data := []byte("binary file content")
	binary := BinaryData{
		Filename: "test.txt",
		Content:  data,
		MimeType: "text/plain",
	}

	assert.Equal(t, "test.txt", binary.Filename)
	assert.Equal(t, data, binary.Content)
	assert.Equal(t, "text/plain", binary.MimeType)
}

func TestCardData_Structure(t *testing.T) {
	card := CardData{
		Number:     "1234567890123456",
		Holder:     "John Doe",
		ExpiryDate: "12/25",
		CVV:        "123",
		Bank:       "Test Bank",
	}

	assert.Equal(t, "1234567890123456", card.Number)
	assert.Equal(t, "John Doe", card.Holder)
	assert.Equal(t, "12/25", card.ExpiryDate)
	assert.Equal(t, "123", card.CVV)
	assert.Equal(t, "Test Bank", card.Bank)
}

func TestSyncRequest_Structure(t *testing.T) {
	req := SyncRequest{
		LastSyncVersion: 42,
	}

	assert.Equal(t, int64(42), req.LastSyncVersion)
}

func TestSyncResponse_Structure(t *testing.T) {
	entries := []DataEntry{
		{
			ID:      uuid.New(),
			Name:    "Entry 1",
			Type:    DataTypeText,
			Version: 1,
		},
		{
			ID:      uuid.New(),
			Name:    "Entry 2",
			Type:    DataTypeCredentials,
			Version: 2,
		},
	}

	resp := SyncResponse{
		Entries:     entries,
		LastVersion: 2,
	}

	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, "Entry 1", resp.Entries[0].Name)
	assert.Equal(t, "Entry 2", resp.Entries[1].Name)
	assert.Equal(t, int64(2), resp.LastVersion)
}

func TestDataEntry_AllTypes(t *testing.T) {
	userID := uuid.New()
	now := time.Now()

	testCases := []struct {
		name     string
		dataType DataType
	}{
		{"Credentials Entry", DataTypeCredentials},
		{"Text Entry", DataTypeText},
		{"Binary Entry", DataTypeBinary},
		{"Card Entry", DataTypeCard},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			entry := DataEntry{
				ID:        uuid.New(),
				UserID:    userID,
				Name:      tc.name,
				Type:      tc.dataType,
				Data:      "encrypted_data",
				Metadata:  "metadata",
				CreatedAt: now,
				UpdatedAt: now,
				Version:   1,
			}

			assert.Equal(t, tc.dataType, entry.Type)
			assert.Equal(t, tc.name, entry.Name)
		})
	}
}
