package webserver

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// Get a user by ID
func getUser(c echo.Context) error {
	id := c.Param("id")

	// Query the database for the user
	var userDetail User
	err := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id).Scan(&userDetail.ID, &userDetail.Username, &userDetail.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "User not found")
		}
		return err
	}

	return c.JSON(http.StatusOK, userDetail)
}

// Create a new user
func createUser(c echo.Context) error {
	user := new(User)
	if err := c.Bind(user); err != nil {
		return err
	}

	// Hash the password
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}

	// Insert the user into the database
	result, err := db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", user.Username, hashedPassword, user.Role)
	if err != nil {
		return err
	}

	// Get the auto-generated ID of the new user
	id, _ := result.LastInsertId()
	user.ID = int(id)
	user.Password = "" // Clear the password field for security

	return c.JSON(http.StatusCreated, user)
}

// Update an existing user
func updateUser(c echo.Context) error {
	id := c.Param("id")
	user := new(User)
	if err := c.Bind(user); err != nil {
		return err
	}

	// Hash the password if it's provided
	if user.Password != "" {
		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	// Update the user in the database
	_, err := db.Exec("UPDATE users SET username = ?, password = ?, role = ? WHERE id = ?", user.Username, user.Password, user.Role, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{"message": "User info has updated"})
}

// Delete a user
func deleteUser(c echo.Context) error {
	id := c.Param("id")

	// Delete the user from the database
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK,
		map[string]interface{}{"message": "User is no longer present"})
}

// Authenticate an admin
func loginAdmin(username, password string, c echo.Context) (bool, error) {
	// Query the database for the user
	var user User
	err := db.QueryRow(`
        SELECT id, username, password, role
        FROM users
        WHERE role = "admin" AND username = ?`,
		username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	// Verify the password
	err = verifyPassword(user.Password, password)
	if err != nil {
		return false, nil
	}

	return true, nil
}

// Hash a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Verify a password against its hash
func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
