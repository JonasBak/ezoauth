package oauthgorm

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
)

var (
	oauthStateString = "pseudo-random"
)

type UserStructMapper func(data []byte) (interface{}, string, error)

type OauthGormConfig struct {
	OauthConfig oauth2.Config
	DB          *gorm.DB

	UserStructMapper    UserStructMapper
	UserStruct          interface{}
	UserIdentifierField string
	GormUserTable       string
}

func (c OauthGormConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := c.OauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c OauthGormConfig) HandleCallback(w http.ResponseWriter, r *http.Request) {
	_, err := c.getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// TODO
	// get user from getUserInfo
	// Create session
	// Set session token and redirect to /
	fmt.Fprintf(w, "OK!")
}
func (c OauthGormConfig) getUserInfo(state string, code string) (interface{}, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}
	token, err := c.OauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	// response, err := http.Get("http://127.0.0.1/api/v1/users/oauth2_userdata/")
	req, _ := http.NewRequest("GET", "http://127.0.0.1:8000/api/v1/users/oauth2_userdata/", nil)
	req.Header.Add("AUTHORIZATION", "Bearer "+token.AccessToken)
	response, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading response body: %s", err.Error())
	}

	user, identifier, err := c.UserStructMapper(contents)
	if err != nil {
		return nil, fmt.Errorf("failed mapping response to user object: %s", err.Error())
	}
	result := c.DB.Debug().Table(c.GormUserTable).Where(map[string]interface{}{c.UserIdentifierField: identifier}).FirstOrCreate(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed querying database for user: %s", result.Error.Error())
	}

	return user, nil
}
