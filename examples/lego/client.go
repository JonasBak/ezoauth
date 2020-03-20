package main

import (
	"encoding/json"
	"fmt"
	"github.com/JonasBak/ezoauth"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/oauth2"
	"net/http"
	"os"
)

func handleMainNotAuthorized() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var htmlIndex = `<html>
<body>
	<a href="/login">Log In</a>
</body>
</html>`
		fmt.Fprintf(w, htmlIndex)
	})
}

func handleMainAuthorized() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value("user").(*User)
		var htmlIndex = `<html>
<body>
  <p>
  <img src="%s" width="40px" height="40px"/>
	Logged in as %s!
  </p>
	<a href="/logout">Log out</a>
</body>
</html>`
		fmt.Fprintf(w, htmlIndex, user.ProfilePicture, user.Username)
	})
}

var (
	config ezoauth.EzOauthConfig
)

type User struct {
	ezoauth.BaseUser
	Username string `gorm:"not null;unique"`
	Email    string `gorm:"not null"`

	ProfilePicture string

	Admin bool `gorm:"not null;default:false"`
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
	db.AutoMigrate(&User{}, &ezoauth.Session{})
	config = ezoauth.EzOauthConfig{
		DB: db,

		OauthConfig:      oauthConfig,
		OauthUserDataURL: "http://127.0.0.1:8000/api/v1/users/oauth2_userdata/",

		MutateRequestFunction: ezoauth.MutateRequestBearerHeader,

		UserStructMapper: func(data []byte) (ezoauth.UserInterface, string, error) {
			fmt.Println(string(data))
			user := struct {
				Username       string
				Email          string
				ProfilePicture string
				AbakusGroups   []struct {
					Name string
				}
			}{}
			err := json.Unmarshal(data, &user)
			if err != nil {
				return nil, "", err
			}
			isAdmin := false
			for i := range user.AbakusGroups {
				if user.AbakusGroups[i].Name == "Webkom" {
					isAdmin = true
				}
			}
			return &User{Username: user.Username, Email: user.Email, ProfilePicture: user.ProfilePicture, Admin: isAdmin}, user.Username, nil
		},
		UserStruct:          &User{},
		UserIdentifierField: "username",
		GormUserTable:       "users",
	}
}

func main() {
	http.Handle("/", config.AuthMuxMiddleware(handleMainAuthorized(), handleMainNotAuthorized()))
	http.HandleFunc("/login", config.HandleLogin)
	http.HandleFunc("/logout", config.HandleLogout)
	http.HandleFunc("/callback", config.HandleCallback)
	http.ListenAndServe(":8080", nil)
}
