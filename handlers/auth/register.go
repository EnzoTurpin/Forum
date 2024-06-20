package auth

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register gère l'inscription des utilisateurs
func Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire d'inscription")

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour l'inscription")

		// Analyse du formulaire d'inscription
		if err := r.ParseForm(); err != nil {
			log.Printf("Erreur lors de l'analyse du formulaire : %v", err)
			http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusBadRequest)
			return
		}

		// Récupération des valeurs du formulaire
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Vérification de l'unicité du nom d'utilisateur ou de l'email
		var existingUser models.User
		if err := common.DB.Where("username = ? OR email = ?", username, email).First(&existingUser).Error; err != gorm.ErrRecordNotFound {
			log.Printf("Nom d'utilisateur ou Email déjà utilisé : %v", err)
			data := map[string]interface{}{
				"Error":    "Nom d'utilisateur ou Email déjà utilisé",
				"Username": username,
				"Email":    email,
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		// Validation du mot de passe
		if err := common.ValidatePassword(password); err != nil {
			log.Printf("Échec de la validation du mot de passe : %v", err)
			data := map[string]interface{}{
				"PasswordError": err.Error(),
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		// Hachage du mot de passe
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Erreur lors du hachage du mot de passe : %v", err)
			http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
			return
		}

		// Création de l'utilisateur avec les réponses aux questions de sécurité
		user := models.User{
			Username:          username,
			Email:             email,
			Password:          string(hashedPassword),
			SecurityQuestion1: r.FormValue("securityQuestion1"),
			SecurityAnswer1:   r.FormValue("securityAnswer1"),
			SecurityQuestion2: r.FormValue("securityQuestion2"),
			SecurityAnswer2:   r.FormValue("securityAnswer2"),
			SecurityQuestion3: r.FormValue("securityQuestion3"),
			SecurityAnswer3:   r.FormValue("securityAnswer3"),
		}

		// Enregistrement de l'utilisateur dans la base de données
		if err := common.DB.Create(&user).Error; err != nil {
			log.Printf("Impossible d'enregistrer l'utilisateur : %v", err)
			data := map[string]interface{}{
				"Error": "Impossible d'enregistrer l'utilisateur",
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		log.Printf("Utilisateur enregistré avec succès avec l'ID : %d", user.ID)
		// Redirection vers la page de connexion après inscription réussie
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		// Affichage de la page d'inscription
		common.RenderTemplate(w, "register", nil)
	}
}
