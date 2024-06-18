package handlers

import (
	"html/template"
	"net/http"
)

func Page_index(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		Index(w, "template/index.html")
		return
	}
	http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
}

func Index(w http.ResponseWriter, tmpl string) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, nil)
}
