package validator

import (
	"testing"
)

func TestValidateUsername(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid username", "user123", false},
		{"valid with underscore", "user_name", false},
		{"valid with hyphen", "user-name", false},
		{"empty username", "", true},
		{"too short", "ab", true},
		{"too long", "a123456789012345678901234567890123456789012345678901", true},
		{"invalid chars", "user@name", true},
		{"spaces", "user name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	v := New()

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "password123", false},
		{"minimum length", "12345678", false},
		{"empty password", "", true},
		{"too short", "pass", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCardNumber(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		number  string
		wantErr bool
	}{
		{"valid visa", "4532015112830366", false},
		{"valid mastercard", "5425233430109903", false},
		{"with spaces", "4532 0151 1283 0366", false},
		{"with hyphens", "4532-0151-1283-0366", false},
		{"empty", "", true},
		{"too short", "12345", true},
		{"invalid luhn", "1234567890123456", true},
		{"non-digits", "abcd1234efgh5678", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCardNumber(tt.number)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCardNumber() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateExpiryDate(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		date    string
		wantErr bool
	}{
		{"valid format", "12/25", false},
		{"empty", "", true},
		{"wrong format", "2025-12", true},
		{"no slash", "1225", true},
		{"three digit month", "123/25", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateExpiryDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExpiryDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCVV(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		cvv     string
		wantErr bool
	}{
		{"valid 3 digits", "123", false},
		{"valid 4 digits", "1234", false},
		{"empty", "", true},
		{"too short", "12", true},
		{"too long", "12345", true},
		{"non-digits", "abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateCVV(tt.cvv)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCVV() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEntryType(t *testing.T) {
	v := New()

	tests := []struct {
		name      string
		entryType string
		wantErr   bool
	}{
		{"credentials", "credentials", false},
		{"text", "text", false},
		{"binary", "binary", false},
		{"card", "card", false},
		{"invalid", "unknown", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateEntryType(tt.entryType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEntryType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
