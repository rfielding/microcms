package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/open-policy-agent/opa/rego"
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

// Evaluate an opa string against some parsed json claims
func CalculateRego(claims interface{}, s string) (map[string]interface{}, error) {
	ctx := context.TODO()
	compiler := rego.New(
		rego.Query("data.gosqlite"),
		rego.Module("gosqlite.rego", s),
	)
	query, err := compiler.PrepareForEval(ctx)
	if err != nil {
		return nil, err
	}

	results, err := query.Eval(ctx, rego.EvalInput(claims))
	if err != nil {
		return nil, err
	}
	calculation := results[0].Expressions[0].Value.(map[string]interface{})
	return calculation, err
}

func LoadConfig() {
	f, err := ioutil.ReadFile("./config.json")
	CheckErr(err, "Could not open config file")
	err = json.Unmarshal(f, &theConfig)
	CheckErr(err, "Could not parse config file")
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		log.Printf("using AWS key: %s", os.Getenv("AWS_ACCESS_KEY_ID"))
	} else {
		log.Printf("please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY to creds that can use AWSRekognition")
	}

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
	http.Redirect(w, r, "..", http.StatusTemporaryRedirect)
}
