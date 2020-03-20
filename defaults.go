package ezoauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
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

func TagMapper(userStruct UserInterface, identifier string) UserStructMapper {
	v := reflect.TypeOf(userStruct)
	fields := []reflect.StructField{}
	for i := 0; i < v.NumField(); i++ {
		if f, ok := v.Field(i).Tag.Lookup("ez"); ok {
			fields = append(fields, reflect.StructField{
				Name: v.Field(i).Name,
				Type: v.Field(i).Type,
				Tag:  reflect.StructTag(fmt.Sprintf(`json:"%s"`, f)),
			})
		}
	}
	typ := reflect.StructOf(fields)
	return func(data []byte) (UserInterface, string, error) {
		u := reflect.New(typ).Elem()
		err := json.Unmarshal(data, u.Addr().Interface())
		if err != nil {
			return nil, "", err
		}
		user := reflect.New(reflect.TypeOf(userStruct)).Elem()
		for i := 0; i < typ.NumField(); i++ {
			fn := typ.Field(i).Name
			switch typ.Field(i).Type.Kind() {
			case reflect.Float32, reflect.Float64:
				user.FieldByName(fn).SetFloat(u.FieldByName(fn).Float())
			case reflect.Bool:
				user.FieldByName(fn).SetBool(u.FieldByName(fn).Bool())
			case reflect.String:
				user.FieldByName(fn).SetString(u.FieldByName(fn).String())
			}
		}
		return user.Addr().Interface().(UserInterface), user.FieldByName(identifier).String(), nil
	}
}
