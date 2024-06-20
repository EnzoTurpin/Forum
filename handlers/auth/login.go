package auth

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Login gère les requêtes de connexion des utilisateurs
func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire de connexion")

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour la connexion")

		// Analyse du formulaire de connexion
		if err := r.ParseForm(); err != nil {
			log.Printf("Erreur lors de l'analyse du formulaire : %v", err)
			http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusBadRequest)
			return
		}

		// Récupération des valeurs du formulaire
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Vérification que les champs email et mot de passe ne sont pas vides
		if email == "" || password == "" {
			log.Println("Email ou mot de passe manquant")
			http.Error(w, "Email ou mot de passe manquant", http.StatusBadRequest)
			return
		}

		// Recherche de l'utilisateur dans la base de données par email
		var user models.User
		result := common.DB.Where("email = ?", email).First(&user)
		if result.Error != nil {
			log.Printf("Email ou mot de passe invalide : %v", result.Error)
			http.Error(w, "Email ou mot de passe invalide", http.StatusUnauthorized)
			return
		}

		// Vérification du mot de passe
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
			log.Printf("Email ou mot de passe invalide : %v", err)
			http.Error(w, "Email ou mot de passe invalide", http.StatusUnauthorized)
			return
		}

		// Création de la session utilisateur
		session, _ := common.Store.Get(r, "session")
		session.Values["user"] = user.Username
		session.Values["userID"] = user.ID
		if err := session.Save(r, w); err != nil {
			log.Printf("Erreur lors de l'enregistrement de la session : %v", err)
			http.Error(w, "Erreur lors de l'enregistrement de la session", http.StatusInternalServerError)
			return
		}

		log.Println("Utilisateur connecté avec succès avec le nom d'utilisateur :", user.Username)
		// Redirection vers la page d'accueil après connexion réussie
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		log.Println("Rendu du modèle de connexion")
		// Affichage de la page de connexion
		common.RenderTemplate(w, "login", nil)
	}
}

// Logout gère les requêtes de déconnexion des utilisateurs
func Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire de déconnexion")

	// Récupération de la session
	session, _ := common.Store.Get(r, "session")

	// Suppression des valeurs de session
	delete(session.Values, "user")
	delete(session.Values, "userID")
	if err := session.Save(r, w); err != nil {
		log.Printf("Erreur lors de l'enregistrement de la session : %v", err)
		http.Error(w, "Erreur lors de l'enregistrement de la session", http.StatusInternalServerError)
		return
	}

	log.Println("Utilisateur déconnecté avec succès")
	// Redirection vers la page d'accueil après déconnexion
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
