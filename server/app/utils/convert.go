package utils

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func ToInteger(val string, defaultValue int) int {
	intVar, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return intVar
}

func GetIntOption(r *http.Request, name string, defaultValue int) int {
	params := mux.Vars(r)
	value, found := params[name]
	if !found {
		return defaultValue
	}
	return ToInteger(value, defaultValue)
}
