package auth

import (
	"forum/common"
	"forum/models"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterStep1 gère la première étape de l'inscription des utilisateurs
func RegisterStep1(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire d'inscription Étape 1")

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour l'inscription étape 1")

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

		// Vérification que les champs requis ne sont pas vides
		if username == "" || email == "" || password == "" {
			log.Println("Valeurs de formulaire requises manquantes")
			data := map[string]interface{}{
				"Error": "Valeurs de formulaire requises manquantes",
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		// Vérification de la longueur du nom d'utilisateur
		if len(username) > 18 {
			log.Println("Nom d'utilisateur trop long")
			data := map[string]interface{}{
				"UsernameError": "Le nom d'utilisateur ne doit pas dépasser 18 caractères",
				"Username":      username,
				"Email":         email,
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		// Vérification de l'unicité de l'email
		var existingUser models.User
		if err := common.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
			log.Println("Email déjà utilisé")
			data := map[string]interface{}{
				"EmailError": "Email déjà utilisé",
				"Username":   username,
				"Email":      email,
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		// Vérification de l'unicité du nom d'utilisateur
		if err := common.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
			log.Println("Nom d'utilisateur déjà utilisé")
			data := map[string]interface{}{
				"UsernameError": "Nom d'utilisateur déjà utilisé",
				"Username":      username,
				"Email":         email,
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		// Validation du mot de passe
		if err := common.ValidatePassword(password); err != nil {
			log.Printf("Échec de la validation du mot de passe : %v", err)
			data := map[string]interface{}{
				"PasswordError": err.Error(),
				"Username":      username,
				"Email":         email,
			}
			common.RenderTemplate(w, "register", data)
			return
		}

		// Passer à l'étape suivante avec les données collectées
		data := map[string]interface{}{
			"Username": username,
			"Email":    email,
			"Password": password,
		}
		common.RenderTemplate(w, "security_questions", data)
	} else {
		log.Println("Rendu du modèle d'inscription étape 1")
		common.RenderTemplate(w, "register", nil)
	}
}

// RegisterStep2 gère la deuxième étape de l'inscription des utilisateurs
func RegisterStep2(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire d'inscription Étape 2")

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour l'inscription étape 2")

		// Analyse du formulaire de la deuxième étape d'inscription
		if err := r.ParseForm(); err != nil {
			log.Printf("Erreur lors de l'analyse du formulaire : %v", err)
			http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusBadRequest)
			return
		}

		// Récupération des valeurs du formulaire
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		securityAnswer1 := r.FormValue("securityAnswer1")
		securityAnswer2 := r.FormValue("securityAnswer2")
		securityAnswer3 := r.FormValue("securityAnswer3")

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
			SecurityQuestion1: "Quel est le nom de jeune fille de votre mère ?",
			SecurityAnswer1:   securityAnswer1,
			SecurityQuestion2: "Quel était le nom de votre premier animal de compagnie ?",
			SecurityAnswer2:   securityAnswer2,
			SecurityQuestion3: "Quel est votre livre préféré ?",
			SecurityAnswer3:   securityAnswer3,
		}

		// Enregistrement de l'utilisateur dans une transaction
		result := common.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&user).Error; err != nil {
				return err
			}
			return nil
		})

		// Vérification de l'enregistrement réussi
		if result != nil {
			log.Printf("Impossible d'enregistrer l'utilisateur : %v", result)
			data := map[string]interface{}{
				"Error": "Impossible d'enregistrer l'utilisateur",
			}
			common.RenderTemplate(w, "security_questions", data)
			return
		}

		log.Printf("Utilisateur enregistré avec succès avec l'ID : %d", user.ID)
		// Redirection vers la page de connexion après inscription réussie
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
