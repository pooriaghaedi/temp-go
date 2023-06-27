package main

import (
	"gorm.io/gorm"
)

//	type User struct {
//		gorm.Model
//		Username string   `gorm:"type:varchar(100);uniqueIndex" json:"Username"`
//		Password string   `gorm:"type:varchar(100)" json:"Password"`
//		Tweets   []Tweet  `json:"Tweets"`
//		Friends  []Friend `json:"Friend"`
//	}
type User struct {
	gorm.Model
	Email     string   `json:"email"`
	Verified  bool     `json:"email_verified"`
	Name      string   `json:"name"`
	Password  string   `gorm:"type:varchar(100)" json:"Password"`
	GivenName string   `json:"given_name"`
	Username  string   `gorm:"type:varchar(100);uniqueIndex" json:"Username"`
	Nickname  string   `json:"nickname"`
	Groups    []string `gorm:"type:text[]" json:"groups"`
	Tweets    []Tweet  `gorm:"foreignKey:UserID" json:"Tweets"`
	Friends   []Friend `gorm:"foreignKey:UserID" json:"Friend"`
}

type Tweet struct {
	gorm.Model
	UserID uint   `json:"UserID"`
	Text   string `gorm:"type:varchar(280)" json:"Text"`
	Likes  []Like `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"Likes"`
}

type Like struct {
	gorm.Model
	UserID  uint `json:"UserID"`
	TweetID uint `json:"TweetID"`
}

type Friend struct {
	gorm.Model
	UserID       uint `json:"UserID"`
	FriendUserID uint `json:"FriendUserID"`
}
