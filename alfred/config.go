package alfred

import (
	"encoding/json"
	"fmt"
	aw "github.com/deanishe/awgo"
	"github.com/jychri/tilde"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func GetOutputFolder(wf *aw.Workflow, folder string) string {
	if folder == "" {
		defaultFolder, err := os.UserHomeDir()
		if err != nil {
			log.Print("Error getting UserHomeDir ", err)
		}
		folder = fmt.Sprintf("%s/Downloads/", defaultFolder)
	} else {
		// tilde.Abs() expands ~ to /Users/$user
		folder = tilde.Abs(folder)
		// in case a "/" exist on the right side, remove it and add it again to be sure it exists.
		folder = fmt.Sprintf("%s/", strings.TrimRight(folder, "/"))
	}
	return folder
}

func GetEmail(wf *aw.Workflow, email string) string {
	if email == "" {
		var bwData BwData
		succ, err := OpenBitwardenData(&bwData)
		if err != nil {
			log.Println(err)
			return ""
		}
		if succ {
			err := SetEmail(wf, bwData.UserEmail)
			if err != nil {
				log.Println(err)
				return ""
			}
			email = bwData.UserEmail
		}
	}
	return email
}

//Set keys
func SetServer(wf *aw.Workflow, url string) error {
	return wf.Config.Set("SERVER_URL", url, false).Do()
}

func SetEmail(wf *aw.Workflow, address string) error {
	return wf.Config.Set("EMAIL", address, false).Do()
}

func SetSfa(wf *aw.Workflow, enabled string) error {
	return wf.Config.Set("2FA_ENABLED", enabled, true).Do()
}

func SetSfaMode(wf *aw.Workflow, id string) error {
	return wf.Config.Set("2FA_MODE", id, true).Do()
}

func OpenBitwardenData(bwData interface{}) (bool, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}
	bwDataPath := fmt.Sprintf("%s/Library/Application Support/Bitwarden CLI/data.json", homedir)
	log.Println("BW DataPath", bwDataPath)
	if _, err := os.Stat(bwDataPath); err != nil {
		log.Println("Couldn't find the Bitwarden data.json ", err)
		return false, err
	}
	data, err := ioutil.ReadFile(bwDataPath)
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(data, &bwData); err != nil {
		log.Printf("Couldn't load the items cache, error: %s", err)
		return false, err
	}
	log.Println("Got existing Bitwarden CLI data")
	return true, nil
}

type BwData struct {
	InstalledVersion string                 `json:"installedVersion"`
	UserEmail        string                 `json:"userEmail"`
	Unused           map[string]interface{} `json:"-"`
}
