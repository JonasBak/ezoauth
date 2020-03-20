package ezoauth

import (
	"fmt"
	"net/http"
	"net/url"
)

func MutateRequestQueryParameter(req *http.Request, token string) *http.Request {
	q, _ := url.ParseQuery(req.URL.RawQuery)
	q.Add("access_token", token)
	req.URL.RawQuery = q.Encode()
	return req
}

func MutateRequestBearerHeader(req *http.Request, token string) *http.Request {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	return req
}
