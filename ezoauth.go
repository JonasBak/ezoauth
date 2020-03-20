package ezoauth

import (
	"crypto/rand"
	"fmt"
	"github.com/jinzhu/gorm"
	"golang.org/x/oauth2"
	"io/ioutil"
	"net/http"
)

var (
	oauthStateString  = "pseudo-random"
	sessionCookieName = "session"
)

type MutateRequestFunction func(*http.Request, string) *http.Request

type BaseUser struct {
	gorm.Model

	Sessions []Session `gorm:"foreignkey:UserID"`
}

type Session struct {
	gorm.Model
	UserID uint   `gorm:"not null"`
	Value  string `gorm:"not null;unique"`
}

type UserStructMapper func(data []byte) (interface{}, string, error)

type EzOauthConfig struct {
	DB *gorm.DB

	OauthConfig      oauth2.Config
	OauthUserDataURL string

	MutateRequestFunction MutateRequestFunction

	UserStructMapper    UserStructMapper
	UserStruct          interface{}
	UserIdentifierField string
	GormUserTable       string
}

func (c EzOauthConfig) newSession(user interface{}) Session {
	b := make([]byte, 32)
	rand.Read(b)
	value := fmt.Sprintf("%x", b)
	s := Session{UserID: getID(user), Value: value}
	c.DB.Create(&s)
	return s
}

func (c EzOauthConfig) getUserInfo(state string, code string) (interface{}, error) {
	if state != oauthStateString {
		return nil, fmt.Errorf("invalid oauth state")
	}

	token, err := c.OauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}

	req, _ := http.NewRequest("GET", c.OauthUserDataURL, nil)
	req = c.MutateRequestFunction(req, token.AccessToken)
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
