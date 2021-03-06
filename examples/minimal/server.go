package main

import (
	"fmt"
	"github.com/JonasBak/ezoauth"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/oauth2"
	"net/http"
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
	Logged in as %s!
	<a href="/logout">Log out</a>
</body>
</html>`
		fmt.Fprintf(w, htmlIndex, user.Username)
	})
}

var (
	config ezoauth.EzOauthConfig
)

type User struct {
	ezoauth.BaseUser
	Username string `gorm:"not null;unique" ez:"username"`
	Email    string `ez:"email"`
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
	db.AutoMigrate(&User{}, &ezoauth.Session{})
	config = ezoauth.EzOauthConfig{
		DB:                    db,
		OauthConfig:           oauthConfig,
		OauthUserDataURL:      "http://127.0.0.1:9096/oauth2_userdata",
		MutateRequestFunction: ezoauth.MutateRequestQueryParameter,
		UserStructMapper:      ezoauth.TagMapper(User{}, "Username"),
		UserStruct:            User{},
		UserIdentifierField:   "username",
		GormUserTable:         "users",
	}
}

func main() {
	http.Handle("/", config.AuthMuxMiddleware(handleMainAuthorized(), handleMainNotAuthorized()))
	http.HandleFunc("/login", config.HandleLogin)
	http.HandleFunc("/logout", config.HandleLogout)
	http.HandleFunc("/callback", config.HandleCallback)
	http.ListenAndServe(":8080", nil)
}
