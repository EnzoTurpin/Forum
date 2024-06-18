package main

import (
	"fmt"
	"forum/handlers"
	"forum/models"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

func main() {
	// Custom logger with minimal output
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Silent, // Set log level to Silent
			Colorful:      false,
		},
	)

	// Open the database with custom parameters
	db, err := gorm.Open(sqlite.Open("file:bdd/forum.db?cache=shared&_timeout=5000"), &gorm.Config{
		Logger:      newLogger,
		PrepareStmt: true, // Use prepared statements
	})

	if err != nil {
		fmt.Println("Échec de la connexion à la base de données :", err)
		return
	}

	// Set the connection pool parameters
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Explicitly set the journal mode to DELETE
	if _, err := sqlDB.Exec("PRAGMA journal_mode = DELETE;"); err != nil {
		fmt.Println("Échec de la configuration du mode journal DELETE :", err)
		return
	}

	// Migrer les modèles
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Follower{}, &models.Category{}); err != nil {
		log.Fatalf("Échec de la migration des modèles : %v", err)
	}
	handlers.SetDB(db)
	handlers.SetStore(store)

	// Ajouter des catégories par défaut si elles n'existent pas déjà
	categories := []string{"Action", "Aventure", "RPG", "FPS", "TPS", "Stratégie", "Simulation", "Sport", "Course", "Puzzle", "Combat", "Plateforme", "Horreur", "MMO", "VR", "Jeux de rythme", "Party Games", "Rogue-like", "Metroidvania", "Sandbox", "Visual Novel", "Jeux de cartes", "Jeux de société", "Jeux de gestion", "Survival"}
	for _, name := range categories {
		var category models.Category
		if err := db.Where("name = ?", name).First(&category).Error; err != nil {
			db.Create(&models.Category{Name: name})
		}
	}

	// Initialiser le routeur
	r := mux.NewRouter()

	// Routes pour les handlers
	r.HandleFunc("/", handlers.PageIndex).Methods("GET")
	r.HandleFunc("/register", handlers.Register).Methods("GET", "POST")
	r.HandleFunc("/login", handlers.Login).Methods("GET", "POST")
	r.HandleFunc("/logout", handlers.Logout).Methods("GET")
	r.HandleFunc("/create-post", handlers.CreatePost).Methods("GET", "POST")
	r.HandleFunc("/post/{id}", handlers.ViewPost).Methods("GET")
	r.HandleFunc("/post/{id}/comment", handlers.CreateComment).Methods("POST")
	r.HandleFunc("/post/{id}/like", handlers.LikePost).Methods("POST")
	r.HandleFunc("/post/{id}", handlers.ViewPostWithComments).Methods("GET")
	r.HandleFunc("/post/{postID}/comments", handlers.ViewAllComments).Methods("GET")
	r.HandleFunc("/post/{postID}/comment/{id}/like", handlers.LikeComment).Methods("POST")
	r.HandleFunc("/category/{category}", handlers.PostsByCategory).Methods("GET")
	r.HandleFunc("/profile/{username}", handlers.ViewProfile).Methods("GET")
	r.HandleFunc("/profile/{username}/follow", handlers.FollowUser).Methods("POST")
	r.HandleFunc("/profile/{username}/unfollow", handlers.UnfollowUser).Methods("POST")
	r.HandleFunc("/profile/{username}/edit", handlers.EditProfile).Methods("GET", "POST")
	r.HandleFunc("/profile/{username}/followers", handlers.ViewFollowers).Methods("GET")
	r.HandleFunc("/profile/{username}/following", handlers.ViewFollowing).Methods("GET")
	r.HandleFunc("/profile/{username}/delete", handlers.DeleteProfile).Methods("POST")
	r.HandleFunc("/categories", handlers.Categories).Methods("GET")
	r.HandleFunc("/post/{id}/edit", handlers.ShowEditPostForm).Methods("GET") // Nouvelle route pour afficher le formulaire d'édition
	r.HandleFunc("/edit-post/{id}", handlers.EditPost).Methods("POST") // Nouvelle route pour modifier le post
	r.HandleFunc("/delete-post/{id}", handlers.DeletePost).Methods("POST") // Nouvelle route pour supprimer le post

	// Servir les fichiers statiques (HTML, CSS, JS)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// Démarrer le serveur sur le port 8080
	fmt.Println("Le serveur est en cours d'exécution sur http://localhost:8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Println("Échec du démarrage du serveur :", err)
	}
}
