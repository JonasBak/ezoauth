package ezoauth

import (
	"context"
	"net/http"
	"reflect"
)

func (c EzOauthConfig) AuthMuxMiddleware(success, fail http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := ""
		if cookie, err := r.Cookie(sessionCookieName); err == nil {
			value = cookie.Value
		}
		s := Session{}
		if value == "" || c.DB.Where("value = ?", value).First(&s).RecordNotFound() {
			fail.ServeHTTP(w, r)
			return
		}
		user := reflect.New(reflect.TypeOf(c.UserStruct)).Elem().Addr().Interface()
		c.DB.Table(c.GormUserTable).Where("id = ?", s.UserID).First(user)
		ctx := context.WithValue(r.Context(), "user", user)
		success.ServeHTTP(w, r.WithContext(ctx))
	})
}
