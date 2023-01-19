package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

// User represents a user in the system
type User struct {
	gorm.Model
	Username string
	Password string
}

func init() {
	// Connect to the MySQL database
	var err error
	db, err = gorm.Open("mysql", "root:123@tcp(localhost:3306)/dbapi?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		fmt.Println(err)
	}

	// Automigrate the User table
	db.AutoMigrate(&User{})
}

func main() {
	router := gin.Default()
	// Set up the session store
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))

	// Load HTML templates
	router.LoadHTMLGlob("templates/*")

	// Define the login and signup routes
	router.GET("/login", loginForm)
	router.POST("/login", login)
	router.GET("/", dashboard)
	router.GET("/logout", logout)
	router.GET("/signup", signupForm)
	router.GET("/dashboard", dashboard)
	router.POST("/signup", signup)

	// Start the server
	router.Run(":8080")
}


func loginForm(c *gin.Context) {
	// Render the login form template
	c.HTML(http.StatusOK, "login.tmpl", gin.H{})
}

func login(c *gin.Context) {
	// Get the username and password from the form
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Find the user in the database
	var user User
	if err := db.Where("username = ? AND password = ?", username, password).First(&user).Error; err != nil {
		c.HTML(http.StatusUnauthorized, "login.tmpl", gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	// Set the user's ID in the session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

	// Redirect to the protected page
	c.Redirect(http.StatusFound, "/dashboard")
}

func signupForm(c *gin.Context) {
	// Render the signup form template
	c.HTML(http.StatusOK, "signup.tmpl", gin.H{})
}

func dashboard(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// Render the dashboard template
	c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{})
}

func logout(c *gin.Context) {
    session := sessions.Default(c)
    session.Delete("user_id")
    session.Save()
    c.Redirect(http.StatusFound, "/login")
}

func signup(c *gin.Context) {
	// Get the username and password from the form
	username := c.PostForm("username")
	password := c.PostForm("password")

	// Check if the username is already taken
	var user User
	if db.Where("username = ?", username).First(&user).RecordNotFound() {
		// Create a new user
		user := User{Username: username, Password: password}
		if err := db.Create(&user).Error; err != nil {
		c.HTML(http.StatusInternalServerError, "signup.tmpl", gin.H{
		"error": "Failed to create the user",
		})
		return
		}
		// Set the user's ID in the session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	session.Save()

	// Redirect to the login page
	c.Redirect(http.StatusFound, "/login")
} else {
	c.HTML(http.StatusBadRequest, "signup.tmpl", gin.H{
		"error": "The email is already taken",
	})
}
}
