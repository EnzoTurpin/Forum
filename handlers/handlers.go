package handlers

import (
	"fmt"
	"forum/models"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var db *gorm.DB
var store *sessions.CookieStore

// SetDB sets the database connection to be used by handlers
func SetDB(database *gorm.DB) {
	db = database
}

// SetStore sets the session store to be used by handlers
func SetStore(s *sessions.CookieStore) {
	store = s
}

// renderTemplate renders HTML templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmpl = "./templates/" + tmpl + ".html"
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

// formatTimeAgo formats the duration since the post was created
func formatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d.Seconds() < 60:
		return fmt.Sprintf("il y a %.f secondes", d.Seconds())
	case d.Minutes() < 60:
		return fmt.Sprintf("il y a %.f minutes", d.Minutes())
	case d.Hours() < 24:
		return fmt.Sprintf("il y a %.f heures", d.Hours())
	case d.Hours() < 48:
		return "il y a 1 jour"
	default:
		return fmt.Sprintf("il y a %.f jours", d.Hours()/24)
	}
}

// PageIndex handles the rendering of the index page
func PageIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("Rendering index page")
	session, _ := store.Get(r, "session")
	user, ok := session.Values["user"]
	data := map[string]interface{}{
		"User": user,
	}
	if !ok {
		data["User"] = ""
	}

	var posts []models.Post
	categories := r.URL.Query()["categories"]
	if len(categories) > 0 {
		db.Preload("User").Preload("Comments.User").Where("category_id IN (?)", categories).Find(&posts)
	} else {
		db.Preload("User").Preload("Comments.User").Find(&posts)
	}

	for i := range posts {
		posts[i].TimeAgo = formatTimeAgo(posts[i].CreatedAt)
	}
	data["Posts"] = posts

	var categoriesList []models.Category
	if err := db.Find(&categoriesList).Error; err == nil {
		data["Categories"] = categoriesList
	}

	renderTemplate(w, "index", data)
}

// ForgotPassword handles the process of requesting a password reset
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")

		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Render the security questions page with the user data
		data := map[string]interface{}{
			"Email": email,
		}
		renderTemplate(w, "security_questions_reset", data)
		return
	}
	renderTemplate(w, "forgot_password", nil)
}

// ResetPassword handles the process of resetting the password
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		securityAnswer1 := r.FormValue("securityAnswer1")
		securityAnswer2 := r.FormValue("securityAnswer2")
		securityAnswer3 := r.FormValue("securityAnswer3")
		newPassword := r.FormValue("newPassword")

		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		if user.SecurityAnswer1 != securityAnswer1 || user.SecurityAnswer2 != securityAnswer2 || user.SecurityAnswer3 != securityAnswer3 {
			http.Error(w, "Security answers do not match", http.StatusUnauthorized)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		user.Password = string(hashedPassword)
		if err := db.Save(&user).Error; err != nil {
			http.Error(w, "Error updating password", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
