package passwordx

import (
	"errors"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordEmpty    = errors.New("密码不能为空")
	ErrPasswordTooShort = errors.New("密码长度过短")
	ErrPasswordTooWeak  = errors.New("密码强度过低")
	ErrPasswordTooLong  = errors.New("密码长度过长")
	ErrPasswordMismatch = errors.New("密码错误")
)

const (
	minLen      = 8
	maxLen      = 64
	defaultCost = 12
)

func ValidatePasswordStrength(password string) error {
	if password == "" {
		return ErrPasswordEmpty
	}
	if len(password) < minLen {
		return ErrPasswordTooShort
	}
	if len(password) > maxLen {
		return ErrPasswordTooLong
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}
	count := 0
	if hasUpper {
		count++
	}
	if hasLower {
		count++
	}
	if hasDigit {
		count++
	}
	if hasSpecial {
		count++
	}
	if count < 3 {
		return ErrPasswordTooWeak
	}
	return nil
}

func HashAnyPassword(password string) (string, error) {
	if password == "" {
		return "", ErrPasswordEmpty
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), defaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), defaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ComparePassword(hashed string, plain string) error {
	if hashed == "" || plain == "" {
		return ErrPasswordEmpty
	}
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return err
	}
	return nil
}

func IsBcryptHash(s string) bool {
	return strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$")
}
