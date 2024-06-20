package profile

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// ViewFollowers gère l'affichage des abonnés d'un utilisateur
func ViewFollowers(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire ViewFollowers")

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

	// Récupération des abonnés de l'utilisateur
	var followers []models.Follower
	common.DB.Preload("Follower").Where("follows_id = ?", user.ID).Find(&followers)

	// Préparation des données à passer au template
	data := map[string]interface{}{
		"ProfileUser": user,
		"Followers":   followers,
	}

	// Récupération de la session utilisateur
	session, err := common.Store.Get(r, "session")
	if err != nil {
		log.Printf("Impossible de récupérer la session : %v", err)
		http.Error(w, "Impossible de récupérer la session", http.StatusInternalServerError)
		return
	}

	// Ajout des informations sur l'utilisateur actuel à la session
	currentUser, ok := session.Values["user"]
	currentUserID := session.Values["userID"]
	if ok {
		data["CurrentUser"] = currentUser
		data["CurrentUserID"] = currentUserID
	} else {
		data["CurrentUser"] = ""
		data["CurrentUserID"] = uint(0)
	}

	// Rendu du template avec les données
	common.RenderTemplate(w, "followers", data)
}
