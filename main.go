package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"golang.org/x/oauth2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ctx = context.Background()

	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("AUTHENTIK_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTHENTIK_CLIENT_SECRET"),
		Scopes:       []string{"openid", "email", "profile"},
		RedirectURL:  os.Getenv("REDIRECT_URL"),
		Endpoint: oauth2.Endpoint{
			AuthURL: os.Getenv("AUTH_URL"),
			TokenURL: os.Getenv("AUTH_TOKEN"),
		},
	}
)

func main() {
	router := gin.Default()
	InitDB()

	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
	app := &App{
		DB: db, // assuming db is your initialized *gorm.DB connection
	}
	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	InitProviderWithJaegerExporter(ctx)
	if err != nil {
		log.Fatalf("%s: %v", "Failed to initialize opentelemetry provider", err)
	}
	// defer shutdown(ctx)

	// router.Use(UserIDMiddleware(, &sql.DB{}))
	router.Use(otelgin.Middleware("go-twitter"))
	router.Use(UserIDMiddleware(getAccessToken, sqlDB))
	router.GET("/", handleMain)
	router.GET("/login", handleLogin)
	router.GET("/callback", handleCallback)
	router.GET("/querydb", queryDB)
	// router.GET("/protected", protectedRoute)

	router.POST("/user", app.createUser)
	router.POST("/tweet", app.createTweet)
	router.POST("/like", app.like)
	router.POST("/follow", app.followUser)

	router.GET("/tweet/:id", app.getTweet)
	router.GET("/user/:username/tweets", app.getUserTweets)
	router.GET("/user/:username/followings", app.getUserFollowings)
	router.Run(":8888")

}

func getAccessToken(c *gin.Context) string {
	accessToken := c.GetHeader("Authorization")
	return accessToken
}

func UserIDMiddleware(getAccessToken func(*gin.Context) string, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		if strings.HasPrefix(c.Request.URL.Path, "/login") ||
			strings.HasPrefix(c.Request.URL.Path, "/callback") ||
			strings.HasPrefix(c.Request.URL.Path, "/open") {
			// Skip the middleware and call the next handler directly
			c.Next()
			return
		}

		accessToken := getAccessToken(c)

		client := &http.Client{}
		req, _ := http.NewRequest("GET", "http://auth.local/application/o/userinfo/", nil)
		req.Header.Add("Authorization", "Bearer "+accessToken)
		resp, err := client.Do(req)

		if err != nil {
			log.Fatal(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user information"})
			return
		}
		defer resp.Body.Close()

		if c.Writer.Status() == http.StatusUnauthorized {
			// Redirect to the login page
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			log.Fatal(fmt.Printf("User info retrieval failed with status: %d", resp.StatusCode))
			c.AbortWithStatusJSON(resp.StatusCode, gin.H{"error": "Failed to retrieve user information"})
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user information"})
			return
		}

		var userInfo User
		if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&userInfo); err != nil {
			log.Fatal(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user information"})
			return
		}
		userID, err := addUser(db, userInfo.Email)
		if err != nil {
			log.Fatal(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to the database"})
			return
		}
		fmt.Println(userID)
		c.Request.Header.Set("userID", fmt.Sprint(userID))
		c.Request.Header.Set("Groups", fmt.Sprint(userInfo.Groups))

		c.Next()
	}
}

// func getUserInfo(accessToken string) (User, error) {

// 	client := &http.Client{}
// 	req, _ := http.NewRequest("GET", "http://auth.local/application/o/userinfo/", nil)
// 	req.Header.Add("Authorization", "Bearer "+accessToken)

// 	resp, err := client.Do(req)

// 	if err != nil {
// 		// Return an error response if the user info retrieval fails
// 		// c.Abort() // Stop the request processing
// 		return User{}, err
// 	}
// 	defer resp.Body.Close()
// 	data, _ := ioutil.ReadAll(resp.Body)

// 	var userInfo User
// 	if err = json.NewDecoder(strings.NewReader(string(data))).Decode(&userInfo); err != nil {
// 		return userInfo, err
// 	}

// 	return userInfo, nil
// }

func handleMain(c *gin.Context) {
	c.String(http.StatusOK, `<a href="/login">Login with OAuth2</a>`)
}

func handleLogin(c *gin.Context) {
	url := oauth2Config.AuthCodeURL("state", oauth2.AccessTypeOnline)

	c.Redirect(http.StatusFound, url)
}

func handleCallback(c *gin.Context) {
	code := c.Query("code")

	token, err := oauth2Config.Exchange(ctx, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://auth.local/application/o/userinfo/", nil)
	req.Header.Add("Authorization", "Bearer "+token.AccessToken)

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	c.JSON(http.StatusOK, gin.H{"access_token": token.AccessToken})
}

func queryDB(c *gin.Context) {
	header := c.Request.Header.Get("Groups")
	fmt.Println(reflect.TypeOf(header))
	header = strings.Trim(header, "[]")
	fmt.Println(reflect.TypeOf(header))

	groups := strings.Split(header, " ")
	fmt.Println(groups)
	specificGroup := "Admins" // Group you want to check.

	found := false
	for _, group := range groups {
		if strings.TrimSpace(group) == specificGroup {
			found = true
			break
		}
	}

	if found {
		fmt.Println("Group found!")
	} else {
		fmt.Println("Group not found!")
	}

}
