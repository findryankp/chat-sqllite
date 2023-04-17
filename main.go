package main

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Database connection
	var err error
	db, err = sql.Open("sqlite3", "mydb.sqlite3")
	if err != nil {
		e.Logger.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS chats (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		from_user TEXT NOT NULL,
		to_user TEXT NOT NULL,
		message TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		panic(err)
	}

	defer db.Close()

	e.GET("/chats/:from_user/:to_user", getChat)
	e.POST("/chats", postChat)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func getChat(c echo.Context) error {
	fromUser := c.Param("from_user")
	toUser := c.Param("to_user")

	rows, err := db.Query("SELECT id, from_user, to_user, message, created_at FROM chats WHERE (from_user=? AND to_user=?) OR (from_user=? AND to_user=?) ORDER BY created_at DESC", fromUser, toUser, toUser, fromUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	defer rows.Close()

	chats := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int
		var fromUser string
		var toUser string
		var message string
		var createdAt string
		if err := rows.Scan(&id, &fromUser, &toUser, &message, &createdAt); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		}

		chat := map[string]interface{}{
			"id":         id,
			"from_user":  fromUser,
			"to_user":    toUser,
			"message":    message,
			"created_at": createdAt,
		}
		chats = append(chats, chat)
	}

	return c.JSON(http.StatusOK, chats)
}

func postChat(c echo.Context) error {
	fromUser := c.FormValue("from_user")
	toUser := c.FormValue("to_user")
	message := c.FormValue("message")

	// Insert the chat message into the database
	_, err := db.Exec("INSERT INTO chats (from_user, to_user, message) VALUES (?, ?, ?)", fromUser, toUser, message)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}

	response := map[string]any{
		"message": "Chat created successfully",
	}
	return c.JSON(http.StatusCreated, response)
}
