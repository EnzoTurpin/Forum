package general

import (
	"encoding/json"
	"forum/common"
	"forum/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// LikeComment gère les likes et unlikes sur les commentaires
func LikeComment(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire LikeComment")

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	userID, ok := session.Values["userID"]
	if !ok {
		log.Println("Utilisateur non connecté, redirection vers la page de connexion")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Récupération de l'ID du commentaire depuis les variables de l'URL
	vars := mux.Vars(r)
	commentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("ID de commentaire invalide : %v", err)
		http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
		return
	}

	// Récupération du commentaire depuis la base de données
	var comment models.Comment
	if err := common.DB.First(&comment, commentID).Error; err != nil {
		log.Printf("Commentaire introuvable : %v", err)
		http.Error(w, "Commentaire introuvable", http.StatusNotFound)
		return
	}

	// Vérification si l'utilisateur a déjà liké le commentaire
	var like models.Like
	result := common.DB.Where("user_id = ? AND comment_id = ?", userID, commentID).First(&like)
	if result.Error == nil {
		// Si l'utilisateur a déjà liké, supprimer le like
		common.DB.Delete(&like)
		comment.Likes--
	} else {
		// Si l'utilisateur n'a pas encore liké, ajouter un like
		like = models.Like{UserID: userID.(uint), CommentID: &comment.ID}
		common.DB.Create(&like)
		comment.Likes++
	}

	// Mise à jour du nombre de likes du commentaire
	if err := common.DB.Save(&comment).Error; err != nil {
		log.Printf("Impossible de mettre à jour le like du commentaire : %v", err)
		http.Error(w, "Impossible de mettre à jour le like du commentaire", http.StatusInternalServerError)
		return
	}

	log.Println("Like du commentaire mis à jour avec succès")

	// Réponse JSON avec le nouveau nombre de likes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"likes": comment.Likes})
}
