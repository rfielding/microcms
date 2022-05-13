package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

var theConfig Config

type UserSecret string
type UserAttribute string
type User map[UserAttribute][]string

// Include users in config for now, to get off the ground
// The users are mapped to a secret cookie value
type Config struct {
	Users map[UserSecret]User `json:"users"`
}

func LoadConfig() {
	f, err := ioutil.ReadFile("./config.json")
	CheckErr(err, "Could not open config file")
	err = json.Unmarshal(f, &theConfig)
	CheckErr(err, "Could not parse config file")
}

func GetUser(r *http.Request) User {
	// Get the user from the cookie
	cookie, err := r.Cookie("account")
	if err != nil {
		return nil
	}
	return theConfig.Users[UserSecret(cookie.Value)]
}

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// Get the account from the cookie
	//
	// The account number was a secret mailed to the user, in a url like:
	//   http://localhost:9321/registration/?account=123456789
	//
	q := r.URL.Query()
	account := q.Get("account")
	http.SetCookie(w, &http.Cookie{
		Name:  "account",
		Value: string(UserSecret(account)),
		Path:  "/",
	})
}
