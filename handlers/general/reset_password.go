package general

import (
	"forum/common"
	"forum/models"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// ResetPassword gère la réinitialisation du mot de passe des utilisateurs
func ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Récupération de la session utilisateur
		session, _ := common.Store.Get(r, "session")
		email := session.Values["reset_email"].(string)
		correctAnswer := session.Values["security_answer"].(string)
		question := session.Values["security_question"].(string)

		// Récupération des valeurs du formulaire
		securityAnswer := r.FormValue("securityAnswer")
		newPassword := r.FormValue("newPassword")

		// Recherche de l'utilisateur par email dans la base de données
		var user models.User
		if err := common.DB.Where("email = ?", email).First(&user).Error; err != nil {
			http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
			return
		}

		// Vérification de la réponse à la question de sécurité
		if correctAnswer != securityAnswer {
			data := map[string]interface{}{
				"Email":         email,
				"Question":      question,
				"SecurityError": "La réponse de sécurité ne correspond pas",
				"PasswordError": "",
			}
			common.RenderTemplate(w, "security_questions_reset", data)
			return
		}

		// Vérification si le nouveau mot de passe est le même que l'actuel
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(newPassword)) == nil {
			data := map[string]interface{}{
				"Email":         email,
				"Question":      question,
				"SecurityError": "",
				"PasswordError": "Le nouveau mot de passe ne peut pas être le même que l'actuel",
			}
			common.RenderTemplate(w, "security_questions_reset", data)
			return
		}

		// Hachage du nouveau mot de passe
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
			return
		}

		// Mise à jour du mot de passe de l'utilisateur dans la base de données
		user.Password = string(hashedPassword)
		if err := common.DB.Save(&user).Error; err != nil {
			http.Error(w, "Erreur lors de la mise à jour du mot de passe", http.StatusInternalServerError)
			return
		}

		// Redirection vers la page de connexion après succès
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}
