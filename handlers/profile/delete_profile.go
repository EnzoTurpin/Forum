package profile

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// DeleteProfile gère la suppression des profils utilisateurs
func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire DeleteProfile")

	// Récupération de la session utilisateur
	session, err := common.Store.Get(r, "session")
	if err != nil {
		log.Printf("Impossible de récupérer la session : %v", err)
		http.Error(w, "Impossible de récupérer la session : "+err.Error(), http.StatusInternalServerError)
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

	// Vérification que l'utilisateur actuel a la permission de supprimer ce profil
	if user.ID != currentUserID.(uint) {
		log.Println("L'utilisateur n'a pas la permission de supprimer ce profil")
		http.Error(w, "Vous n'avez pas la permission de supprimer ce profil", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour la suppression du profil")

		// Démarrer une transaction
		tx := common.DB.Begin()

		// Suppression des publications de l'utilisateur
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.Post{}).Error; err != nil {
			log.Printf("Erreur lors de la suppression des publications de l'utilisateur : %v", err)
			tx.Rollback()
			http.Error(w, "Erreur lors de la suppression des publications de l'utilisateur", http.StatusInternalServerError)
			return
		}

		// Suppression des abonnements de l'utilisateur
		if err := tx.Where("follower_id = ? OR follows_id = ?", user.ID, user.ID).Unscoped().Delete(&models.Follower{}).Error; err != nil {
			log.Printf("Erreur lors de la suppression des abonnements de l'utilisateur : %v", err)
			tx.Rollback()
			http.Error(w, "Erreur lors de la suppression des abonnements de l'utilisateur", http.StatusInternalServerError)
			return
		}

		// Suppression de l'utilisateur
		if err := tx.Unscoped().Delete(&user).Error; err != nil {
			log.Printf("Erreur lors de la suppression de l'utilisateur : %v", err)
			tx.Rollback()
			http.Error(w, "Erreur lors de la suppression de l'utilisateur", http.StatusInternalServerError)
			return
		}

		// Validation de la transaction
		if err := tx.Commit().Error; err != nil {
			log.Printf("Échec de la validation de la transaction : %v", err)
			http.Error(w, "Erreur pendant la transaction", http.StatusInternalServerError)
			return
		}

		// Suppression des valeurs de session et enregistrement de la session
		delete(session.Values, "user")
		delete(session.Values, "userID")
		if err := session.Save(r, w); err != nil {
			log.Printf("Erreur lors de l'enregistrement de la session : %v", err)
			http.Error(w, "Erreur lors de l'enregistrement de la session", http.StatusInternalServerError)
			return
		}

		log.Println("Profil utilisateur et publications supprimés avec succès")
		// Redirection vers la page d'accueil après suppression réussie
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
