package ssh

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// UserAPIHandler handles HTTP requests for SSH user management.
type UserAPIHandler struct{}

func (h *UserAPIHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users", h.ListUsers).Methods("GET")
	router.HandleFunc("/users", h.AddUser).Methods("POST")
	router.HandleFunc("/users/{username}", h.DeleteUser).Methods("DELETE")
	router.HandleFunc("/users/{username}", h.UpdateUser).Methods("PUT")
}

// ListUsers godoc
// @Summary List SSH users
// @Description get all SSH users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} User
// @Router /users [get]
func (h *UserAPIHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	var users []User
	DB.Find(&users)
	json.NewEncoder(w).Encode(users)
}

// AddUser godoc
// @Summary Add a new SSH user
// @Description add a new SSH user to the system, generating a key pair
// @Tags users
// @Accept json
// @Produce json
// @Param user body User true "User object"
// @Success 200 {object} User
// @Failure 400 {string} string "Bad Request"
// @Router /users [post]
func (h *UserAPIHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Implement proper password hashing (e.g., bcrypt)
	hashedPassword := req.Password // Placeholder for now

	privateKeyBytes, err := GenerateKeyPair(req.Username)
	if err != nil {
		logrus.Errorf("Error generating key pair for %s: %v", req.Username, err)
		http.Error(w, fmt.Sprintf("Error generating key pair: %v", err), http.StatusBadRequest)
		return
	}

	// GenerateKeyPair already saves the public key, but we need to update the password
	var user User
	if err := DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		logrus.Errorf("Error retrieving user %s after key generation: %v", req.Username, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	user.Password = hashedPassword
	DB.Save(&user)

	// Return the user details and the private key
	response := struct {
		User       User   `json:"user"`
		PrivateKey string `json:"private_key"`
	}{
		User:       user,
		PrivateKey: string(privateKeyBytes),
	}

	json.NewEncoder(w).Encode(response)
}

// DeleteUser godoc
// @Summary Delete an SSH user
// @Description delete an SSH user from the system
// @Tags users
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Success 200 {string} string "User deleted"
// @Failure 404 {string} string "User not found"
// @Router /users/{username} [delete]
func (h *UserAPIHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var user User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	DB.Delete(&user)
	fmt.Fprintf(w, "User %s deleted", username)
}

// UpdateUser godoc
// @Summary Update an SSH user
// @Description update an existing SSH user's details
// @Tags users
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Param user body User true "User object"
// @Success 200 {object} User
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "User not found"
// @Router /users/{username} [put]
func (h *UserAPIHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	var user User
	if err := DB.Where("username = ?", username).First(&user).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewDecoder(r.Body).Decode(&user)
	DB.Save(&user)
	json.NewEncoder(w).Encode(user)
}
