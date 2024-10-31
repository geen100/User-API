package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"onb2/api/handlers"
	"os"
)

func ConnectDB() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("MYSQL_USER"),
		os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"),
		os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DATABASE"))
	return sql.Open("mysql", dsn)
}

func main() {
	db, err := ConnectDB()
	if err != nil {
		log.Fatalln("Failed to connect to DB", err)
	}
	defer db.Close()

	handlers.SetDB(db)

	http.HandleFunc("POST /users", handlers.CreateUser)
	http.HandleFunc("GET /", handlers.GetAllUsers)
	http.HandleFunc("GET /users/{user_id}", handlers.GetUserByID)
	http.HandleFunc("PUT /users/{user_id}", handlers.UpdateUser)
	http.HandleFunc("DELETE /users/{user_id}", handlers.DeleteUser)

	log.Println("Server stating on :8081")
	http.ListenAndServe(":8081", nil)
}
