package models

import "gorm.io/gorm"

// Post représente une publication dans le forum
type Post struct {
	gorm.Model
	Title      string
	Content    string
	UserID     uint
	User       User
	Comments   []Comment
	Categories []Category `gorm:"many2many:post_categories;"`
	Likes      int
	Dislikes   int
	TimeAgo    string `gorm:"-"`
}

// Comment représente un commentaire sur une publication
type Comment struct {
	gorm.Model
	Content  string
	PostID   uint
	UserID   uint
	User     User
	Likes    int
	Dislikes int
}

// User représente un utilisateur du forum
type User struct {
	gorm.Model
	Username          string `gorm:"unique;not null"`
	Email             string `gorm:"unique;not null"`
	Password          string `gorm:"not null"`
	ProfilePicture    string
	SecurityQuestion1 string
	SecurityAnswer1   string
	SecurityQuestion2 string
	SecurityAnswer2   string
	SecurityQuestion3 string
	SecurityAnswer3   string
	Followers         []Follower     `gorm:"foreignKey:FollowsID"`
	Following         []Follower     `gorm:"foreignKey:FollowerID"`
	DeletedAt         gorm.DeletedAt `gorm:"-"`
}

// Like représente un like sur une publication ou un commentaire
type Like struct {
	gorm.Model
	UserID    uint
	PostID    *uint
	CommentID *uint
}

// Follower représente le lien de suivi entre deux utilisateurs
type Follower struct {
	gorm.Model
	FollowerID uint
	FollowsID  uint
	Follower   User `gorm:"foreignKey:FollowerID"`
	Follows    User `gorm:"foreignKey:FollowsID"`
}

// Category représente une catégorie de publication dans le forum
type Category struct {
	gorm.Model
	Name  string `gorm:"unique;not null"`
	Posts []Post `gorm:"many2many:post_categories;"`
}
