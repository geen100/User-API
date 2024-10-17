package main

import (
	"errors"
	"fmt"
	"regexp"
	"unicode/utf8"
)

type User struct {
	ID        uint64 `json:"id"`
	AccountID string `json:"account_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Age       uint8  `json:"age"`
}

var validID = regexp.MustCompile(`^[a-zA-Z0-9_]{4,64}$`)

func AccountIDCharacterLimit(accountID string) error {
	if !validID.MatchString(accountID) {
		return errors.New("account ID must between 4 and 64 characters and contain only letters, numbers, and underscores")
	}
	return nil
}

func ValidateNameLength(name string) error {
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	return nil
}

func ValidateUserName(firstName, lastName string) error {
	if err := ValidateNameLength(firstName); err != nil {
		return fmt.Errorf("invalid first name: %w", err)
	}

	if err := ValidateNameLength(lastName); err != nil {
		return fmt.Errorf("invalid first name: %w", err)
	}
	return nil
}
