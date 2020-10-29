package alfred

import (
	"fmt"
	aw "github.com/deanishe/awgo"
	"github.com/jychri/tilde"
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

func GetEmail(wf *aw.Workflow, configEmail string, bwEmail string) string {
	if configEmail == "" {
		err := SetEmail(wf, bwEmail)
		if err != nil {
			log.Println(err)
			return ""
		}
		configEmail = bwEmail
	}
	return configEmail
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
