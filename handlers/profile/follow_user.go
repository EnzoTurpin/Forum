package profile

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// FollowUser gère le suivi d'un utilisateur
func FollowUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire FollowUser")

	// Récupération de la session utilisateur
	session, err := common.Store.Get(r, "session")
	if err != nil {
		log.Printf("Impossible de récupérer la session : %v", err)
		http.Error(w, "Impossible de récupérer la session", http.StatusInternalServerError)
		return
	}

	// Vérification de la connexion de l'utilisateur
	currentUserID, ok := session.Values["userID"]
	if !ok {
		log.Println("Utilisateur non connecté, redirection vers la page de connexion")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Récupération du nom d'utilisateur depuis les variables de l'URL
	vars := mux.Vars(r)
	username := vars["username"]

	// Recherche de l'utilisateur par nom d'utilisateur dans la base de données
	var user models.User
	if err := common.DB.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("Utilisateur introuvable : %v", err)
		http.NotFound(w, r)
		return
	}

	// Vérification si l'utilisateur suit déjà cet utilisateur
	var follower models.Follower
	if err := common.DB.Where("follower_id = ? AND follows_id = ?", currentUserID, user.ID).First(&follower).Error; err == nil {
		log.Println("L'utilisateur suit déjà")
		http.Redirect(w, r, "/profile/"+username, http.StatusSeeOther)
		return
	}

	// Création du lien de suivi
	follower = models.Follower{
		FollowerID: currentUserID.(uint),
		FollowsID:  user.ID,
	}
	if err := common.DB.Create(&follower).Error; err != nil {
		log.Printf("Impossible de suivre l'utilisateur : %v", err)
		http.Error(w, "Impossible de suivre l'utilisateur", http.StatusInternalServerError)
		return
	}

	log.Println("Utilisateur suivi avec succès")
	// Redirection vers la page de profil après succès
	http.Redirect(w, r, "/profile/"+username, http.StatusSeeOther)
}
