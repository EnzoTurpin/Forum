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

// Déclaration des variables globales pour la base de données et le store de sessions
var (
	DB    *gorm.DB
	Store *sessions.CookieStore
)

// SetDB configure la base de données globale
func SetDB(database *gorm.DB) {
	DB = database
}

// SetStore configure le store de sessions global
func SetStore(store *sessions.CookieStore) {
	Store = store
}

// RenderTemplate rend un template HTML avec les données fournies
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmpl = "./ressources/templates/" + tmpl + ".html"
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		// Affiche une erreur HTTP interne en cas d'erreur de parsing du template
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

// FormatTimeAgo formate un temps donné en une chaîne de caractères indiquant le temps écoulé depuis
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

// ValidatePassword valide un mot de passe selon des critères de sécurité
func ValidatePassword(password string) error {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	// Vérifie la longueur minimale du mot de passe
	if len(password) >= 8 {
		hasMinLen = true
	}

	// Vérifie la présence de différents types de caractères
	for _, char := range password {
		switch {
		case regexp.MustCompile(`[A-Z]`).MatchString(string(char)):
			hasUpper = true
		case regexp.MustCompile(`[a-z]`).MatchString(string(char)):
			hasLower = true
		case regexp.MustCompile(`[0-9]`).MatchString(string(char)):
			hasNumber = true
		case regexp.MustCompile(`[!\@\#\$\%\^\&\*\(\)\-\_\=\+\{\}\[\]\|\\:;'\"<>,\.?/]`).MatchString(string(char)):
			hasSpecial = true
		}
	}

	// Si le mot de passe ne remplit pas tous les critères, retourne une erreur
	if !hasMinLen || !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return errors.New("le mot de passe doit comporter au moins 8 caractères et inclure au moins une lettre majuscule, une lettre minuscule, un chiffre et un caractère spécial")
	}

	// Si tous les critères sont remplis, retourne nil (pas d'erreur)
	return nil
}
