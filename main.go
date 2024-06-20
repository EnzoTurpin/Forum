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

// Définition des constantes pour la coloration du texte dans le terminal
const (
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorRed    = "\033[31m"
	colorReset  = "\033[0m"
)

func main() {
	// Configuration du logger pour Gorm
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,   // Seuil de lenteur pour les requêtes
			LogLevel:      logger.Silent, // Niveau de log : silencieux
			Colorful:      false,         // Pas de couleur dans les logs
		},
	)

	// Connexion à la base de données SQLite
	db, err := gorm.Open(sqlite.Open("data/forum.db?cache=shared&_timeout=5000"), &gorm.Config{
		Logger:      newLogger, // Utilisation du logger configuré
		PrepareStmt: true,      // Préparation des requêtes
	})
	if err != nil {
		fmt.Println("Échec de la connexion à la base de données :", err)
		return
	}

	// Configuration des connexions de la base de données
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Configuration du mode journal de SQLite
	if _, err := sqlDB.Exec("PRAGMA journal_mode = DELETE;"); err != nil {
		fmt.Println("Échec de la configuration du mode journal DELETE :", err)
		return
	}

	// Migration des modèles vers la base de données
	if err := db.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{}, &models.Like{}, &models.Follower{}, &models.Category{}); err != nil {
		log.Fatalf("Échec de la migration des modèles : %v", err)
	}

	// Configuration de la base de données et du store de sessions dans le package commun
	common.SetDB(db)
	common.SetStore(sessions.NewCookieStore([]byte("secret-key")))

	// Initialisation des catégories prédéfinies dans la base de données
	categories := []string{"Action", "Aventure", "RPG", "FPS", "TPS", "Stratégie", "Simulation", "Sport", "Course", "Puzzle", "Combat", "Plateforme", "Horreur", "MMO", "VR", "Jeux de rythme", "Party Games", "Rogue-like", "Metroidvania", "Sandbox", "Visual Novel", "Jeux de cartes", "Jeux de société", "Jeux de gestion", "Survival"}
	for _, name := range categories {
		var category models.Category
		if err := db.Where("name = ?", name).First(&category).Error; err != nil {
			db.Create(&models.Category{Name: name})
		}
	}

	// Initialisation du routeur
	r := mux.NewRouter()

	// Définition des routes et de leurs gestionnaires
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

	// Configuration du serveur de fichiers statiques
	staticDir := "./ressources/static/"
	r.PathPrefix("/ressources/static/").Handler(http.StripPrefix("/ressources/static/", http.FileServer(http.Dir(staticDir))))

	// Affichage des messages indiquant que le serveur est prêt et comment l'arrêter
	fmt.Printf("%s[SERVER_READY] Serveur démarré sur: http://localhost:8080%s\n", colorGreen, colorReset)
	fmt.Printf("%s[SERVER_INFO] Appuyez sur Ctrl+C pour arrêter le serveur.%s\n", colorYellow, colorReset)

	// Configuration et démarrage du serveur HTTP
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Échec du démarrage du serveur : %v", err)
		}
	}()

	// Gestion des signaux pour arrêter le serveur proprement
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Printf("%s[SERVER_STOP] Le serveur a été arrêté proprement.%s\n", colorRed, colorReset)

	// Fermeture du serveur HTTP
	if err := srv.Close(); err != nil {
		log.Fatal("Server Close:", err)
	}
}
