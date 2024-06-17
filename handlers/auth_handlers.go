package handlers

import (
	"errors"
	"forum/models"
	"log"
	"net/http"
	"regexp"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// validatePassword checks the robustness of a password
func validatePassword(password string) error {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(password) >= 8 {
		hasMinLen = true
	}
	for _, char := range password {
		switch {
		case regexp.MustCompile(`[A-Z]`).MatchString(string(char)):
			hasUpper = true
		case regexp.MustCompile(`[a-z]`).MatchString(string(char)):
			hasLower = true
		case regexp.MustCompile(`[0-9]`).MatchString(string(char)):
			hasNumber = true
		case regexp.MustCompile(`[!@#\$%\^&\*]`).MatchString(string(char)):
			hasSpecial = true
		}
	}
	if !hasMinLen || !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("password must be at least 8 characters long and include at least one uppercase letter, one lowercase letter, one number, and one special character")
	}
	return nil
}

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Register handler")

	if r.Method == http.MethodPost {
		log.Println("Handling POST request for registration")

		// Parse the form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Create a new user instance
		user := models.User{
			Username: r.FormValue("username"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		// Log the form values
		log.Printf("Parsed form values - Username: %s, Email: %s", user.Username, user.Email)

		// Validate form values
		if user.Username == "" || user.Email == "" || user.Password == "" {
			log.Println("Missing required form values")
			data := map[string]interface{}{
				"Error": "Missing required form values",
			}
			renderTemplate(w, "register", data)
			return
		}

		// Check if the username already exists
		var existingUser models.User
		if err := db.Where("username = ?", user.Username).First(&existingUser).Error; err == nil {
			log.Println("Username already in use")
			data := map[string]interface{}{
				"UsernameError": "Username already in use",
			}
			renderTemplate(w, "register", data)
			return
		}

		// Validate password robustness
		if err := validatePassword(user.Password); err != nil {
			log.Printf("Password validation failed: %v", err)
			data := map[string]interface{}{
				"PasswordError": err.Error(),
			}
			renderTemplate(w, "register", data)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		user.Password = string(hashedPassword)

		// Use a transaction to create the user
		result := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			return nil
		})

		if result != nil {
			log.Printf("Unable to register user: %v", result)
			data := map[string]interface{}{
				"Error": "Unable to register user",
			}
			renderTemplate(w, "register", data)
			return
		}

		log.Printf("User registered successfully with ID: %d", user.ID)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		log.Println("Rendering registration template")
		renderTemplate(w, "register", nil)
	}
}

// Login handles user login
func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Login handler")

	if r.Method == http.MethodPost {
		log.Println("Handling POST request for login")

		// Parse the form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Get user credentials from the form
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Validate form values
		if email == "" || password == "" {
			log.Println("Missing email or password")
			http.Error(w, "Missing email or password", http.StatusBadRequest)
			return
		}

		// Fetch the user from the database
		var user models.User
		result := db.Where("email = ?", email).First(&user)
		if result.Error != nil {
			log.Printf("Invalid email or password: %v", result.Error)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Compare the hashed password with the plain text password
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			log.Printf("Invalid email or password: %v", err)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Create a session and store user information
		session, _ := store.Get(r, "session")
		session.Values["user"] = user.Username
		session.Values["userID"] = user.ID
		if err := session.Save(r, w); err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, "Error saving session", http.StatusInternalServerError)
			return
		}

		log.Println("User logged in successfully")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		log.Println("Rendering login template")
		renderTemplate(w, "login", nil)
	}
}

// Logout handles user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Logout handler")

	// Fetch the session
	session, _ := store.Get(r, "session")

	// Remove user information from the session
	delete(session.Values, "user")
	delete(session.Values, "userID")
	if err := session.Save(r, w); err != nil {
		log.Printf("Error saving session: %v", err)
		http.Error(w, "Error saving session", http.StatusInternalServerError)
		return
	}

	log.Println("User logged out successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
