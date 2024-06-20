package general

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"
)

// PageIndex gère l'affichage de la page d'accueil avec la liste des publications
func PageIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire PageIndex")

	// Récupération de la session utilisateur
	session, err := common.Store.Get(r, "session")
	if err != nil {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
	user, ok := session.Values["user"]

	// Récupération des publications depuis la base de données
	var posts []models.Post
	categoryIDs := r.URL.Query()["categories"]
	if len(categoryIDs) > 0 {
		// Si des catégories sont spécifiées, filtrer les publications par ces catégories
		var categories []models.Category
		if err := common.DB.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
			http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
			return
		}
		if len(categories) > 0 {
			categoryCount := len(categories)
			common.DB.Preload("User").Preload("Comments.User").Preload("Categories").
				Joins("JOIN post_categories ON post_categories.post_id = posts.id").
				Where("post_categories.category_id IN ?", categoryIDs).
				Group("posts.id").
				Having("COUNT(post_categories.category_id) = ?", categoryCount).
				Order("created_at desc").
				Find(&posts)
		}
	} else {
		// Si aucune catégorie n'est spécifiée, récupérer toutes les publications
		common.DB.Preload("User").Preload("Comments.User").Preload("Categories").
			Order("created_at desc").
			Find(&posts)
	}

	// Formatage du temps écoulé pour chaque publication
	for i := range posts {
		posts[i].TimeAgo = common.FormatTimeAgo(posts[i].CreatedAt)
	}

	// Récupération de la liste des catégories pour l'affichage dans le template
	var categoriesList []models.Category
	if err := common.DB.Find(&categoriesList).Error; err == nil {
		// Préparation des données à passer au template
		data := map[string]interface{}{
			"User":       user,
			"Posts":      posts,
			"Categories": categoriesList,
		}
		if !ok {
			data["User"] = ""
		}
		log.Println("Rendu du modèle index")
		// Rendu du template avec les données
		common.RenderTemplate(w, "index", data)
	} else {
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}
