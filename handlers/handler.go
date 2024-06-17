package handlers

import (
	"fmt"
	"forum/models"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

var db *gorm.DB
var store *sessions.CookieStore

func SetDB(database *gorm.DB) {
	db = database
}

func SetStore(s *sessions.CookieStore) {
	store = s
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	tmpl = "./templates/" + tmpl + ".html"
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

func PageIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("Rendu de la page d'accueil")
	session, _ := store.Get(r, "session")
	user, ok := session.Values["user"]
	data := map[string]interface{}{
		"User": user,
	}
	if !ok {
		data["User"] = ""
	}

	var posts []models.Post

	// Vérifier si un filtre de catégorie est appliqué
	category := r.FormValue("category")
	if category != "" {
		if err := db.Preload("User").Preload("Comments.User").Where("category = ?", category).Find(&posts).Error; err != nil {
			log.Printf("Erreur lors de la récupération des publications par catégorie : %v", err)
			http.Error(w, "Erreur lors de la récupération des publications par catégorie", http.StatusInternalServerError)
			return
		}
		data["Category"] = category
	} else {
		if err := db.Preload("User").Preload("Comments.User").Find(&posts).Error; err != nil {
			log.Printf("Erreur lors de la récupération de toutes les publications : %v", err)
			http.Error(w, "Erreur lors de la récupération de toutes les publications", http.StatusInternalServerError)
			return
		}
	}

	data["Posts"] = posts

	renderTemplate(w, "index", data)
}

func PostsByCategory(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting PostsByCategory handler")

	vars := mux.Vars(r)
	category := vars["category"]

	var posts []models.Post
	if err := db.Preload("User").Preload("Comments.User").Where("category = ?", category).Find(&posts).Error; err != nil {
		log.Printf("Category not found: %v", err)
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	session, _ := store.Get(r, "session")
	user, ok := session.Values["user"]
	data := map[string]interface{}{
		"Posts":    posts,
		"User":     user,
		"Category": category,
	}
	if !ok {
		data["User"] = ""
	}

	renderTemplate(w, "category_posts", data)
}



func Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Register handler")

	if r.Method == http.MethodPost {
		log.Println("Handling POST request for registration")
		r.ParseForm()
		user := models.User{
			Username: r.FormValue("username"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}

		log.Printf("Parsed form values - Username: %s, Email: %s", user.Username, user.Email)

		result := db.Create(&user)
		if result.Error != nil {
			log.Printf("Unable to register user: %v", result.Error)
			http.Error(w, "Unable to register user", http.StatusInternalServerError)
			return
		}

		log.Println("User registered successfully")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		log.Println("Rendering registration template")
		renderTemplate(w, "register", nil)
	}
}

func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Login handler")

	if r.Method == http.MethodPost {
		log.Println("Handling POST request for login")
		r.ParseForm()
		var user models.User
		email := r.FormValue("email")
		password := r.FormValue("password")
		result := db.Where("email = ? AND password = ?", email, password).First(&user)
		if result.Error != nil {
			log.Printf("Invalid email or password: %v", result.Error)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}
		session, _ := store.Get(r, "session")
		session.Values["user"] = user.Username
		session.Values["userID"] = user.ID
		session.Save(r, w)
		log.Println("User logged in successfully")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		log.Println("Rendering login template")
		renderTemplate(w, "login", nil)
	}
}

func Logout(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting Logout handler")

	session, _ := store.Get(r, "session")
	delete(session.Values, "user")
	delete(session.Values, "userID")
	session.Save(r, w)
	log.Println("User logged out successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

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



func ViewProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting ViewProfile handler")

	vars := mux.Vars(r)
	username := vars["username"]

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		http.NotFound(w, r)
		return
	}

	var posts []models.Post
	db.Where("user_id = ?", user.ID).Find(&posts)

	var followers []models.Follower
	var following []models.Follower
	db.Preload("Follower").Where("follows_id = ?", user.ID).Find(&followers)
	db.Preload("Follows").Where("follower_id = ?", user.ID).Find(&following)

	data := map[string]interface{}{
		"ProfileUser":    user,
		"Posts":          posts,
		"Followers":      followers,
		"Following":      following,
		"FollowersCount": len(followers),
		"FollowingCount": len(following),
	}

	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Unable to get session: %v", err)
		http.Error(w, "Unable to get session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	currentUser, ok := session.Values["user"]
	currentUserID := session.Values["userID"]
	if ok {
		data["CurrentUser"] = currentUser
		data["CurrentUserID"] = currentUserID
	} else {
		data["CurrentUser"] = ""
		data["CurrentUserID"] = uint(0)
	}

	// Check if the current user is following the profile user
	var follower models.Follower
	if db.Where("follower_id = ? AND follows_id = ?", currentUserID, user.ID).First(&follower).Error == nil {
		data["IsFollowing"] = true
	} else {
		data["IsFollowing"] = false
	}

	renderTemplate(w, "profile", data)
}

func FollowUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting FollowUser handler")

	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Unable to get session: %v", err)
		http.Error(w, "Unable to get session", http.StatusInternalServerError)
		return
	}
	currentUserID, ok := session.Values["userID"]
	if !ok {
		log.Println("User not logged in, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		http.NotFound(w, r)
		return
	}

	// Vérifiez si l'utilisateur suit déjà
	var follower models.Follower
	if err := db.Where("follower_id = ? AND follows_id = ?", currentUserID, user.ID).First(&follower).Error; err == nil {
		log.Println("User is already following")
		http.Redirect(w, r, "/profile/"+username, http.StatusSeeOther)
		return
	}

	follower = models.Follower{
		FollowerID: currentUserID.(uint),
		FollowsID:  user.ID,
	}

	if err := db.Create(&follower).Error; err != nil {
		log.Printf("Unable to follow user: %v", err)
		http.Error(w, "Unable to follow user", http.StatusInternalServerError)
		return
	}

	log.Println("User followed successfully")
	http.Redirect(w, r, "/profile/"+username, http.StatusSeeOther)
}

func UnfollowUser(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting UnfollowUser handler")

	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Unable to get session: %v", err)
		http.Error(w, "Unable to get session", http.StatusInternalServerError)
		return
	}
	currentUserID, ok := session.Values["userID"]
	if !ok {
		log.Println("User not logged in, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		http.NotFound(w, r)
		return
	}

	// Supprimer la relation de suivi
	if err := db.Where("follower_id = ? AND follows_id = ?", currentUserID, user.ID).Delete(&models.Follower{}).Error; err != nil {
		log.Printf("Unable to unfollow user: %v", err)
		http.Error(w, "Unable to unfollow user", http.StatusInternalServerError)
		return
	}

	log.Println("User unfollowed successfully")
	http.Redirect(w, r, "/profile/"+username, http.StatusSeeOther)
}

func EditProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting EditProfile handler")

	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Unable to get session: %v", err)
		http.Error(w, "Unable to get session: "+err.Error(), http.StatusInternalServerError)
		return
	}
	currentUserID, ok := session.Values["userID"]
	if !ok {
		log.Println("User not logged in, redirecting to login page")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	username := vars["username"]

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		http.NotFound(w, r)
		return
	}

	if user.ID != currentUserID.(uint) {
		log.Println("User does not have permission to edit this profile")
		http.Error(w, "You do not have permission to edit this profile", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		log.Println("Handling POST request for profile edit")

		// Log the request headers
		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("%s: %s", name, value)
			}
		}

		// Log the raw body
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "Error reading body", http.StatusInternalServerError)
			return
		}
		log.Printf("Raw body: %s", body)

		r.Body = io.NopCloser(strings.NewReader(string(body)))
		r.ParseMultipartForm(32 << 20)

		log.Println("Request form values after parsing:")
		log.Printf("Form values: %v", r.Form)

		newUsername := r.FormValue("username")
		newEmail := r.FormValue("email")
		password := r.FormValue("password")

		log.Printf("Parsed form values - Username: %s, Email: %s, Password: %s", newUsername, newEmail, password)

		// Vérifiez l'unicité de l'adresse e-mail pour les autres utilisateurs
		var existingUser models.User
		if err := db.Where("email = ? AND id != ?", newEmail, user.ID).First(&existingUser).Error; err == nil {
			log.Println("Email already in use by another user")
			http.Error(w, "Email already in use", http.StatusBadRequest)
			return
		}

		user.Username = newUsername
		user.Email = newEmail
		if password != "" {
			user.Password = password // Dans une application réelle, hachez le mot de passe avant de le stocker
		}

		log.Println("Checking for profile picture upload")
		// Gérer l'upload de la photo de profil
		file, header, err := r.FormFile("profile_picture")
		if err != nil && err != http.ErrMissingFile {
			log.Printf("Error uploading file: %v", err)
			http.Error(w, "Error uploading file: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if err == nil {
			defer file.Close()
			log.Printf("Uploaded file: %s", header.Filename)
			fileExtension := strings.ToLower(header.Filename[strings.LastIndex(header.Filename, "."):])
			filePath := fmt.Sprintf("static/uploads/%d%s", user.ID, fileExtension)
			f, err := os.Create(filePath)
			if err != nil {
				log.Printf("Error creating file: %v", err)
				http.Error(w, "Error creating file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer f.Close()
			_, err = io.Copy(f, file)
			if err != nil {
				log.Printf("Error saving file: %v", err)
				http.Error(w, "Error saving file: "+err.Error(), http.StatusInternalServerError)
				return
			}
			user.ProfilePicture = filePath
			log.Printf("Profile picture saved to: %s", filePath)
		}

		if err := db.Save(&user).Error; err != nil {
			log.Printf("Unable to update profile: %v", err)
			http.Error(w, "Unable to update profile: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Mettre à jour la session avec le nouveau nom d'utilisateur
		session.Values["user"] = user.Username
		session.Save(r, w)
		log.Println("Profile updated successfully and session saved")

		http.Redirect(w, r, fmt.Sprintf("/profile/%s", user.Username), http.StatusSeeOther)
		return
	}

	log.Println("Handling GET request for profile edit")

	data := map[string]interface{}{
		"User": user,
	}

	renderTemplate(w, "edit_profile", data)
}

func ViewFollowers(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting ViewFollowers handler")

	vars := mux.Vars(r)
	username := vars["username"]

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		http.NotFound(w, r)
		return
	}

	var followers []models.Follower
	db.Preload("Follower").Where("follows_id = ?", user.ID).Find(&followers)

	data := map[string]interface{}{
		"ProfileUser": user,
		"Followers":   followers,
	}

	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Unable to get session: %v", err)
		http.Error(w, "Unable to get session", http.StatusInternalServerError)
		return
	}
	currentUser, ok := session.Values["user"]
	currentUserID := session.Values["userID"]
	if ok {
		data["CurrentUser"] = currentUser
		data["CurrentUserID"] = currentUserID
	} else {
		data["CurrentUser"] = ""
		data["CurrentUserID"] = uint(0)
	}

	renderTemplate(w, "followers", data)
}

func ViewFollowing(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting ViewFollowing handler")

	vars := mux.Vars(r)
	username := vars["username"]

	var user models.User
	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		log.Printf("User not found: %v", err)
		http.NotFound(w, r)
		return
	}

	var following []models.Follower
	db.Preload("Follows").Where("follower_id = ?", user.ID).Find(&following)

	data := map[string]interface{}{
		"ProfileUser": user,
		"Following":   following,
	}

	session, err := store.Get(r, "session")
	if err != nil {
		log.Printf("Unable to get session: %v", err)
		http.Error(w, "Unable to get session", http.StatusInternalServerError)
		return
	}
	currentUser, ok := session.Values["user"]
	currentUserID := session.Values["userID"]
	if ok {
		data["CurrentUser"] = currentUser
		data["CurrentUserID"] = currentUserID
	} else {
		data["CurrentUser"] = ""
		data["CurrentUserID"] = uint(0)
	}

	renderTemplate(w, "following", data)
}
