package main

import (
	"fmt"
	"github.com/JonasBak/infrastucture/containers/oauth"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

func handleMain(w http.ResponseWriter, r *http.Request) {
	// TODO
	// Middleware checks session token, sets user in context
	var htmlIndex = `<html>
<body>
	<a href="/login"> Log In</a>
</body>
</html>`
	fmt.Fprintf(w, htmlIndex)
}

var (
	config oauthgorm.OauthGormConfig
)

type User struct {
	gorm.Model
	Username string `gorm:"not null;unique"`
}

func init() {
	oauthConfig := oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"user"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://127.0.0.1:8000/authorization/oauth2/authorize/",
			TokenURL: "http://127.0.0.1:8000/authorization/oauth2/token/",
		},
	}
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	db.LogMode(true)
	db.AutoMigrate(&User{})
	config = oauthgorm.OauthGormConfig{
		OauthConfig: oauthConfig,
		DB:          db,
		UserStructMapper: func(data []byte) (map[string]interface{}, string, error) {
			return map[string]interface{}{"username": "webkom"}, "webkom", nil
		},
		UserStruct:          User{},
		UserIdentifierField: "username",
		GormUserTable:       "users",
	}
}

// TODO
// For testing and dev:
// https://github.com/RichardKnop/go-oauth2-server (or maybe alternative, hydra?)

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", config.HandleLogin)
	http.HandleFunc("/callback", config.HandleCallback)
	http.ListenAndServe(":8080", nil)
	fmt.Println("yeet")
}
