package general

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ViewAllComments gère l'affichage de tous les commentaires pour une publication donnée
func ViewAllComments(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire ViewAllComments")

	// Récupération de l'ID de la publication depuis les variables de l'URL
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["postID"])
	if err != nil {
		log.Printf("ID de publication invalide : %v", err)
		http.Error(w, "ID de publication invalide", http.StatusBadRequest)
		return
	}

	// Récupération de la publication avec les commentaires associés depuis la base de données
	var post models.Post
	if err := common.DB.Preload("User").Preload("Comments.User").First(&post, postID).Error; err != nil {
		log.Printf("Publication introuvable : %v", err)
		http.Error(w, "Publication introuvable", http.StatusNotFound)
		return
	}

	// Préparation des données à passer au template
	data := map[string]interface{}{
		"Post": post,
	}

	// Rendu du template avec les données
	common.RenderTemplate(w, "view_all_comments", data)
}
