package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"onb2/api/models"
	"onb2/api/valid"
	"strconv"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := valid.AccountIDCharacterLimit(user.AccountID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := valid.ValidateUserName(user.FirstName, user.LastName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Failed to start transtation", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	var existingID uint64
	query := "SELECT COUNT(*) FROM `users` WHERE `account_id` = ? "
	err = tx.QueryRowContext(ctx, query, user.AccountID).Scan(&existingID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingID > 0 {
		http.Error(w, "account_id is already exists", http.StatusConflict)
		return
	}

	query = "INSERT INTO `users` (`account_id`,`first_name`, `last_name`, `age`) VALUES (?, ?, ?, ?)"
	result, err := tx.ExecContext(ctx, query, user.AccountID, user.FirstName, user.LastName, user.Age)

	var duplictErr *mysql.MySQLError
	if errors.As(err, &duplictErr) && duplictErr.Number == 1062 {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	} else if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve user ID", http.StatusInternalServerError)
		return
	}
	user.ID = uint64(id)

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := db.QueryContext(ctx, "SELECT `id`, `account_id`, `first_name`, `last_name`, `age` FROM `users`")
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.AccountID, &user.FirstName, &user.LastName, &user.Age); err != nil {
			http.Error(w, "Failed to scan user", http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error occurred during rows iteration", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("user_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	query := "SELECT `id`, `account_id`, `first_name`, `last_name`, `age` FROM `users` WHERE `id` = ?"
	err = db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.AccountID, &user.FirstName, &user.LastName, &user.Age)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("user_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updates models.User
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var existingUser models.User

	if err := valid.ValidateUserName(existingUser.FirstName, existingUser.LastName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		http.Error(w, "Failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	query := "SELECT `id`, `account_id`, `first_name`, `last_name`, `age` FROM `users` WHERE `id` = ? FOR UPDATE"
	err = tx.QueryRowContext(ctx, query, id).Scan(&existingUser.ID, &existingUser.AccountID, &existingUser.FirstName, &existingUser.LastName, &existingUser.Age)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	if updates.FirstName != "" {
		existingUser.FirstName = updates.FirstName
	}
	if updates.LastName != "" {
		existingUser.LastName = updates.LastName
	}
	if updates.Age != 0 {
		existingUser.Age = updates.Age
	}

	updateQuery := "UPDATE `users` SET `first_name` = ?, `last_name` = ?, `age` = ? WHERE `id` = ?"
	_, err = tx.ExecContext(ctx, updateQuery, existingUser.FirstName, existingUser.LastName, existingUser.Age, id)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existingUser)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idStr := r.PathValue("user_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM `users` WHERE `id` = ?"
	_, err = db.ExecContext(ctx, query, id)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
