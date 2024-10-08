package main

import (
	"database/sql"
	"errors"
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

func ValidateAccountID(accountID string, currentID uint64, db *sql.DB) error {
	if !validID.MatchString(accountID) {
		return errors.New("account ID must be between 4 and 64 characters and contain only letters, numbers, and underscores")
	}

	var existingID uint64
	query := "SELECT id FROM users WHERE account_id = ? AND id != ?"
	err := db.QueryRow(query, accountID, currentID).Scan(&existingID)
	if err != sql.ErrNoRows {
		return errors.New("account ID already exists")
	}
	return nil
}

func ValidateNameLength(name string) error {
	if utf8.RuneCountInString(name) > 100 {
		return errors.New("name must be 100 characters or less")
	}
	return nil
}

func ValidateUser(user User, db *sql.DB) error {
	if err := ValidateAccountID(user.AccountID, user.ID, db); err != nil {
		return err
	}
	if err := ValidateNameLength(user.FirstName); err != nil {
		return errors.New("invalid first name: " + err.Error())
	}
	if err := ValidateNameLength(user.LastName); err != nil {
		return errors.New("invalid last name: " + err.Error())
	}
	return nil
}
