// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"
	"github.com/blacs30/bitwarden-alfred-workflow/alfred"
	aw "github.com/deanishe/awgo"
	"github.com/oliveagle/jsonpath"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	NOT_LOGGED_IN_MSG = "Not logged in. Need to login first."
	NOT_UNLOCKED_MSG  = "Not unlocked. Need to unlock first."
)

// Scan for projects and cache results
func runSync(force bool, last bool) {
	log.Println("Clearing items cache.")
	err := wf.Cache.StoreJSON(CACHE_NAME, nil)
	if err != nil {
		log.Println(err)
	}
	err = wf.Cache.StoreJSON(FOLDER_CACHE_NAME, nil)
	if err != nil {
		log.Println(err)
	}
	err = wf.Cache.StoreJSON(AUTO_FETCH_CACHE, nil)
	if err != nil {
		log.Println(err)
	}

	wf.Configure(aw.TextErrors(true))
	email, _, _, _ := getConfigs(wf)
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", BWCONF_KEYWORD))
		wf.Fatal("No email configured.")
	}
	loginErr, unlockErr := BitwardenAuthChecks()
	if loginErr != nil {
		fmt.Println(NOT_LOGGED_IN_MSG)
		searchAlfred(fmt.Sprintf("%s login", BWAUTH_KEYWORD))
		return
	}
	if unlockErr != nil {
		fmt.Println(NOT_UNLOCKED_MSG)
		searchAlfred(fmt.Sprintf("%s unlock", BWAUTH_KEYWORD))
		return
	}

	log.Println("Background?", opts.Background)
	if opts.Background {
		if !wf.IsRunning("sync") {
			cmd := exec.Command(os.Args[0], "-sync", "-force")
			log.Println("Sync cmd: ", cmd)
			if err := wf.RunInBackground("sync", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Printf("Sync job already running.")
		}
		searchAlfred(BW_KEYWORD)
		return

	} else {
		token, err := alfred.GetToken(wf)
		if err != nil {
			wf.Fatal("Get Token error")
		}

		args := ""
		message := "Syncing Bitwarden failed."
		output := "Synced."

		if force {
			args = fmt.Sprintf("%s sync --force --session %s", BwExec, token)
		} else if last {
			args = fmt.Sprintf("%s sync --last --session %s", BwExec, token)
			message = "Get last sync date failed."
		} else {
			args = fmt.Sprintf("%s sync --session %s", BwExec, token)
		}

		result, err := runCmd(args, message)
		if err != nil {
			wf.FatalError(err)
		}
		if last {
			formattedTime := "No received date"
			retDate := strings.Join(result[:], "")
			if retDate != "" {
				t, _ := time.Parse(time.RFC3339, retDate)
				formattedTime = t.Format(time.RFC822)
			}
			output = fmt.Sprintf("Last sync date:\n%s", formattedTime)
		}

		if !last {
			getItems()
		}
		fmt.Println(output)
		err = wf.Cache.Store(SYNC_CACH_NAME, []byte(string("sync-cache")))
		if err != nil {
			log.Println(err)
		}
		return
	}
}

// Lock Bitwarden
func runLock() {
	wf.Configure(aw.TextErrors(true))

	err := alfred.RemoveToken(wf)
	if err != nil {
		log.Println(err)
	}

	args := fmt.Sprintf("%s lock", BwExec)

	message := "Locking Bitwarden failed."
	log.Println("Clearing items cache.")
	err = wf.ClearCache()
	if err != nil {
		log.Println(err)
	}
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
	popuplateCacheItems(items)
	popuplateCacheFolders(folders)
}

func runGetItems(token string) []Item {
	message := "Failed to get Bitwarden items."
	args := fmt.Sprintf("%s list items --pretty --session %s", BwExec, token)
	log.Println("Read latest items...")

	result, err := runCmd(args, message)
	if err != nil {
		log.Printf("Error is:\n%s", err)
		wf.FatalError(err)
	}
	// block here and return if no items (secrets) are found
	if len(result) <= 0 {
		log.Println("No items found.")
		return nil
	}
	// unmarshall json
	singleString := strings.Join(result, " ")
	var items []Item
	err = transformToItem(singleString, &items)
	if err != nil {
		log.Println(err)
	}
	if wf.Debug() {
		log.Printf("Bitwarden number of lines of returned data are: %d\n", len(result))
		log.Println("Found ", len(items), " items.")
		for _, item := range items {
			log.Println("Name: ", item.Name, ", Id: ", item.Id)
		}
	}
	return items
}

func runGetItem() {
	wf.Configure(aw.TextErrors(true))
	if opts.Id == "" {
		wf.Fatal("No id sent.")
		return
	}
	id := opts.Id
	log.Println("Id is: ", id)
	jsonPath := ""
	if opts.Query != "" {
		jsonPath = opts.Query
		log.Println("Query is: ", jsonPath)
	}
	totp := opts.Totp
	attachment := opts.Attachment
	// first check if Bitwarden is logged in and locked
	loginErr, unlockErr := BitwardenAuthChecks()
	if loginErr != nil {
		searchAlfred(fmt.Sprintf("%s login", BWAUTH_KEYWORD))
		wf.Fatal(NOT_LOGGED_IN_MSG)
		return
	}
	if unlockErr != nil {
		searchAlfred(fmt.Sprintf("%s unlock", BWAUTH_KEYWORD))
		wf.Fatal(NOT_UNLOCKED_MSG)
		return
	}

	// get a token
	wf.Configure(aw.TextErrors(true))
	token, err := alfred.GetToken(wf)
	if err != nil {
		wf.Fatal("Get Token error")
		return
	}

	message := "Failed to get Bitwarden item."
	args := fmt.Sprintf("%s get item %s --pretty --session %s", BwExec, id, token)
	if totp {
		args = fmt.Sprintf("%s get totp %s --session %s", BwExec, id, token)
	} else if attachment != "" {
		args = fmt.Sprintf("%s get attachment %s --itemid %s --output %s --session %s --raw", BwExec, attachment, id, outputFolder, token)
	}
	log.Println("Read item ", id)

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

	receivedItem := ""
	if jsonPath != "" {
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
			log.Println(fmt.Sprintf("Received key is: %s*", receivedItem[0:2]))
		}
	} else {
		receivedItem = strings.Join(result, " ")
	}
	fmt.Print(receivedItem)
}

func runGetFolders(token string) []Folder {
	message := "Failed to get Bitwarden Folders."
	args := fmt.Sprintf("%s list folders --pretty --session %s", BwExec, token)
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
	err = transformToItem(singleString, &folders)
	if err != nil {
		log.Println(err)
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
	email, _, _, _ := getConfigs(wf)
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", BWCONF_KEYWORD))
		wf.Fatal("No email configured.")
	}

	// first check if Bitwarden is logged in and locked
	loginErr, unlockErr := BitwardenAuthChecks()
	if loginErr != nil {
		searchAlfred(fmt.Sprintf("%s login", BWAUTH_KEYWORD))
		wf.Fatal(NOT_LOGGED_IN_MSG)
		return

	}
	if unlockErr == nil {
		searchAlfred(BW_KEYWORD)
		wf.Fatal("Already unlocked")
		return

	}

	inputScriptPassword := fmt.Sprintf("osascript bitwarden-js-pw-promot.js Unlock %s password true", email)
	message := "Failed to get Password to Unlock."
	// user needs to input pasword
	passwordReturn, err := runCmd(inputScriptPassword, message)
	if err != nil {
		wf.FatalError(err)
	}
	// set the password from the returned slice
	password := ""
	if len(passwordReturn) > 0 {
		password = passwordReturn[0]
	} else {
		wf.Fatal("No Password returned.")
	}
	// remove newline characters
	password = strings.TrimRight(password, "\r\n")
	printOutput([]byte(fmt.Sprintf("first few chars of the password is %s", password[0:2])))

	// Unlock Bitwarden now
	message = "Unlocking Bitwarden failed."
	args := fmt.Sprintf("%s unlock --raw %s", BwExec, password)
	tokenReturn, err := runCmd(args, message)
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
	printError(fmt.Errorf("first few chars of the token is %s", token[0:2]))
	searchAlfred(BW_KEYWORD)
	fmt.Println("Unlocked")
}

// Login to Bitwarden
func runLogin() {
	wf.Configure(aw.TextErrors(true))
	email, sfa, sfaMode, _ := getConfigs(wf)
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", BWCONF_KEYWORD))
		printError(fmt.Errorf("Email missing. Bitwarden not configured yet"))
		wf.Fatal("No email configured.")
	}

	loginErr, unlockErr := BitwardenAuthChecks()
	if loginErr == nil {
		if unlockErr != nil {
			searchAlfred(fmt.Sprintf("%s unlock", BWAUTH_KEYWORD))
			printError(fmt.Errorf("Already logged in but locked."))
			wf.Fatal("Already logged in but locked")
			return

		} else {
			searchAlfred(BW_KEYWORD)
			printError(fmt.Errorf("Already logged in and unlocked."))
			wf.Fatal("Already logged in and unlocked.")
		}
	}

	inputScriptPassword := fmt.Sprintf("osascript bitwarden-js-pw-promot.js Login %s password true", email)
	message := "Failed to get Password to Login."
	passwordReturn, err := runCmd(inputScriptPassword, message)
	if err != nil {
		wf.FatalError(err)
	}
	// set the password from the returned slice
	password := ""
	if len(passwordReturn) > 0 {
		password = passwordReturn[0]
	}

	printOutput([]byte(fmt.Sprintf("first few chars of the password is %s", password[0:2])))
	password = strings.TrimRight(password, "\r\n")

	args := fmt.Sprintf("%s login %s %s", BwExec, email, password)
	if sfa == "true" {
		display2faMode := map2faMode(sfaMode)
		inputScript2faCode := fmt.Sprintf("osascript bitwarden-js-pw-promot.js Login %s %s false", email, display2faMode)
		message := "Failed to get 2FA code to Login."
		sfacodeReturn, err := runCmd(inputScript2faCode, message)
		sfaCode := ""
		if len(sfacodeReturn) > 0 {
			sfaCode = sfacodeReturn[0]
		} else {
			wf.Fatal("No 2FA code returned.")
		}

		if err != nil {
			wf.Fatalf("Error reading password, %s", err)
		}
		args = fmt.Sprintf("%s login %s %s --raw --method %s --code %s", BwExec, email, password, sfaMode, sfaCode)
	}

	message = "Login to Bitwarden failed."
	tokenReturn, err := runCmd(args, message)
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
	printError(fmt.Errorf("first few chars of the token is %s", token[0:2]))
	searchAlfred(BW_KEYWORD)
	fmt.Println("Logged In.")
}

// Logout from Bitwarden
func runLogout() {
	wf.Configure(aw.TextErrors(true))

	err := alfred.RemoveToken(wf)
	if err != nil {
		log.Println(err)
	}

	args := fmt.Sprintf("%s logout", BwExec)

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
	log.Println("Clearing items cache.")
	err := wf.Cache.StoreJSON(CACHE_NAME, nil)
	if err != nil {
		log.Println(err)
	}
	err = wf.Cache.StoreJSON(FOLDER_CACHE_NAME, nil)
	if err != nil {
		log.Println(err)
	}
	err = wf.Cache.StoreJSON(AUTO_FETCH_CACHE, nil)
	if err != nil {
		log.Println(err)
	}

	wf.Configure(aw.TextErrors(true))
	email, _, _, _ := getConfigs(wf)
	if email == "" {
		searchAlfred(fmt.Sprintf("%s email", BWCONF_KEYWORD))
		wf.Fatal("No email configured.")
	}
	loginErr, unlockErr := BitwardenAuthChecks()
	if loginErr != nil {
		fmt.Println(NOT_LOGGED_IN_MSG)
		searchAlfred(fmt.Sprintf("%s login", BWAUTH_KEYWORD))
		return
	}
	if unlockErr != nil {
		fmt.Println(NOT_UNLOCKED_MSG)
		searchAlfred(fmt.Sprintf("%s unlock", BWAUTH_KEYWORD))
		return
	}

	log.Println("Background?", opts.Background)
	if opts.Background {
		if !wf.IsRunning("cache") {
			cmd := exec.Command(os.Args[0], "-cache")
			log.Println("Cache cmd: ", cmd)
			if err := wf.RunInBackground("cache", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Printf("Cache job already running.")
		}
		searchAlfred(BW_KEYWORD)
		return

	}
	log.Println("Running cache")
	getItems()
}
