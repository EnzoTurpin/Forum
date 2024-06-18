package handlers

import (
	"fmt"
	"forum/models"
	"html/template"
	"log"
	"math/rand"
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

		// Sélectionner une question de sécurité aléatoire
		questions := []string{
			user.SecurityQuestion1,
			user.SecurityQuestion2,
			user.SecurityQuestion3,
		}
		answers := []string{
			user.SecurityAnswer1,
			user.SecurityAnswer2,
			user.SecurityAnswer3,
		}

		rand.Seed(time.Now().UnixNano())
		idx := rand.Intn(len(questions))

		// Stocker la question sélectionnée dans la session
		session, _ := store.Get(r, "session")
		session.Values["reset_email"] = email
		session.Values["security_question"] = questions[idx]
		session.Values["security_answer"] = answers[idx]
		session.Save(r, w)

		// Passer la question sélectionnée au template
		data := map[string]interface{}{
			"Email":    email,
			"Question": questions[idx],
		}
		renderTemplate(w, "security_questions_reset", data)
		return
	}
	renderTemplate(w, "forgot_password", nil)
}

func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		session, _ := store.Get(r, "session")
		email := session.Values["reset_email"].(string)
		correctAnswer := session.Values["security_answer"].(string)
		question := session.Values["security_question"].(string)

		securityAnswer := r.FormValue("securityAnswer")
		newPassword := r.FormValue("newPassword")

		var user models.User
		if err := db.Where("email = ?", email).First(&user).Error; err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Vérifier que la réponse de sécurité est correcte
		if correctAnswer != securityAnswer {
			data := map[string]interface{}{
				"Email":         email,
				"Question":      question,
				"SecurityError": "Security answer does not match",
				"PasswordError": "",
			}
			renderTemplate(w, "security_questions_reset", data)
			return
		}

		// Vérifier que le nouveau mot de passe n'est pas le même que l'actuel
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword)) == nil {
			data := map[string]interface{}{
				"Email":         email,
				"Question":      question,
				"SecurityError": "",
				"PasswordError": "New password cannot be the same as the current password",
			}
			renderTemplate(w, "security_questions_reset", data)
			return
		}

		// Hasher le nouveau mot de passe
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
