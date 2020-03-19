package ezoauth

import (
	"fmt"
	"net/http"
	"time"
)

func (c EzOauthConfig) HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := c.OauthConfig.AuthCodeURL(oauthStateString)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (c EzOauthConfig) HandleLogout(w http.ResponseWriter, r *http.Request) {
	value := ""
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		value = cookie.Value
	}
	s := Session{}
	if !c.DB.Where("value = ?", value).First(&s).RecordNotFound() {
		c.DB.Delete(&s)
		cookie := http.Cookie{
			Name:  sessionCookieName,
			Value: "",
		}
		http.SetCookie(w, &cookie)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func (c EzOauthConfig) HandleCallback(w http.ResponseWriter, r *http.Request) {
	user, err := c.getUserInfo(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	s := c.newSession(user)
	expire := time.Now().AddDate(0, 0, 1)
	cookie := http.Cookie{
		Name:    sessionCookieName,
		Value:   s.Value,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
