package handler

import (
	"encoding/json"
	"log"
	"path"
	"strings"

	"github.com/rfielding/microcms/data"
	"github.com/rfielding/microcms/fs"
)

// Permission attributes are dynamic, and can come from parent directories.
// The first one found is used to set them all.
// fsPath does NOT begin with a slash, and ends with a slash
func getAttrsPermission(claims data.User, fsPath string, fsName string, initial map[string]interface{}) map[string]interface{} {
	// Try exact file if fName is not blank
	regoFile := fsPath + "permissions.rego"
	if fsName != "" {
		regoFile = fsPath + fsName + "--permissions.rego"
		if fs.IsDir(fsPath + fsName) {
			regoFile = fsPath + fsName + "/permissions.rego"
		}
	}
	if fs.IsExist(regoFile) {
		jf, err := fs.ReadFile(regoFile)
		if err != nil {
			log.Printf("Failed to open %s!: %v", regoFile, err)
		} else {
			regoString := string(jf)
			calculation, err := CalculateRego(claims, regoString)
			if err != nil {
				log.Printf("Failed to parse rego %s!: %v\n%s", regoFile, err, regoString)
			}
			for k, v := range calculation {
				initial[k] = v
			}
		}
		return initial
	} else {
		if fsName != "" {
			return getAttrsPermission(claims, fsPath, "", initial)
		} else {
			if fsPath == "/files/" {
				return initial
			} else {
				// careful! if it ends in slash, then parent is same file, fsName is blank!
				fsPath := path.Dir(path.Clean(fsPath)) + "/"
				fsName := ""
				return getAttrsPermission(claims, fsPath, fsName, initial)
			}
		}
	}
}

func GetAttrs(claims data.User, fsPath string, fsName string) map[string]interface{} {
	// always get attributes according to the original file
	if strings.Contains(fsName, "--") {
		//a better pattern would be:  *.*--*.*
		fNameOriginal := fsName[0:strings.LastIndex(fsName, "--")]
		fsName = fNameOriginal
	}
	// Start parsing attributes with a custom set of values that
	// get overridden with calculated values
	attrs := loadCustomAttrs(fsPath, fsName)
	// If there is content moderation, then add it in here
	loadModerationAttrs(fsPath, fsName, attrs)
	// overwrite with calculated values
	return getAttrsPermission(claims, fsPath, fsName, attrs)
}

type Moderation struct {
	Name string `json:"Name"`
}

type ModerationData struct {
	ModerationLabels []Moderation `json:"ModerationLabels"`
}

func loadModerationAttrs(fsPath string, fsName string, attrs map[string]interface{}) {
	mods := ModerationData{}
	moderationFileName := fsPath + fsName + "--moderation.json"
	if fs.IsExist(moderationFileName) {
		jf, err := fs.ReadFile(moderationFileName)
		if err != nil {
			log.Printf("Failed to open %s!: %v", moderationFileName, err)
			return
		}
		err = json.Unmarshal(jf, &mods)
		if err != nil {
			log.Printf("Failed to parse json %s!: %v", moderationFileName, err)
		}
		if len(mods.ModerationLabels) > 0 {
			attrs["Moderation"] = true
			attrs["ModerationLabel"] = mods.ModerationLabels[0].Name
		}
	}
}

func loadCustomAttrs(fsPath string, fsName string) map[string]interface{} {
	attrs := make(map[string]interface{})

	customFileName := fsPath + fsName + "--custom.json"
	if fs.IsExist(customFileName) {
		jf, err := fs.ReadFile(customFileName)
		if err != nil {
			log.Printf("Failed to open %s!: %v", customFileName, err)
		} else {
			err := json.Unmarshal(jf, &attrs)
			if err != nil {
				log.Printf("Failed to parse json %s!: %v", customFileName, err)
			}
		}
	}
	return attrs
}
