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

		// Extract form values
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Check if username or email already exists
		var existingUser models.User
		if err := db.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err != gorm.ErrRecordNotFound {
			log.Printf("Username or Email already in use: %v", err)
			data := map[string]interface{}{
				"Error":    "Username or Email already exists",
				"Username": username,
				"Email":    email,
			}
			renderTemplate(w, "register", data)
			return
		}

		// Validate password robustness
		if err := validatePassword(password); err != nil {
			log.Printf("Password validation failed: %v", err)
			data := map[string]interface{}{
				"PasswordError": err.Error(),
			}
			renderTemplate(w, "register", data)
			return
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Create a new user instance and save it to the database
		user := models.User{
			Username: username,
			Email:    email,
			Password: string(hashedPassword),
			// Assuming security questions and answers are also provided in the form
			SecurityQuestion1: r.FormValue("securityQuestion1"),
			SecurityAnswer1:   r.FormValue("securityAnswer1"),
			SecurityQuestion2: r.FormValue("securityQuestion2"),
			SecurityAnswer2:   r.FormValue("securityAnswer2"),
			SecurityQuestion3: r.FormValue("securityQuestion3"),
			SecurityAnswer3:   r.FormValue("securityAnswer3"),
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Unable to register user: %v", err)
			data := map[string]interface{}{
				"Error": "Unable to register user",
			}
			renderTemplate(w, "register", data)
			return
		}

		log.Printf("User registered successfully with ID: %d", user.ID)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
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

func RegisterStep1(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Register Step 1 handler")

	if r.Method == http.MethodPost {
		log.Println("Handling POST request for registration step 1")

		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		if username == "" || email == "" || password == "" {
			log.Println("Missing required form values")
			data := map[string]interface{}{
				"Error": "Missing required form values",
			}
			renderTemplate(w, "register", data)
			return
		}

		var existingUser models.User
		// Check if the email already exists
		if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
			log.Println("Email already in use")
			data := map[string]interface{}{
				"EmailError": "Email already in use",
				"Username":   username,
				"Email":      email,
			}
			renderTemplate(w, "register", data)
			return
		}

		// Check if the username already exists
		if err := db.Where("username = ?", username).First(&existingUser).Error; err == nil {
			log.Println("Username already in use")
			data := map[string]interface{}{
				"UsernameError": "Username already in use",
				"Username":      username,
				"Email":         email,
			}
			renderTemplate(w, "register", data)
			return
		}

		if err := validatePassword(password); err != nil {
			log.Printf("Password validation failed: %v", err)
			data := map[string]interface{}{
				"PasswordError": err.Error(),
				"Username":      username,
				"Email":         email,
			}
			renderTemplate(w, "register", data)
			return
		}

		data := map[string]interface{}{
			"Username": username,
			"Email":    email,
			"Password": password,
		}
		renderTemplate(w, "security_questions", data)
	} else {
		log.Println("Rendering registration step 1 template")
		renderTemplate(w, "register", nil)
	}
}

func RegisterStep2(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Register Step 2 handler")

	if r.Method == http.MethodPost {
		log.Println("Handling POST request for registration step 2")

		// Parse the form data
		if err := r.ParseForm(); err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		securityAnswer1 := r.FormValue("securityAnswer1")
		securityAnswer2 := r.FormValue("securityAnswer2")
		securityAnswer3 := r.FormValue("securityAnswer3")

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Error hashing password: %v", err)
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		// Create a new user instance
		user := models.User{
			Username:          username,
			Email:             email,
			Password:          string(hashedPassword),
			SecurityQuestion1: "What is your mother's maiden name?",
			SecurityAnswer1:   securityAnswer1,
			SecurityQuestion2: "What was the name of your first pet?",
			SecurityAnswer2:   securityAnswer2,
			SecurityQuestion3: "What is your favorite book?",
			SecurityAnswer3:   securityAnswer3,
		}

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
			renderTemplate(w, "security_questions", data)
			return
		}

		log.Printf("User registered successfully with ID: %d", user.ID)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
