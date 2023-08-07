package data

import "net/http"

type Node struct {
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Path       string                 `json:"path"`
	Name       string                 `json:"name"`
	IsDir      bool                   `json:"isDir"`
	Context    string                 `json:"context,omitempty"`
	Size       int64                  `json:"size,omitempty"`
	Part       int                    `json:"part,omitempty"`
}

type Listing struct {
	Node
	Children []Node `json:"children"`
}

type UserSecret string
type UserAttribute string
type User map[UserAttribute][]string

// Include users in config for now, to get off the ground
// The users are mapped to a secret cookie value
type Config struct {
	Users map[UserSecret]User `json:"users"`
}

var TheConfig Config

func GetUser(r *http.Request) User {
	// Get the user from the cookie
	cookie, err := r.Cookie("account")
	if err != nil {
		return nil
	}
	return TheConfig.Users[UserSecret(cookie.Value)]
}

type Attrs struct {
	Read            bool        `json:"Read"`
	Write           bool        `json:"Write"`
	Label           string      `json:"Label"`
	LabelFg         string      `json:"LabelFg"`
	LabelBg         string      `json:"LabelBg"`
	ModerationLabel string      `json:"ModerationLabel"`
	Moderation      bool        `json:"Moderation"`
	Custom          interface{} `json:"custom,omitempty"`
}
