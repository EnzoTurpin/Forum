package models

import "gorm.io/gorm"

type Post struct {
	gorm.Model
	Title      string
	Content    string
	UserID     uint
	User       User
	Comments   []Comment
	CategoryID uint
	Category   Category
	Likes      int
	Dislikes   int
	TimeAgo    string `gorm:"-"`
}

type Comment struct {
	gorm.Model
	Content  string
	PostID   uint
	UserID   uint
	User     User
	Likes    int
	Dislikes int
}

type User struct {
	gorm.Model
	Username       string `gorm:"unique;not null"`
	Email          string `gorm:"unique;not null"`
	Password       string `gorm:"not null"`
	ProfilePicture string
	Followers      []Follower `gorm:"foreignKey:FollowsID"`
	Following      []Follower `gorm:"foreignKey:FollowerID"`
}

type Like struct {
	gorm.Model
	UserID    uint
	PostID    *uint
	CommentID *uint
}

type Follower struct {
	gorm.Model
	FollowerID uint
	FollowsID  uint
	Follower   User `gorm:"foreignKey:FollowerID"`
	Follows    User `gorm:"foreignKey:FollowsID"`
}

type Category struct {
	gorm.Model
	Name  string `gorm:"unique;not null"`
	Posts []Post `gorm:"foreignKey:CategoryID"`
}
