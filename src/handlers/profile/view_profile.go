package profile

import (
	"forum/src/common"
	"forum/src/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// ViewProfile gère l'affichage du profil utilisateur
func ViewProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire ViewProfile")

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

	// Récupération des publications de l'utilisateur
	var posts []models.Post
	common.DB.Where("user_id = ?", user.ID).Find(&posts)
	for i := range posts {
		posts[i].TimeAgo = common.FormatTimeAgo(posts[i].CreatedAt)
	}

	// Récupération des abonnés et des utilisateurs suivis par l'utilisateur
	var followers []models.Follower
	var following []models.Follower
	common.DB.Preload("Follower").Where("follows_id = ?", user.ID).Find(&followers)
	common.DB.Preload("Follows").Where("follower_id = ?", user.ID).Find(&following)

	// Préparation des données à passer au template
	data := map[string]interface{}{
		"ProfileUser":    user,
		"Posts":          posts,
		"Followers":      followers,
		"Following":      following,
		"FollowersCount": len(followers),
		"FollowingCount": len(following),
	}

	// Récupération de la session utilisateur
	session, err := common.Store.Get(r, "session")
	if err != nil {
		log.Printf("Impossible de récupérer la session : %v", err)
		http.Error(w, "Impossible de récupérer la session : "+err.Error(), http.StatusInternalServerError)
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

	// Vérification si l'utilisateur actuel suit déjà l'utilisateur du profil
	var follower models.Follower
	if common.DB.Where("follower_id = ? AND follows_id = ?", currentUserID, user.ID).First(&follower).Error == nil {
		data["IsFollowing"] = true
	} else {
		data["IsFollowing"] = false
	}

	// Rendu du template avec les données
	common.RenderTemplate(w, "profile", data)
}
