package general

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"
)

// Categories gère la récupération et l'affichage des catégories
func Categories(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire Categories")

	// Récupération de toutes les catégories depuis la base de données
	var categories []models.Category
	if err := common.DB.Find(&categories).Error; err != nil {
		log.Printf("Impossible de récupérer les catégories : %v", err)
		http.Error(w, "Impossible de récupérer les catégories", http.StatusInternalServerError)
		return
	}

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	user, ok := session.Values["user"]

	// Préparation des données à passer au template
	data := map[string]interface{}{
		"Categories": categories,
		"User":       user,
	}
	if !ok {
		data["User"] = ""
	}

	// Rendu du template avec les données
	common.RenderTemplate(w, "index", data)
}
