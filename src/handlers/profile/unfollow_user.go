package profile

import (
	"forum/src/common"
	"forum/src/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// UnfollowUser gère le désabonnement d'un utilisateur
func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire UnfollowUser")

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

	// Suppression du lien de suivi
	if err := common.DB.Where("follower_id = ? AND follows_id = ?", currentUserID, user.ID).Delete(&models.Follower{}).Error; err != nil {
		log.Printf("Impossible de se désabonner de l'utilisateur : %v", err)
		http.Error(w, "Impossible de se désabonner de l'utilisateur", http.StatusInternalServerError)
		return
	}

	log.Println("Désabonnement de l'utilisateur réussi")
	// Redirection vers la page de profil après succès
	http.Redirect(w, r, "/profile/"+username, http.StatusSeeOther)
}
