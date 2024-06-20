package general

import (
	"forum/src/common"
	"forum/src/models"
	"math/rand"
	"net/http"
	"time"
)

// ForgotPassword gère la demande de réinitialisation de mot de passe
func ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Récupération de l'email du formulaire
		email := r.FormValue("email")

		// Recherche de l'utilisateur par email dans la base de données
		var user models.User
		if err := common.DB.Where("email = ?", email).First(&user).Error; err != nil {
			http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
			return
		}

		// Récupération des questions et réponses de sécurité de l'utilisateur
		questions := []string{
			user.SecurityQuestion1,
			user.SecurityQuestion2,
			user.SecurityQuestion3,
		}
		answers := []string{
			user.SecurityAnswer1,
			user.SecurityAnswer2,
			user.SecurityAnswer3,
		}

		// Sélection aléatoire d'une question de sécurité
		rand.Seed(time.Now().UnixNano())
		idx := rand.Intn(len(questions))

		// Enregistrement de l'email et de la question/réponse de sécurité dans la session
		session, _ := common.Store.Get(r, "session")
		session.Values["reset_email"] = email
		session.Values["security_question"] = questions[idx]
		session.Values["security_answer"] = answers[idx]
		session.Save(r, w)

		// Préparation des données à passer au template
		data := map[string]interface{}{
			"Email":    email,
			"Question": questions[idx],
		}
		// Rendu du template pour la vérification de la question de sécurité
		common.RenderTemplate(w, "security_questions_reset", data)
		return
	}

	// Rendu du template pour le formulaire de demande de réinitialisation de mot de passe
	common.RenderTemplate(w, "forgot_password", nil)
}
