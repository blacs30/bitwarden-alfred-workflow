// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/blacs30/bitwarden-alfred-workflow/alfred"
	aw "github.com/deanishe/awgo"
	"github.com/ncruces/zenity"
	"github.com/oliveagle/jsonpath"
	"github.com/tidwall/gjson"
)

const (
	NOT_LOGGED_IN_MSG = "Not logged in. Need to login first."
	NOT_UNLOCKED_MSG  = "Not unlocked. Need to unlock first."
)

// Scan for projects and cache results
func runSync(force bool, last bool) {

	wf.Configure(aw.TextErrors(true))
	email := conf.Email
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", conf.BwconfKeyword))
		wf.Fatal("No email configured.")
	}
	loginErr, unlockErr := BitwardenAuthChecks()
	if loginErr != nil {
		fmt.Println(NOT_LOGGED_IN_MSG)
		searchAlfred(fmt.Sprintf("%s login", conf.BwauthKeyword))
		return
	}
	if unlockErr != nil {
		fmt.Println(NOT_UNLOCKED_MSG)
		searchAlfred(fmt.Sprintf("%s unlock", conf.BwauthKeyword))
		return
	}

	if opts.Background {
		log.Println("Running sync in background")
		if !wf.IsRunning("sync") {
			log.Printf("Starting sync job.")
			cmd := exec.Command(os.Args[0], "-sync", "-force")
			log.Println("Sync cmd: ", cmd)
			if err := wf.RunInBackground("sync", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Printf("Sync job already running.")
		}
		searchAlfred(conf.BwKeyword)
		return

	} else {
		token, err := alfred.GetToken(wf)
		if err != nil {
			wf.Fatal("Get Token error")
		}

		args := fmt.Sprintf("%s sync --session %s", conf.BwExec, token)
		message := "Syncing Bitwarden failed."
		output := "Synced."

		if force {
			args = fmt.Sprintf("%s sync --force --session %s", conf.BwExec, token)
		} else if last {
			args = fmt.Sprintf("%s sync --last --session %s", conf.BwExec, token)
			message = "Get last sync date failed."
			result, err := runCmd(args, message)
			if err != nil {
				wf.FatalError(err)
			}

			formattedTime := "No received date"
			retDate := strings.Join(result[:], "")
			if retDate != "" {
				t, _ := time.Parse(time.RFC3339, retDate)
				formattedTime = t.Format(time.RFC822)
			}
			output = fmt.Sprintf("Last sync date:\n%s", formattedTime)
			fmt.Println(output)
			return
		}

		_, err = runCmd(args, message)
		if err != nil {
			wf.FatalError(err)
		}
		// Printing the "Last sync date" or the message "synced"
		fmt.Println(output)

		// Writing the sync-cache to ensure that the sync completed
		err = wf.Cache.Store(SYNC_CACHE_NAME, []byte("sync-cache"))
		if err != nil {
			log.Println(err)
		}

		// Creating the items cache
		runCache()

		// If we ran the sync via the daemon
		// don't open the search
		backgroundSyncDaemon := os.Getenv("BACKGROUND_SYNC_DAEMON")
		if backgroundSyncDaemon == "true" {
			return
		}
		searchAlfred(conf.BwKeyword)
	}
}

// Lock Bitwarden
func runLock() {
	wf.Configure(aw.TextErrors(true))

	err := alfred.RemoveToken(wf)
	if err != nil {
		log.Println(err)
	}

	message := "Locking Bitwarden failed."
	log.Println("Clearing items cache.")
	err = clearCache()
	if err != nil {
		log.Println(err)
	}

	args := fmt.Sprintf("%s lock", conf.BwExec)
	_, err = runCmd(args, message)
	if err != nil {
		wf.FatalError(err)
	}
	fmt.Println("Locked")
}

func getItems() {
	wf.Configure(aw.TextErrors(true))
	token, err := alfred.GetToken(wf)
	if err != nil {
		wf.Fatal("Get Token error")
	}

	items := runGetItems(token)
	folders := runGetFolders(token)

	// prepare cached struct which excludes all secret data
	populateCacheItems(items)
	populateCacheFolders(folders)
}

// runGetItems uses the Bitwarden CLI to get all items and returns them to the calling function
func runGetItems(token string) []Item {
	message := "Failed to get Bitwarden items."
	args := fmt.Sprintf("%s list items --pretty --session %s", conf.BwExec, token)
	log.Println("Read latest items...")

	result, err := runCmd(args, message)
	if err != nil {
		log.Printf("Error is:\n%s", err)
		wf.FatalError(err)
	}
	// block here and return if no items (secrets) are found
	if len(result) < 1 {
		log.Println("No items found.")
		return nil
	}
	// unmarshall json
	singleString := strings.Join(result, " ")
	var items []Item
	err = json.Unmarshal([]byte(singleString), &items)
	if err != nil {
		log.Printf("Failed to unmarshall body. Err: %s\n", err)
	}
	debugLog(fmt.Sprintf("Bitwarden number of lines of returned data are: %d\n", len(result)))
	debugLog(fmt.Sprintf("Found %d items.", len(items)))
	for _, item := range items {
		debugLog(fmt.Sprintf("Name: %s Id: %s", item.Name, item.Id))
	}
	return items
}

// runGetItem gets a particular item from Bitwarden.
// It first tries to read it directly from the data.json
// if that fails it will use the Bitwarden CLI
func runGetItem() {
	wf.Configure(aw.TextErrors(true))

	// checking if -id was sent together with -getitem
	if opts.Id == "" {
		wf.Fatal("No id sent.")
		return
	}
	id := opts.Id

	// checking if -jsonpath was sent together with -getitem and -id
	jsonPath := ""
	if opts.Query != "" {
		jsonPath = opts.Query
	}
	totp := opts.Totp
	attachment := opts.Attachment

	// this assumes that the data.json was read successfully at loadBitwardenJSON()
	if bwData.UserId == "" {
		searchAlfred(fmt.Sprintf("%s login", conf.BwauthKeyword))
		wf.Fatal(NOT_LOGGED_IN_MSG)
		return
	}

	// this assumes that the data.json was read successfully at loadBitwardenJSON()
	if bwData.UserId != "" && bwData.ProtectedKey == "" {
		searchAlfred(fmt.Sprintf("%s unlock", conf.BwauthKeyword))
		wf.Fatal(NOT_UNLOCKED_MSG)
		return
	}

	// get the token from keychain
	wf.Configure(aw.TextErrors(true))
	token, err := alfred.GetToken(wf)
	if err != nil {
		wf.Fatal("Get Token error")
		return
	}

	receivedItem := ""
	isDecryptSecretFromJsonFailed := false

	// handle attachments later, via Bitwarden CLI
	// this decrypts the secrets in the data.json
	if bwData.UserId != "" && (attachment == "") {
		log.Printf("Getting item for id %s", id)
		sourceKey, err := MakeDecryptKeyFromSession(bwData.ProtectedKey, token)
		if err != nil {
			log.Printf("Error making source key is:\n%s", err)
			isDecryptSecretFromJsonFailed = true
		}

		encryptedSecret := ""
		if bwData.path != "" {
			data, err := os.ReadFile(bwData.path)
			if err != nil {
				log.Print("Error reading file ", bwData.path)
				isDecryptSecretFromJsonFailed = true
			}
			// replace starting bracket with dot as gsub uses a dot for the first group in an array
			jsonPath = strings.Replace(jsonPath, "[", ".", -1)
			jsonPath = strings.Replace(jsonPath, "]", "", -1)
			if totp {
				jsonPath = "login.totp"
			}

			var value gjson.Result
			if bwData.ActiveUserId != "" {
				// different location for version 1.21.1 and above
				value = gjson.Get(string(data), fmt.Sprintf("%s.data.ciphers.encrypted.%s.%s", bwData.UserId, id, jsonPath))
			} else {
				value = gjson.Get(string(data), fmt.Sprintf("ciphers_%s.%s.%s", bwData.UserId, id, jsonPath))
			}
			if value.Exists() {
				encryptedSecret = value.String()
				debugLog(fmt.Sprintf("encryptedSecret value is: %v [truncated]", encryptedSecret[:5]))
			} else {
				log.Print("Error, value for gjson not found.")
				isDecryptSecretFromJsonFailed = true
			}
		}

		decryptedString, err := DecryptString(encryptedSecret, sourceKey)
		if err != nil {
			log.Print(err)
			isDecryptSecretFromJsonFailed = true
		}
		if totp {
			decryptedString, err = otpKey(decryptedString)
			if err != nil {
				log.Print("Error getting topt key, ", err)
				isDecryptSecretFromJsonFailed = true
			}
		}
		receivedItem = decryptedString
	}
	if bwData.UserId == "" || isDecryptSecretFromJsonFailed || attachment != "" {
		log.Println("Falling back to Bitwarden CLI to get item.")

		// Run the Bitwarden CLI to get the secret
		// Use it also for getting attachments
		if attachment != "" {
			log.Printf("Getting attachment %s for id %s", attachment, id)
		}

		message := "Failed to get Bitwarden item."
		args := fmt.Sprintf("%s get item %s --pretty --session %s", conf.BwExec, id, token)
		if totp {
			log.Println("Trying to get totp  via Bitwarden cli")
			args = fmt.Sprintf("%s get totp %s --session %s", conf.BwExec, id, token)
		} else if attachment != "" {
			log.Println("Trying to get attachment  via Bitwarden cli")
			args = fmt.Sprintf("%s get attachment %s --itemid %s --output %s --session %s --raw", conf.BwExec, attachment, id, conf.OutputFolder, token)
		}

		result, err := runCmd(args, message)
		if err != nil {
			log.Printf("Error is:\n%s", err)
			wf.FatalError(err)
			return
		}
		// block here and return if no items (secrets) are found
		if len(result) <= 0 {
			log.Println("No items found.")
			return
		}

		if totp {
		} else {
			receivedItem = ""
		}
		if jsonPath != "" && !totp {
			log.Println("Sent jsonPath is", jsonPath)
			// jsonpath operation to get only required part of the item
			singleString := strings.Join(result, " ")
			var item interface{}
			err = json.Unmarshal([]byte(singleString), &item)
			if err != nil {
				log.Println(err)
			}
			res, err := jsonpath.JsonPathLookup(item, fmt.Sprintf("$.%s", jsonPath))
			if err != nil {
				log.Println(err)
				return
			}
			receivedItem = fmt.Sprintf("%v", res)
			if wf.Debug() {
				log.Printf("Received key is: %s*", receivedItem[0:2])
			}
		} else {
			receivedItem = strings.Join(result, " ")
		}
	}
	fmt.Print(receivedItem)
}

func runGetFolders(token string) []Folder {
	message := "Failed to get Bitwarden Folders."
	args := fmt.Sprintf("%s list folders --pretty --session %s", conf.BwExec, token)
	log.Println("Read latest folders...")

	result, err := runCmd(args, message)
	if err != nil {
		log.Printf("Error is:\n%s", err)
		wf.FatalError(err)
	}
	// block here and return if no items (secrets) are found
	if len(result) <= 0 {
		log.Println("No folders found.")
		return nil
	}
	// unmarshall json
	singleString := strings.Join(result, " ")
	var folders []Folder
	err = json.Unmarshal([]byte(singleString), &folders)
	if err != nil {
		log.Printf("Failed to unmarshall body. Err: %s", err)
	}
	if wf.Debug() {
		log.Printf("Bitwarden number of lines of returned data are: %d\n", len(result))
		log.Println("Found ", len(folders), " items.")
		for _, item := range folders {
			log.Println("Name: ", item.Name, ", Id: ", item.Id)
		}
	}
	return folders
}

// Unlock Bitwarden
func runUnlock() {
	wf.Configure(aw.TextErrors(true))
	email := conf.Email
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", conf.BwconfKeyword))
		wf.Fatal("No email configured.")
	}

	_, pw, _ := zenity.Password(
		zenity.Title(fmt.Sprintf("Unlock account %s", email)),
	)

	// set the password from the returned slice
	if len(pw) < 1 {
		wf.Fatal("No Password returned.")
	}

	os.Setenv("PASS", pw)

	// Unlock Bitwarden now
	message := "Unlocking Bitwarden failed."
	args := fmt.Sprintf("%s unlock --raw --passwordenv PASS", conf.BwExec)
	debugLog(fmt.Sprintf("bw unlock command is %s", args))
	tokenReturn, err := runCmd(args, message)
	os.Unsetenv("PASS")
	if err != nil {
		wf.FatalError(err)
	}

	// set the password from the returned slice
	token := ""
	if len(tokenReturn) > 0 {
		token = tokenReturn[0]
	} else {
		wf.Fatal("No token returned after unlocking.")
	}
	err = alfred.SetToken(wf, token)
	if err != nil {
		log.Println(err)
	}
	debugLog(fmt.Sprintf("first few chars of the token is %s", token[0:2]))

	if conf.UseApikey {
		// Writing the sync-cache because we have unlocked the vault in apikey mode
		// Items should be present
		err = wf.Cache.Store(SYNC_CACHE_NAME, []byte("sync-cache"))
		if err != nil {
			log.Println(err)
		}
	}

	// Creating the items cache
	if wf.Cache.Exists(SYNC_CACHE_NAME) {
		runCache()
		searchAlfred(conf.BwKeyword)
	}
	fmt.Println("Unlocked")
}

// Login to Bitwarden
func runLogin() {
	wf.Configure(aw.TextErrors(true))
	email := conf.Email
	sfa := conf.Sfa
	sfaMode := conf.SfaMode
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", conf.BwconfKeyword))
		if wf.Debug() {
			log.Println("[ERROR] ==> Email missing. Bitwarden not configured yet")
		}
		wf.Fatal("No email configured.")
	}

	if !conf.UseApikey {
		_, pw, _ := zenity.Password(
			zenity.Title(fmt.Sprintf("Login account %s", email)),
		)
		if len(pw) < 1 {
			return
		}
		os.Setenv("PASS", pw)
	}

	args := fmt.Sprintf("%s login %s --passwordenv PASS", conf.BwExec, email)
	debugLog(fmt.Sprintf("bw login command is %s", args))

	if conf.UseApikey {
		log.Println("Use apikey", conf.UseApikey)
		client_id, _ := zenity.Entry("Enter API Key client_id:",
			zenity.Title(fmt.Sprintf("Login account %s", email)))
		if len(client_id) < 1 {
			if wf.Debug() {
				log.Println("[DEBUG] ==> client_id is empty")
			}
			fmt.Println("Empty client_id received")
			return
		}
		client_secret, _ := zenity.Entry("Enter API Key client_secret:",
			zenity.Title(fmt.Sprintf("Login account %s", email)))
		if len(client_secret) < 1 {
			if wf.Debug() {
				log.Println("[DEBUG] ==> client_secret is empty")
			}
			fmt.Println("Empty client_secret received")
			return
		}

		os.Setenv("BW_CLIENTID", client_id)
		os.Setenv("BW_CLIENTSECRET", client_secret)
		args = fmt.Sprintf("%s login --apikey", conf.BwExec)

	} else if sfa {
		display2faMode := map2faMode(sfaMode)
		sfacodeReturn := ""
		if sfaMode == 0 {
			sfacodeReturn, _ = zenity.Entry("Enter Authentictor code:",
				zenity.Title(fmt.Sprintf("Login account %s", email)))
		} else if sfaMode == 1 {

			emailMessage := "Failed to request Bitwarden email token."
			emailArgs := fmt.Sprintf("%s login %s --raw --method %d --passwordenv PASS", conf.BwExec, email, sfaMode)
			emailReturn, err := runCmdWithContext(conf.EmailMaxWait, emailArgs, emailMessage)
			if err != nil {
				wf.FatalError(err)
			}
			if emailReturn[0] != "Two-step login code" {
				log.Println("[DEBUG] ==> unexpected email code response.")
			}

			sfacodeReturn, _ = zenity.Entry("Enter Email authentication code that was sent to you:",
				zenity.Title(fmt.Sprintf("Login account %s", email)))
		} else if sfaMode == 3 {
			sfacodeReturn, _ = zenity.Entry("Enter Yubicey OTP code:",
				zenity.Title(fmt.Sprintf("Login account %s", email)))
		} else {
			fmt.Printf("Unsupported 2fa mode %q.\n", display2faMode)
			return
		}

		sfaCode := ""
		if len(sfacodeReturn) > 0 {
			sfaCode = sfacodeReturn
		} else {
			wf.Fatal("No 2FA code returned.")
		}

		args = fmt.Sprintf("%s login %s --passwordenv PASS --raw --method %d --code %s", conf.BwExec, email, sfaMode, sfaCode)
	}

	message := "Login to Bitwarden failed."
	tokenReturn, err := runCmd(args, message)
	os.Unsetenv("PASS")
	if err != nil {
		wf.FatalError(err)
	}

	// If we use the api key no token is returned, we first need to run unlock
	if conf.UseApikey {
		fmt.Println("APIKEY.")
		return
	}

	// set the token from the returned result
	token := ""
	if len(tokenReturn) > 0 {
		token = tokenReturn[0]
	} else {
		wf.Fatal("No token returned after unlocking.")
	}
	err = alfred.SetToken(wf, token)
	if err != nil {
		log.Println(err)
	}
	debugLog(fmt.Sprintf("first few chars of the token is %s", token[0:2]))

	// Writing the sync-cache because data is synced for Yubikey and Authenticator login
	// Just the APIKEY login needs a separate unlock and therefore sync
	err = wf.Cache.Store(SYNC_CACHE_NAME, []byte("sync-cache"))
	if err != nil {
		log.Println(err)
	}

	// Creating the items cache
	runCache()

	searchAlfred(conf.BwKeyword)
	fmt.Println("Logged In.")
}

// Logout from Bitwarden
func runLogout() {
	wf.Configure(aw.TextErrors(true))

	err := alfred.RemoveToken(wf)
	if err != nil {
		log.Println(err)
	}

	args := fmt.Sprintf("%s logout", conf.BwExec)

	log.Println("Clearing items cache.")
	err = wf.ClearCache()
	if err != nil {
		log.Println(err)
	}
	message := "Logout of Bitwarden failed."
	_, err = runCmd(args, message)
	if err != nil {
		wf.FatalError(err)
	}
	fmt.Println("Logged Out")
}

func runCache() {
	err := clearCache()
	if err != nil {
		log.Print("Error while deleting Caches ", err)
	}

	wf.Configure(aw.TextErrors(true))
	email := conf.Email
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", conf.BwconfKeyword))
		wf.Fatal("No email configured.")
	}

	log.Println("Running cache")
	getItems()
}
