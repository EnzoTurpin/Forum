package handlers

import (
	"fmt"
	"forum/models"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

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

	// Format the time ago for each post
	for i := range posts {
		posts[i].TimeAgo = formatTimeAgo(posts[i].CreatedAt)
	}

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
		r.Body = http.MaxBytesReader(w, r.Body, 1048576)
		r.ParseMultipartForm(32 << 20)

		newUsername := r.FormValue("username")
		newEmail := r.FormValue("email")
		password := r.FormValue("password")
		log.Printf("Parsed form values - Username: %s, Email: %s, Password: %s", newUsername, newEmail, password)

		// Check if the new username already exists
		var existingUser models.User
		if err := db.Where("username = ? AND id != ?", newUsername, user.ID).First(&existingUser).Error; err == nil {
			log.Println("Username already in use")
			data := map[string]interface{}{
				"User":          user,
				"UsernameError": "Username already in use",
			}
			renderTemplate(w, "edit_profile", data)
			return
		}

		// Validate new email if changed
		if err := db.Where("email = ? AND id != ?", newEmail, user.ID).First(&existingUser).Error; err == nil {
			log.Println("Email already in use by another user")
			data := map[string]interface{}{
				"User":       user,
				"EmailError": "Email already in use",
			}
			renderTemplate(w, "edit_profile", data)
			return
		}

		user.Username = newUsername
		user.Email = newEmail

		// Validate and hash new password if provided
		if password != "" {
			if err := validatePassword(password); err != nil {
				log.Printf("Password validation failed: %v", err)
				data := map[string]interface{}{
					"User":          user,
					"PasswordError": err.Error(),
				}
				renderTemplate(w, "edit_profile", data)
				return
			}
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("Error hashing password: %v", err)
				http.Error(w, "Error hashing password", http.StatusInternalServerError)
				return
			}
			user.Password = string(hashedPassword)
		}

		if err := db.Save(&user).Error; err != nil {
			log.Printf("Unable to update profile: %v", err)
			http.Error(w, "Unable to update profile: "+err.Error(), http.StatusInternalServerError)
			return
		}
		session.Values["user"] = user.Username
		session.Save(r, w)
		log.Println("Profile updated successfully and session saved")
		http.Redirect(w, r, fmt.Sprintf("/profile/%s", user.Username), http.StatusSeeOther)
		return
	}
	log.Println("Rendering edit profile template")
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

func DeleteProfile(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting DeleteProfile handler")
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
		log.Println("User does not have permission to delete this profile")
		http.Error(w, "You do not have permission to delete this profile", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodPost {
		log.Println("Handling POST request for profile deletion")

		// Delete user's posts
		if err := db.Where("user_id = ?", user.ID).Delete(&models.Post{}).Error; err != nil {
			log.Printf("Error deleting user's posts: %v", err)
			http.Error(w, "Error deleting user's posts", http.StatusInternalServerError)
			return
		}

		// Delete the user
		if err := db.Delete(&user).Error; err != nil {
			log.Printf("Error deleting user: %v", err)
			http.Error(w, "Error deleting user", http.StatusInternalServerError)
			return
		}

		// Clear the session
		delete(session.Values, "user")
		delete(session.Values, "userID")
		if err := session.Save(r, w); err != nil {
			log.Printf("Error saving session: %v", err)
			http.Error(w, "Error saving session", http.StatusInternalServerError)
			return
		}

		log.Println("User profile and posts deleted successfully")
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
