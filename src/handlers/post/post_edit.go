package post

import (
	"forum/src/common"
	"forum/src/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// ShowEditPostForm affiche le formulaire de modification d'une publication
func ShowEditPostForm(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire ShowEditPostForm")

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

	// Récupération de la publication et de ses catégories depuis la base de données
	var post models.Post
	if err := common.DB.Preload("Categories").First(&post, postID).Error; err != nil {
		log.Printf("Publication introuvable : %v", err)
		http.NotFound(w, r)
		return
	}

	// Vérification que l'utilisateur est le propriétaire de la publication
	if post.UserID != userID.(uint) {
		log.Println("Tentative non autorisée de modifier la publication")
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Récupération de toutes les catégories pour les afficher dans le formulaire
	var categories []models.Category
	if err := common.DB.Find(&categories).Error; err != nil {
		log.Printf("Impossible de récupérer les catégories : %v", err)
		http.Error(w, "Impossible de récupérer les catégories", http.StatusInternalServerError)
		return
	}

	// Rendu du template avec les données de la publication et des catégories
	common.RenderTemplate(w, "edit_post", map[string]interface{}{
		"Post":       post,
		"Categories": categories,
	})
}

// EditPost gère la modification d'une publication existante
func EditPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire EditPost")

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
	if err := common.DB.Preload("Categories").First(&post, postID).Error; err != nil {
		log.Printf("Publication introuvable : %v", err)
		http.NotFound(w, r)
		return
	}

	// Vérification que l'utilisateur est le propriétaire de la publication
	if post.UserID != userID.(uint) {
		log.Println("Tentative non autorisée de modifier la publication")
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()

		// Récupération des catégories sélectionnées
		var categories []models.Category
		categoryIDs := r.Form["categories[]"]
		for _, idStr := range categoryIDs {
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				continue
			}
			var category models.Category
			if common.DB.First(&category, uint(id)).Error == nil {
				categories = append(categories, category)
			}
		}

		// Mise à jour des informations de la publication
		post.Title = r.FormValue("title")
		post.Content = r.FormValue("content")
		post.Categories = categories

		// Enregistrement des modifications dans la base de données
		if err := common.DB.Save(&post).Error; err != nil {
			log.Printf("Impossible de mettre à jour la publication : %v", err)
			http.Error(w, "Impossible de mettre à jour la publication", http.StatusInternalServerError)
			return
		}

		// Redirection vers la page de la publication après modification réussie
		http.Redirect(w, r, "/post/"+vars["id"], http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode de requête invalide", http.StatusMethodNotAllowed)
	}
}

// DeletePost gère la suppression d'une publication
func DeletePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire DeletePost")

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
		http.NotFound(w, r)
		return
	}

	// Vérification que l'utilisateur est le propriétaire de la publication
	if post.UserID != userID.(uint) {
		log.Println("Tentative non autorisée de supprimer la publication")
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Suppression de la publication dans la base de données
	if err := common.DB.Delete(&post).Error; err != nil {
		log.Printf("Impossible de supprimer la publication : %v", err)
		http.Error(w, "Impossible de supprimer la publication", http.StatusInternalServerError)
		return
	}

	log.Println("Publication supprimée avec succès")
	// Redirection vers la page d'accueil après suppression réussie
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
