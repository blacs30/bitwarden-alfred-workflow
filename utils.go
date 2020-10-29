// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"
	aw "github.com/deanishe/awgo"
	"github.com/go-cmd/cmd"
	"log"
	"os"
	"os/exec"
	"strings"
)

func transformToItem(input string, target interface{}) error {
	err := json.Unmarshal([]byte(input), &target)
	if err != nil {
		return fmt.Errorf("Failed to unmarshall body. Err: %s", err)
	}
	return nil
}

func checkReturn(status cmd.Status, message string) ([]string, error) {
	exitCode := status.Exit
	if exitCode == 127 {
		if wf.Debug() {
			log.Printf("[ERROR] ==> Exit code 127. %q not found in path %q\n", conf.BwExec, os.Getenv("PATH"))
		}
		return []string{}, fmt.Errorf("%q not found in path %q\n", conf.BwExec, os.Getenv("PATH"))
	} else if exitCode == 126 {
		if wf.Debug() {
			log.Printf("[ERROR] ==> Exit code 126. %q has wrong permissions. Must be executable.\n", conf.BwExec)
		}
		return []string{}, fmt.Errorf("%q has wrong permissions. Must be executable.\n", conf.BwExec)
	} else if exitCode == 1 {
		if wf.Debug() {
			log.Println("[ERROR] ==> ", status.Stderr)
		}
		for _, stderr := range status.Stderr {
			if strings.Contains(stderr, "User cancelled.") {
				if wf.Debug() {
					log.Println("[ERROR] ==> ", stderr)
				}
				return []string{}, fmt.Errorf("User cancelled.")
			}
		}
		errorString := strings.Join(status.Stderr[:], "")
		if wf.Debug() {
			log.Printf("[ERROR] ==> Exit code 1. %s Err: %s\n", message, errorString)
		}
		return []string{}, fmt.Errorf(fmt.Sprintf("%s Error:\n%s", message, errorString))
	} else if exitCode == 0 {
		return status.Stdout, nil
	} else {
		if wf.Debug() {
			log.Println("[DEBUG] Unexpected exit code: => ", exitCode)
			// Print each line of STDOUT and STDERR from Cmd
			for _, line := range status.Stdout {
				log.Println("[DEBUG] Stdout: => ", line)
			}
			for _, line := range status.Stderr {
				log.Println("[DEBUG] Stderr: => ", line)
			}
		}
		return []string{}, fmt.Errorf("Unexpected error. Exit code %d.", exitCode)
	}
}

func runCmd(args string, message string) ([]string, error) {
	// Start a long-running process, capture stdout and stderr
	argSet := strings.Fields(args)
	runCmd := cmd.NewCmd(argSet[0], argSet[1:]...)
	status := <-runCmd.Start()

	return checkReturn(status, message)
}

func searchAlfred(search string) {
	// Open Alfred
	log.Println("Search called with argument ", search)
	a := aw.NewAlfred()
	err := a.Search(search)
	if err != nil {
		log.Println(err)
	}
}

func getItemsInFolderCount(folderId string, items []Item) int {
	counter := 0
	for _, item := range items {
		if item.FolderId == folderId {
			counter += 1
		}
	}
	return counter
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func map2faMode(mode int) string {
	switch mode {
	case 0:
		return "Authenticator-app"
	case 1:
		return "Email"
	case 2:
		return "Duo"
	case 3:
		return "YubiKey"
	case 4:
		return "U2F"
	}
	return " "
}

func typeName(typeInt int) string {
	switch typeInt {
	case 1:
		return "Login"
	case 2:
		return "SecureNote"
	case 3:
		return "Card"
	case 4:
		return "Identity"
	}
	return "Type Name Not Found"
}

func clearCache() error {
	log.Println("Clearing items cache.")
	err := wf.Cache.StoreJSON(CACHE_NAME, nil)
	if err != nil {
		return err
	}
	err = wf.Cache.StoreJSON(FOLDER_CACHE_NAME, nil)
	if err != nil {
		return err
	}
	err = wf.Cache.StoreJSON(AUTO_FETCH_CACHE, nil)
	if err != nil {
		return err
	}
	return nil
}
