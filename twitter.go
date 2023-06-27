package main

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	DB *gorm.DB
}

func (app *App) createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON provided."})

		fmt.Println(err)
		return
	}

	app.DB.Create(&user)

	c.JSON(200, user)
}

func (app *App) createTweet(c *gin.Context) {
	var tweet Tweet
	if err := c.ShouldBindJSON(&tweet); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON provided."})
		return
	}

	u64, err := strconv.ParseUint(c.Request.Header.Get("userID"), 10, 32)
	fmt.Println(c.Request.Header.Get("userID"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid userID provided in header."})
		return
	}

	tweet.UserID = uint(u64)

	app.DB.Create(&tweet)

	c.JSON(200, tweet)
}

func (app *App) followUser(c *gin.Context) {
	var friend Friend
	if err := c.ShouldBindJSON(&friend); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON provided."})
		return
	}

	app.DB.Create(&friend)

	c.JSON(200, friend)
}

func (app *App) like(c *gin.Context) {
	var like Like
	if err := c.ShouldBindJSON(&like); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON provided."})
		return
	}
	u64, err := strconv.ParseUint(c.Request.Header.Get("userID"), 10, 32)
	fmt.Println("userID:")
	fmt.Println(c.Request.Header.Get("userID"))
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid userID provided in header."})
		return
	}

	like.UserID = uint(u64)
	app.DB.Create(&like)

	c.JSON(200, like)
}

func (app *App) getTweet(c *gin.Context) {
	var tweet Tweet
	all := c.GetHeader("Token")
	fmt.Println("all:")
	fmt.Println(all)
	if err := app.DB.Preload("Likes").First(&tweet, c.Param("id")).Error; err != nil {
		c.JSON(400, gin.H{"error": "Record not found."})
		return
	}

	c.JSON(200, tweet)
}

func (app *App) getUserTweets(c *gin.Context) {
	var user User
	if err := app.DB.Preload("Tweets").Where("username = ?", c.Param("username")).First(&user).Error; err != nil {
		c.JSON(400, gin.H{"error": "Record not found."})
		return
	}

	c.JSON(200, user.Tweets)
}

func (app *App) getUserFollowings(c *gin.Context) {
	var user User
	if err := app.DB.Preload("Friends").Where("username = ?", c.Param("username")).First(&user).Error; err != nil {
		c.JSON(400, gin.H{"error": "Record not found."})
		return
	}

	c.JSON(200, user.Friends)
}
