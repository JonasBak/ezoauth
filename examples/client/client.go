package main

import (
	"encoding/json"
	"fmt"
	"github.com/JonasBak/ezoauth"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/oauth2"
	"net/http"
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
	config ezoauth.EzOauthConfig
)

type User struct {
	gorm.Model
	Username string `gorm:"not null;unique"`
	Email    string
}

func init() {
	oauthConfig := oauth2.Config{
		RedirectURL:  "http://localhost:8080/callback",
		ClientID:     "222222",
		ClientSecret: "22222222",
		Scopes:       []string{"all"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://127.0.0.1:9096/authorize",
			TokenURL: "http://127.0.0.1:9096/token",
		},
	}
	db, err := gorm.Open("sqlite3", "test.db")
	if err != nil {
		panic("failed to connect database")
	}
	db.LogMode(true)
	db.AutoMigrate(&User{})
	config = ezoauth.EzOauthConfig{
		DB: db,

		OauthConfig:      oauthConfig,
		OauthUserDataURL: "http://127.0.0.1:9096/oauth2_userdata",

		UserStructMapper: func(data []byte) (interface{}, string, error) {
			user := struct {
				Username string
				Email    string
			}{}
			err := json.Unmarshal(data, &user)
			if err != nil {
				return nil, "", err
			}
			return &User{Username: user.Username, Email: user.Email}, user.Username, nil
		},
		UserStruct:          User{},
		UserIdentifierField: "username",
		GormUserTable:       "users",
	}
}

func main() {
	http.HandleFunc("/", handleMain)
	http.HandleFunc("/login", config.HandleLogin)
	http.HandleFunc("/callback", config.HandleCallback)
	http.ListenAndServe(":8080", nil)
}
