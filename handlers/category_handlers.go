package handlers

import (
	"forum/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// PostsByCategory handles the retrieval and rendering of posts by category
func PostsByCategory(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting PostsByCategory handler")

	vars := mux.Vars(r)
	category := vars["category"]

	var posts []models.Post
	if err := db.Preload("User").Preload("Comments.User").Where("category = ?", category).Find(&posts).Error; err != nil {
		log.Printf("Category not found: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	session, _ := store.Get(r, "session")
	user, ok := session.Values["user"]
	data := map[string]interface{}{
		"Posts":    posts,
		"User":     user,
		"Category": category,
	}
	if !ok {
		data["User"] = ""
	}

	renderTemplate(w, "category_posts", data)
}

// Categories handles the retrieval and rendering of all available categories
func Categories(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Categories handler")

	var categories []models.Category
	if err := db.Find(&categories).Error; err != nil {
		log.Printf("Unable to fetch categories: %v", err)
		http.Error(w, "Unable to fetch categories", http.StatusInternalServerError)
		return
	}

	session, _ := store.Get(r, "session")
	user, ok := session.Values["user"]
	data := map[string]interface{}{
		"Categories": categories,
		"User":       user,
	}
	if !ok {
		data["User"] = ""
	}

	renderTemplate(w, "index", data)
}
