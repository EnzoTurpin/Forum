package post

import (
	"encoding/json"
	"forum/common"
	"forum/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ViewPost gère l'affichage d'une publication spécifique
func ViewPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire ViewPost")

	// Récupération de l'ID de la publication depuis les variables de l'URL
	vars := mux.Vars(r)
	var post models.Post
	if err := common.DB.Preload("User").Preload("Comments.User").Preload("Categories").First(&post, vars["id"]).Error; err != nil {
		log.Printf("Publication introuvable : %v", err)
		http.NotFound(w, r)
		return
	}

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	user, ok := session.Values["user"]

	// Préparation des données à passer au template
	data := map[string]interface{}{
		"Post": post,
		"User": user,
	}
	if !ok {
		data["User"] = ""
	}

	// Rendu du template avec les données
	common.RenderTemplate(w, "view_post", data)
}

// LikePost gère l'ajout et la suppression des likes pour une publication
func LikePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire LikePost")

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	userID, ok := session.Values["userID"]
	if !ok {
		log.Println("Utilisateur non connecté, redirection vers la page de connexion")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Récupération de l'ID de la publication depuis les variables de l'URL
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("ID de publication invalide : %v", err)
		http.Error(w, "ID de publication invalide", http.StatusBadRequest)
		return
	}

	// Récupération de la publication depuis la base de données
	var post models.Post
	if err := common.DB.First(&post, postID).Error; err != nil {
		log.Printf("Publication introuvable : %v", err)
		http.Error(w, "Publication introuvable", http.StatusNotFound)
		return
	}

	// Vérification si l'utilisateur a déjà liké la publication
	var like models.Like
	result := common.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&like)
	if result.Error == nil {
		// Si l'utilisateur a déjà liké, supprimer le like
		common.DB.Delete(&like)
		post.Likes--
	} else {
		// Si l'utilisateur n'a pas encore liké, ajouter un like
		like = models.Like{UserID: userID.(uint), PostID: &post.ID}
		common.DB.Create(&like)
		post.Likes++
	}

	// Mise à jour du nombre de likes de la publication
	if err := common.DB.Save(&post).Error; err != nil {
		log.Printf("Impossible de mettre à jour le like de la publication : %v", err)
		http.Error(w, "Impossible de mettre à jour le like de la publication", http.StatusInternalServerError)
		return
	}

	log.Println("Like de la publication mis à jour avec succès")

	// Réponse JSON avec le nouveau nombre de likes
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"likes": post.Likes})
}

// ViewPostWithComments gère l'affichage d'une publication avec ses commentaires
func ViewPostWithComments(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire ViewPostWithComments")

	// Récupération de l'ID de la publication depuis les variables de l'URL
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("ID de publication invalide : %v", err)
		http.Error(w, "ID de publication invalide", http.StatusBadRequest)
		return
	}

	// Récupération de la publication avec ses commentaires depuis la base de données
	var post models.Post
	if err := common.DB.Preload("User").Preload("Comments.User").First(&post, postID).Error; err != nil {
		log.Printf("Publication introuvable : %v", err)
		http.Error(w, "Publication introuvable", http.StatusNotFound)
		return
	}

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	user, ok := session.Values["user"]

	// Préparation des données à passer au template
	data := map[string]interface{}{
		"Post": post,
		"User": user,
	}
	if !ok {
		data["User"] = ""
	}

	// Rendu du template avec les données
	common.RenderTemplate(w, "view_post", data)
}
