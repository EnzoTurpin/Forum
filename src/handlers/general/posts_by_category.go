package general

import (
	"forum/src/common"
	"forum/src/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// PostsByCategory gère l'affichage des publications par catégorie
func PostsByCategory(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire PostsByCategory")

	// Récupération de la catégorie depuis les variables de l'URL
	vars := mux.Vars(r)
	category := vars["category"]

	// Récupération des publications appartenant à la catégorie spécifiée
	var posts []models.Post
	if err := common.DB.Preload("User").Preload("Comments.User").Where("category = ?", category).Order("created_at desc").Find(&posts).Error; err != nil {
		log.Printf("Catégorie introuvable : %v", err)
		http.Error(w, "Catégorie introuvable", http.StatusNotFound)
		return
	}

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	user, ok := session.Values["user"]

	// Préparation des données à passer au template
	data := map[string]interface{}{
		"Posts":    posts,
		"User":     user,
		"Category": category,
	}
	if !ok {
		data["User"] = ""
	}

	// Rendu du template avec les données
	common.RenderTemplate(w, "category_posts", data)
}
