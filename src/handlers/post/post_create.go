package post

import (
	"forum/src/common"
	"forum/src/models"
	"log"
	"net/http"
	"strconv"

	"gorm.io/gorm"
)

// CreatePost gère la création de nouvelles publications
func CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire CreatePost")

	// Récupération de la session utilisateur
	session, _ := common.Store.Get(r, "session")
	userID, ok := session.Values["userID"]
	user, userOk := session.Values["user"]

	// Vérification de la connexion de l'utilisateur
	if !ok || !userOk {
		log.Println("Utilisateur non connecté, redirection vers la page de connexion")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour créer une publication")

		// Analyse du formulaire de création de publication
		r.ParseForm()
		title := r.FormValue("title")
		content := r.FormValue("content")
		categoryIDs := r.Form["categories[]"]

		// Création de l'objet publication
		post := models.Post{
			Title:   title,
			Content: content,
			UserID:  userID.(uint),
		}

		// Ajout des catégories à la publication
		for _, c := range categoryIDs {
			id, err := strconv.ParseUint(c, 10, 32)
			if err != nil {
				log.Printf("ID de catégorie invalide : %v", err)
				continue
			}
			post.Categories = append(post.Categories, models.Category{Model: gorm.Model{ID: uint(id)}})
		}

		// Enregistrement de la publication dans la base de données
		result := common.DB.Create(&post)
		if result.Error != nil {
			log.Printf("Impossible de créer la publication : %v", result.Error)
			http.Error(w, "Impossible de créer la publication", http.StatusInternalServerError)
			return
		}

		log.Println("Publication créée avec succès")
		// Redirection vers la page d'accueil après création réussie
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		log.Println("Rendu du modèle de création de publication")

		// Récupération des catégories pour l'affichage dans le formulaire de création
		var categories []models.Category
		if err := common.DB.Find(&categories).Error; err != nil {
			log.Printf("Impossible de récupérer les catégories : %v", err)
			http.Error(w, "Impossible de récupérer les catégories", http.StatusInternalServerError)
			return
		}

		// Préparation des données à passer au template
		data := map[string]interface{}{
			"User":       user,
			"Categories": categories,
		}
		// Rendu du template pour la création de publication
		common.RenderTemplate(w, "create_post", data)
	}
}
