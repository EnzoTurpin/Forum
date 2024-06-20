package profile

import (
	"fmt"
	"forum/common"
	"forum/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// EditProfile gère la modification des profils utilisateurs
func EditProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Début du gestionnaire EditProfile")

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

	// Vérification que l'utilisateur actuel a la permission de modifier ce profil
	if user.ID != currentUserID.(uint) {
		log.Println("L'utilisateur n'a pas la permission de modifier ce profil")
		http.Error(w, "Vous n'avez pas la permission de modifier ce profil", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		log.Println("Gestion de la requête POST pour la modification de profil")

		// Limite la taille de la requête et analyse le formulaire multipart
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		r.ParseMultipartForm(32 << 20)

		newUsername := r.FormValue("username")
		newEmail := r.FormValue("email")
		password := r.FormValue("password")
		log.Printf("Valeurs du formulaire analysées - Nom d'utilisateur : %s, Email : %s, Mot de passe : %s", newUsername, newEmail, password)

		// Vérification de la longueur du nom d'utilisateur
		if len(newUsername) > 18 {
			log.Println("Nom d'utilisateur trop long")
			data := map[string]interface{}{
				"User":          user,
				"UsernameError": "Le nom d'utilisateur ne doit pas dépasser 18 caractères",
			}
			common.RenderTemplate(w, "edit_profile", data)
			return
		}

		// Vérification de l'unicité du nom d'utilisateur
		var existingUser models.User
		if err := common.DB.Where("username = ? AND id != ?", newUsername, user.ID).First(&existingUser).Error; err == nil {
			log.Println("Nom d'utilisateur déjà utilisé")
			data := map[string]interface{}{
				"User":          user,
				"UsernameError": "Nom d'utilisateur déjà utilisé",
			}
			common.RenderTemplate(w, "edit_profile", data)
			return
		}

		// Vérification de l'unicité de l'email
		if err := common.DB.Where("email = ? AND id != ?", newEmail, user.ID).First(&existingUser).Error; err == nil {
			log.Println("Email déjà utilisé par un autre utilisateur")
			data := map[string]interface{}{
				"User":       user,
				"EmailError": "Email déjà utilisé",
			}
			common.RenderTemplate(w, "edit_profile", data)
			return
		}

		// Mise à jour des informations de l'utilisateur
		user.Username = newUsername
		user.Email = newEmail

		// Si un nouveau mot de passe est fourni, le valider et le hacher
		if password != "" {
			if err := common.ValidatePassword(password); err != nil {
				log.Printf("Échec de la validation du mot de passe : %v", err)
				data := map[string]interface{}{
					"User":          user,
					"PasswordError": err.Error(),
				}
				common.RenderTemplate(w, "edit_profile", data)
				return
			}
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Erreur lors du hachage du mot de passe : %v", err)
				http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
				return
			}
			user.Password = string(hashedPassword)
		}

		// Enregistrement des modifications dans la base de données
		if err := common.DB.Save(&user).Error; err != nil {
			log.Printf("Impossible de mettre à jour le profil : %v", err)
			http.Error(w, "Impossible de mettre à jour le profil : "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Mise à jour de la session avec le nouveau nom d'utilisateur
		session.Values["user"] = user.Username
		session.Save(r, w)
		log.Println("Profil mis à jour avec succès et session enregistrée")
		http.Redirect(w, r, fmt.Sprintf("/profile/%s", user.Username), http.StatusSeeOther)
		return
	}

	log.Println("Rendu du modèle de modification de profil")
	data := map[string]interface{}{
		"User": user,
	}
	common.RenderTemplate(w, "edit_profile", data)
}
