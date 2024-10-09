package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func ConnectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"))
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := ValidateUser(user, db); err != nil {
		http.Error(w, "input error", http.StatusBadRequest)
	}

	query := "INSERT INTO `users` (`account_id`,`first_name`, `last_name`, `age`) VALUES (?, ?, ?, ?)"
	result, err := db.Exec(query, user.AccountID, user.FirstName, user.LastName, user.Age)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Failed to retrieve user ID", http.StatusInternalServerError)
		return
	}
	user.ID = uint64(id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT `id`, `account_id`, `first_name`, `last_name`, `age` FROM `users`")
	if err != nil {
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
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

func getUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user User
	query := "SELECT `id`, `account_id`, `first_name`, `last_name`, `age` FROM `users` WHERE id = ?"
	err = db.QueryRow(query, id).Scan(&user.ID, &user.AccountID, &user.FirstName, &user.LastName, &user.Age)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updates User
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var existingUser User
	query := "SELECT `id`, `account_id`, `first_name`, `last_name`, `age` FROM `users` WHERE id = ?"
	err = db.QueryRow(query, id).Scan(&existingUser.ID, &existingUser.AccountID, &existingUser.FirstName, &existingUser.LastName, &existingUser.Age)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to retrieve user", http.StatusInternalServerError)
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

	if err := ValidateNameLength(existingUser.FirstName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if err := ValidateNameLength(existingUser.LastName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	updateQuery := "UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id = ?"
	_, err = db.Exec(updateQuery, existingUser.FirstName, existingUser.LastName, existingUser.Age, id)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(existingUser)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	query := "DELETE FROM users WHERE id = ?"
	_, err = db.Exec(query, id)
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	var err error

	db, err := ConnectDB()
	if err != nil {
		log.Fatalln("Failed to connect to DB", err)
	}

	defer db.Close()

	http.HandleFunc("POST /users", CreateUser)
	http.HandleFunc("GET /", GetAllUsers)
	http.HandleFunc("GET /users/{user_id}", getUserByID)
	http.HandleFunc("PUT /users/{user_id}", UpdateUser)
	http.HandleFunc("DELETE /users/{user_id}", DeleteUser)

	log.Println("Sever stating on :8080")
	http.ListenAndServe(":8080", nil)
}
