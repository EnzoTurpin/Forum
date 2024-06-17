package handlers

import (
	"forum/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting CreatePost handler")
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["userID"]
	if !ok {
		log.Println("User not logged in, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodPost {
		log.Println("Handling POST request for creating a post")
		r.ParseForm()
		post := models.Post{
			Title:    r.FormValue("title"),
			Content:  r.FormValue("content"),
			Category: r.FormValue("category"),
			UserID:   userID.(uint),
		}
		result := db.Create(&post)
		if result.Error != nil {
			log.Printf("Unable to create post: %v", result.Error)
			http.Error(w, "Unable to create post", http.StatusInternalServerError)
			return
		}
		log.Println("Post created successfully")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		log.Println("Rendering create post template")
		renderTemplate(w, "create_post", nil)
	}
}

func ViewPost(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting ViewPost handler")
	vars := mux.Vars(r)
	var post models.Post
	if err := db.Preload("User").Preload("Comments.User").First(&post, vars["id"]).Error; err != nil {
		log.Printf("Post not found: %v", err)
		http.NotFound(w, r)
		return
	}
	session, _ := store.Get(r, "session")
	user, ok := session.Values["user"]
	data := map[string]interface{}{
		"Post": post,
		"User": user,
	}
	if !ok {
		data["User"] = ""
	}
	renderTemplate(w, "view_post", data)
}

func LikePost(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting LikePost handler")
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["userID"]
	if !ok {
		log.Println("User not logged in, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	vars := mux.Vars(r)
	postID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	var post models.Post
	if err := db.First(&post, postID).Error; err != nil {
		log.Printf("Post not found: %v", err)
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	var like models.Like
	result := db.Where("user_id = ? AND post_id = ?", userID, postID).First(&like)
	if result.Error == nil {
		db.Delete(&like)
		post.Likes--
	} else {
		like = models.Like{UserID: userID.(uint), PostID: &post.ID}
		db.Create(&like)
		post.Likes++
	}
	if err := db.Save(&post).Error; err != nil {
		log.Printf("Unable to update post like: %v", err)
		http.Error(w, "Unable to update post like", http.StatusInternalServerError)
		return
	}
	log.Println("Post like updated successfully")
	http.Redirect(w, r, "/post/"+vars["id"], http.StatusSeeOther)
}
