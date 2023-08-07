package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/open-policy-agent/opa/rego"
	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/utils"
)

// Evaluate an opa string against some parsed json claims
func CalculateRego(claims data.User, regoString string) (*data.Attrs, error) {
	ctx := context.TODO()
	compiler := rego.New(
		rego.Query("data.microcms"),
		rego.Module("microcms.rego",
			regoString,
		),
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
	return &data.Attrs{
		Read:    calculation["Read"].(bool),
		Write:   calculation["Write"].(bool),
		Label:   calculation["Label"].(string),
		LabelFg: calculation["LabelFg"].(string),
		LabelBg: calculation["LabelBg"].(string),
	}, err
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
		Value: string(data.UserSecret(account)),
		Path:  "/",
	})
	http.Redirect(w, r, "..", http.StatusTemporaryRedirect)
}

func LoadConfig(cfg string) {
	f, err := os.ReadFile(cfg)
	utils.CheckErr(err, "Could not open config file")
	err = json.Unmarshal(f, &data.TheConfig)
	utils.CheckErr(err, "Could not parse config file")
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" {
		log.Printf("using AWS key: %s", os.Getenv("AWS_ACCESS_KEY_ID"))
	} else {
		log.Printf("please set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY to creds that can use AWSRekognition")
	}
}

func CanWrite(user data.User, fsPath string, fsName string) bool {
	if user != nil {
		attrs := GetAttrs(user, fsPath, fsName)
		if attrs != nil {
			canWrite, ok := attrs["Write"].(bool)
			if ok && canWrite {
				return true
			}
		}
	}
	return false
}

func CanRead(user data.User, fsPath string, fsName string) bool {
	if user != nil {
		attrs := GetAttrs(user, fsPath, fsName)
		if attrs != nil {
			canRead, ok := attrs["Read"].(bool)
			if ok && canRead {
				return true
			}
		}
	}
	return false
}
