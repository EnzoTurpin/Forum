package general

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// CreateComment gère la création de commentaires sur les publications
func CreateComment(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire CreateComment")

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	userID, ok := session.Values["userID"]
	if !ok {
		log.Println("Utilisateur non connecté, redirection vers la page de connexion")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour créer un commentaire")

		// Récupération de l'ID de la publication depuis les variables de l'URL
		vars := mux.Vars(r)
		postIDStr := vars["id"]
		postID, err := strconv.ParseUint(postIDStr, 10, 32)
		if err != nil {
			log.Printf("ID de publication invalide : %v", err)
			http.Error(w, "ID de publication invalide", http.StatusBadRequest)
			return
		}

		// Analyse du formulaire pour récupérer le contenu du commentaire
		r.ParseForm()
		comment := models.Comment{
			Content: r.FormValue("content"),
			PostID:  uint(postID),
			UserID:  userID.(uint),
		}

		// Enregistrement du commentaire dans la base de données
		result := common.DB.Create(&comment)
		if result.Error != nil {
			log.Printf("Impossible de créer le commentaire : %v", result.Error)
			http.Error(w, "Impossible de créer le commentaire", http.StatusInternalServerError)
			return
		}

		log.Println("Commentaire créé avec succès")
		// Redirection vers la page d'accueil après création réussie
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
