package handlers

import (
	"forum/models"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
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
	data["Posts"] = posts

	var categoriesList []models.Category
	if err := db.Find(&categoriesList).Error; err == nil {
		data["Categories"] = categoriesList
	}

	renderTemplate(w, "index", data)
}
