package common

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"time"

	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

// Variables globales pour la base de données et le store de sessions
var (
	DB    *gorm.DB
	Store *sessions.CookieStore
)

// SetDB initialise la variable globale de la base de données
func SetDB(database *gorm.DB) {
	DB = database
}

// SetStore initialise la variable globale du store de sessions
func SetStore(store *sessions.CookieStore) {
	Store = store
}

// RenderTemplate rend une template HTML avec les données fournies
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	// Construire le chemin complet du fichier de template
	tmpl = "./templates/" + tmpl + ".html"
	// Parser le fichier de template
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		// En cas d'erreur, retourner une réponse HTTP 500
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Exécuter le template avec les données fournies
	t.Execute(w, data)
}

// FormatTimeAgo retourne une chaîne de caractères représentant le temps écoulé depuis un moment donné
func FormatTimeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d.Seconds() < 60:
		return fmt.Sprintf("il y a %.f secondes", d.Seconds())
	case d.Minutes() < 60:
		return fmt.Sprintf("il y a %.f minutes", d.Minutes())
	case d.Hours() < 24:
		return fmt.Sprintf("il y a %.f heures", d.Hours())
	case d.Hours() < 48:
		return "il y a 1 jour"
	default:
		return fmt.Sprintf("il y a %.f jours", d.Hours()/24)
	}
}

// ValidatePassword valide un mot de passe selon plusieurs critères
func ValidatePassword(password string) error {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	// Vérifier si le mot de passe a au moins 8 caractères
	if len(password) >= 8 {
		hasMinLen = true
	}
	// Vérifier les autres critères de complexité du mot de passe
	for _, char := range password {
		switch {
		case regexp.MustCompile(`[A-Z]`).MatchString(string(char)):
			hasUpper = true
		case regexp.MustCompile(`[a-z]`).MatchString(string(char)):
			hasLower = true
		case regexp.MustCompile(`[0-9]`).MatchString(string(char)):
			hasNumber = true
		case regexp.MustCompile(`[!@#\$%\^&\*]`).MatchString(string(char)):
			hasSpecial = true
		}
	}
	// Si le mot de passe ne respecte pas tous les critères, retourner une erreur
	if !hasMinLen || !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("le mot de passe doit comporter au moins 8 caractères et inclure au moins une lettre majuscule, une lettre minuscule, un chiffre et un caractère spécial")
	}
	return nil
}
