package validator

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrEmptyValue    = errors.New("value cannot be empty")
	ErrInvalidFormat = errors.New("invalid format")
	ErrInvalidLength = errors.New("invalid length")
)

type Validator struct{}

func New() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if username == "" {
		return fmt.Errorf("username: %w", ErrEmptyValue)
	}

	if len(username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}

	if len(username) > 50 {
		return fmt.Errorf("username must not exceed 50 characters")
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, username)
	if !matched {
		return fmt.Errorf("username: %w (only alphanumeric, underscore, and hyphen allowed)", ErrInvalidFormat)
	}

	return nil
}

func (v *Validator) ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("password: %w", ErrEmptyValue)
	}

	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	return nil
}

func (v *Validator) ValidateEntryName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return fmt.Errorf("entry name: %w", ErrEmptyValue)
	}

	if len(name) < 1 {
		return fmt.Errorf("entry name: %w", ErrEmptyValue)
	}

	if len(name) > 100 {
		return fmt.Errorf("entry name must not exceed 100 characters")
	}

	return nil
}

func (v *Validator) ValidateEntryType(entryType string) error {
	validTypes := map[string]bool{
		"credentials": true,
		"text":        true,
		"binary":      true,
		"card":        true,
	}

	if !validTypes[entryType] {
		return fmt.Errorf("entry type: %w (must be one of: credentials, text, binary, card)", ErrInvalidFormat)
	}

	return nil
}

func (v *Validator) ValidateCardNumber(number string) error {
	number = strings.ReplaceAll(number, " ", "")
	number = strings.ReplaceAll(number, "-", "")

	if number == "" {
		return fmt.Errorf("card number: %w", ErrEmptyValue)
	}

	matched, _ := regexp.MatchString(`^\d+$`, number)
	if !matched {
		return fmt.Errorf("card number: %w (must contain only digits)", ErrInvalidFormat)
	}

	if len(number) < 13 || len(number) > 19 {
		return fmt.Errorf("card number: %w (must be between 13 and 19 digits)", ErrInvalidLength)
	}

	if !luhnCheck(number) {
		return fmt.Errorf("card number: %w (failed Luhn check)", ErrInvalidFormat)
	}

	return nil
}

func (v *Validator) ValidateExpiryDate(date string) error {
	if date == "" {
		return fmt.Errorf("expiry date: %w", ErrEmptyValue)
	}

	matched, _ := regexp.MatchString(`^\d{2}/\d{2}$`, date)
	if !matched {
		return fmt.Errorf("expiry date: %w (must be in MM/YY format)", ErrInvalidFormat)
	}

	return nil
}

func (v *Validator) ValidateCVV(cvv string) error {
	if cvv == "" {
		return fmt.Errorf("CVV: %w", ErrEmptyValue)
	}

	matched, _ := regexp.MatchString(`^\d{3,4}$`, cvv)
	if !matched {
		return fmt.Errorf("CVV: %w (must be 3 or 4 digits)", ErrInvalidFormat)
	}

	return nil
}

func (v *Validator) ValidateMetadata(metadata string) error {
	if len(metadata) > 1000 {
		return fmt.Errorf("metadata must not exceed 1000 characters")
	}

	return nil
}

func (v *Validator) ValidateCredentialsData(data map[string]string) error {
	login, exists := data["login"]
	if !exists || strings.TrimSpace(login) == "" {
		return fmt.Errorf("login: %w", ErrEmptyValue)
	}

	password, exists := data["password"]
	if !exists || password == "" {
		return fmt.Errorf("password: %w", ErrEmptyValue)
	}

	return nil
}

func (v *Validator) ValidateTextData(data map[string]string) error {
	content, exists := data["content"]
	if !exists || strings.TrimSpace(content) == "" {
		return fmt.Errorf("text content: %w", ErrEmptyValue)
	}

	return nil
}

func (v *Validator) ValidateBinaryData(data map[string]string) error {
	filename, exists := data["filename"]
	if !exists || strings.TrimSpace(filename) == "" {
		return fmt.Errorf("filename: %w", ErrEmptyValue)
	}

	content, exists := data["content"]
	if !exists || content == "" {
		return fmt.Errorf("file content: %w", ErrEmptyValue)
	}

	return nil
}

func (v *Validator) ValidateCardData(data map[string]string) error {
	number, exists := data["number"]
	if !exists {
		return fmt.Errorf("card number: %w", ErrEmptyValue)
	}
	if err := v.ValidateCardNumber(number); err != nil {
		return err
	}

	holder, exists := data["holder"]
	if !exists || strings.TrimSpace(holder) == "" {
		return fmt.Errorf("cardholder name: %w", ErrEmptyValue)
	}

	expiryDate, exists := data["expiry_date"]
	if !exists {
		return fmt.Errorf("expiry date: %w", ErrEmptyValue)
	}
	if err := v.ValidateExpiryDate(expiryDate); err != nil {
		return err
	}

	cvv, exists := data["cvv"]
	if !exists {
		return fmt.Errorf("CVV: %w", ErrEmptyValue)
	}
	if err := v.ValidateCVV(cvv); err != nil {
		return err
	}

	return nil
}

func luhnCheck(number string) bool {
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		n := int(number[i] - '0')

		if alternate {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}

		sum += n
		alternate = !alternate
	}

	return sum%10 == 0
}
