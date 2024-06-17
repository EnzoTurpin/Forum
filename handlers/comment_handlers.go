package handlers

import (
	"forum/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func CreateComment(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting CreateComment handler")
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["userID"]
	if !ok {
		log.Println("User not logged in, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodPost {
		log.Println("Handling POST request for creating a comment")
		vars := mux.Vars(r)
		postIDStr := vars["id"]
		postID, err := strconv.ParseUint(postIDStr, 10, 32)
		if err != nil {
			log.Printf("Invalid post ID: %v", err)
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		r.ParseForm()
		comment := models.Comment{
			Content: r.FormValue("content"),
			PostID:  uint(postID),
			UserID:  userID.(uint),
		}
		result := db.Create(&comment)
		if result.Error != nil {
			log.Printf("Unable to create comment: %v", result.Error)
			http.Error(w, "Unable to create comment", http.StatusInternalServerError)
			return
		}
		log.Println("Comment created successfully")
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusSeeOther)
	}
}

func LikeComment(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting LikeComment handler")
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["userID"]
	if !ok {
		log.Println("User not logged in, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	vars := mux.Vars(r)
	commentID, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Printf("Invalid comment ID: %v", err)
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}
	var comment models.Comment
	if err := db.First(&comment, commentID).Error; err != nil {
		log.Printf("Comment not found: %v", err)
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}
	var like models.Like
	result := db.Where("user_id = ? AND comment_id = ?", userID, commentID).First(&like)
	if result.Error == nil {
		db.Delete(&like)
		comment.Likes--
	} else {
		like = models.Like{UserID: userID.(uint), CommentID: &comment.ID}
		db.Create(&like)
		comment.Likes++
	}
	if err := db.Save(&comment).Error; err != nil {
		log.Printf("Unable to update comment like: %v", err)
		http.Error(w, "Unable to update comment like", http.StatusInternalServerError)
		return
	}
	log.Println("Comment like updated successfully")
	postID := vars["postID"]
	http.Redirect(w, r, "/post/"+postID, http.StatusSeeOther)
}
