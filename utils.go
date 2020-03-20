package ezoauth

import (
	"reflect"
)

func getID(obj interface{}) uint {
	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		return getID(reflect.ValueOf(obj).Elem().Interface())
	} else {
		return reflect.ValueOf(obj).FieldByName("BaseUser").Interface().(BaseUser).Model.ID
	}
}
