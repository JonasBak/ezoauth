package ezoauth

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

type EzOauthConfig struct {
	DB *gorm.DB

	OauthConfig      oauth2.Config
	OauthUserDataURL string

	UserStructMapper    UserStructMapper
	UserStruct          interface{}
	UserIdentifierField string
	GormUserTable       string
}

func (c EzOauthConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := c.OauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c EzOauthConfig) HandleCallback(w http.ResponseWriter, r *http.Request) {
	user, err := c.getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// TODO
	// get user from getUserInfo
	// Create session
	// Set session token and redirect to /
	fmt.Fprintf(w, "%+v", user)
}
func (c EzOauthConfig) getUserInfo(state string, code string) (interface{}, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := c.OauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?access_token=%s", c.OauthUserDataURL, token.AccessToken), nil)
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

	result := c.DB.Table(c.GormUserTable).Where(map[string]interface{}{c.UserIdentifierField: identifier}).Updates(user).First(user)
	if result.RecordNotFound() {
		if err := c.DB.Create(user).Error; err != nil {
			return nil, fmt.Errorf("failed creating user '%s': %s", identifier, err.Error())
		}
	} else if result.Error != nil {
		return nil, fmt.Errorf("failed updating user info for user '%s': %s", identifier, result.Error.Error())
	}

	return user, nil
}
