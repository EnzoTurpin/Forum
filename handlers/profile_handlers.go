package handlers

import (
	"fmt"
	"forum/models"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
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
		for name, values := range r.Header {
			for _, value := range values {
				log.Printf("%s: %s", name, value)
			}
		}
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
		var existingUser models.User
		if err := db.Where("email = ? AND id != ?", newEmail, user.ID).First(&existingUser).Error; err == nil {
			log.Println("Email already in use by another user")
			http.Error(w, "Email already in use", http.StatusBadRequest)
			return
		}
		user.Username = newUsername
		user.Email = newEmail
		if password != "" {
			user.Password = password
		}
		log.Println("Checking for profile picture upload")
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
