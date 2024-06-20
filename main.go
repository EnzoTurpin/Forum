package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"forum/src/common"
	"forum/src/handlers/auth"
	"forum/src/handlers/general"
	"forum/src/handlers/post"
	"forum/src/handlers/profile"
	"forum/src/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorReset  = "\033[0m"
)

func main() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent,
			Colorful:      false,
		},
	)

	db, err := gorm.Open(sqlite.Open("data/forum.db?cache=shared&_timeout=5000"), &gorm.Config{
		Logger:      newLogger,
		PrepareStmt: true,
	})
	if err != nil {
		fmt.Println("Échec de la connexion à la base de données :", err)
		return
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if _, err := sqlDB.Exec("PRAGMA journal_mode = DELETE;"); err != nil {
		fmt.Println("Échec de la configuration du mode journal DELETE :", err)
		return
	}

	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Follower{}, &models.Category{}); err != nil {
		log.Fatalf("Échec de la migration des modèles : %v", err)
	}

	common.SetDB(db)
	common.SetStore(sessions.NewCookieStore([]byte("secret-key")))

	categories := []string{"Action", "Aventure", "RPG", "FPS", "TPS", "Stratégie", "Simulation", "Sport", "Course", "Puzzle", "Combat", "Plateforme", "Horreur", "MMO", "VR", "Jeux de rythme", "Party Games", "Rogue-like", "Metroidvania", "Sandbox", "Visual Novel", "Jeux de cartes", "Jeux de société", "Jeux de gestion", "Survival"}
	for _, name := range categories {
		var category models.Category
		if err := db.Where("name = ?", name).First(&category).Error; err != nil {
			db.Create(&models.Category{Name: name})
		}
	}

	r := mux.NewRouter()

	r.HandleFunc("/", general.PageIndex).Methods("GET")
	r.HandleFunc("/register", auth.Register).Methods("GET", "POST")
	r.HandleFunc("/login", auth.Login).Methods("GET", "POST")
	r.HandleFunc("/logout", auth.Logout).Methods("GET")
	r.HandleFunc("/create-post", post.CreatePost).Methods("GET", "POST")
	r.HandleFunc("/post/{id}", post.ViewPost).Methods("GET")
	r.HandleFunc("/post/{id}/comment", general.CreateComment).Methods("POST")
	r.HandleFunc("/post/{id}/like", post.LikePost).Methods("POST")
	r.HandleFunc("/post/{id}/view-with-comments", post.ViewPostWithComments).Methods("GET")
	r.HandleFunc("/post/{postID}/comments", general.ViewAllComments).Methods("GET")
	r.HandleFunc("/post/{postID}/comment/{id}/like", general.LikeComment).Methods("POST")
	r.HandleFunc("/category/{category}", general.PostsByCategory).Methods("GET")
	r.HandleFunc("/profile/{username}", profile.ViewProfile).Methods("GET")
	r.HandleFunc("/profile/{username}/follow", profile.FollowUser).Methods("POST")
	r.HandleFunc("/profile/{username}/unfollow", profile.UnfollowUser).Methods("POST")
	r.HandleFunc("/profile/{username}/edit", profile.EditProfile).Methods("GET", "POST")
	r.HandleFunc("/profile/{username}/followers", profile.ViewFollowers).Methods("GET")
	r.HandleFunc("/profile/{username}/following", profile.ViewFollowing).Methods("GET")
	r.HandleFunc("/profile/{username}/delete", profile.DeleteProfile).Methods("POST")
	r.HandleFunc("/categories", general.Categories).Methods("GET")
	r.HandleFunc("/post/{id}/edit", post.ShowEditPostForm).Methods("GET")
	r.HandleFunc("/edit-post/{id}", post.EditPost).Methods("POST")
	r.HandleFunc("/delete-post/{id}", post.DeletePost).Methods("POST")
	r.HandleFunc("/forgot-password", general.ForgotPassword).Methods("GET", "POST")
	r.HandleFunc("/reset-password", general.ResetPassword).Methods("POST")
	r.HandleFunc("/register-step1", auth.RegisterStep1).Methods("POST")
	r.HandleFunc("/register-step2", auth.RegisterStep2).Methods("POST")

	staticDir := "./ressources/static/"
	r.PathPrefix("/ressources/static/").Handler(http.StripPrefix("/ressources/static/", http.FileServer(http.Dir(staticDir))))

	profilePicturesDir := "./ressources/static/profile_pictures/"
	r.PathPrefix("/ressources/static/profile_pictures/").Handler(http.StripPrefix("/ressources/static/profile_pictures/", http.FileServer(http.Dir(profilePicturesDir))))

	fmt.Printf("%s[SERVER_READY] Serveur démarré sur: http://localhost:8080%s\n", colorGreen, colorReset)
	fmt.Printf("%s[SERVER_INFO] Appuyez sur Ctrl+C pour arrêter le serveur.%s\n", colorYellow, colorReset)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Échec du démarrage du serveur : %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Printf("%s[SERVER_STOP] Le serveur a été arrêté proprement.%s\n", colorRed, colorReset)

	if err := srv.Close(); err != nil {
		log.Fatal("Server Close:", err)
	}
}
