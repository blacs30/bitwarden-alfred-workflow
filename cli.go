// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/blacs30/bitwarden-alfred-workflow/alfred"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/util"
	"log"
	"os"
	"os/exec"
	"strconv"
)

var (
	opts = &options{}
	cli  = flag.NewFlagSet("bitwarden-alfred-workflow", flag.ContinueOnError)
)

// CLI flags
type options struct {
	// Commands
	Search     bool
	Config     bool
	SetConfigs bool
	Auth       bool
	Sfa        bool
	Lock       bool
	Icons      bool
	Folder     bool
	Unlock     bool
	Login      bool
	Logout     bool
	Sync       bool
	Open       bool
	GetItem    bool

	// Options
	Force      bool
	Totp       bool
	Last       bool
	Cache      bool
	Background bool

	// Arguments
	Id         string
	Query      string
	Previous   string
	Attachment string
	Output     string
}

func init() {
	cli.BoolVar(&opts.Search, "search", false, "run a new search with options")
	cli.BoolVar(&opts.Config, "conf", false, "show/filter configuration")
	cli.BoolVar(&opts.SetConfigs, "setconfigs", false, "set configs")
	cli.BoolVar(&opts.Auth, "auth", false, "show/filter auth configuration")
	cli.BoolVar(&opts.Sfa, "sfa", false, "Display 2FA options")
	cli.BoolVar(&opts.Open, "open", false, "open specified file in default app")
	cli.BoolVar(&opts.Lock, "lock", false, "lock Bitwarden")
	cli.BoolVar(&opts.Unlock, "unlock", false, "unlock Bitwarden")
	cli.BoolVar(&opts.Icons, "icons", false, "Get favicons")
	cli.BoolVar(&opts.Folder, "folder", false, "Filter Bitwarden Folders")
	cli.StringVar(&opts.Id, "id", "", "Get item by id")
	cli.StringVar(&opts.Previous, "previous", "", "Save previous search term to go back to it")
	cli.StringVar(&opts.Attachment, "attachment", "", "set attachment id")
	cli.StringVar(&opts.Output, "output", "", "set output folder")
	cli.BoolVar(&opts.Login, "login", false, "login to Bitwarden")
	cli.BoolVar(&opts.Logout, "logout", false, "logout Bitwarden")
	cli.BoolVar(&opts.Sync, "sync", false, "sync secrets")
	cli.BoolVar(&opts.Background, "background", false, "Run job in background")
	cli.BoolVar(&opts.Cache, "cache", false, "refresh workflow cache")
	cli.BoolVar(&opts.Last, "last", false, "last sync")
	cli.BoolVar(&opts.Force, "force", false, "force full sync")
	cli.BoolVar(&opts.Totp, "totp", false, "get totp for item id")
	cli.BoolVar(&opts.GetItem, "getitem", false, "get item and an object of it")

	cli.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: bitwarden-alfred-workflow [options] [arguments]

Alfred workflow to get secrets from Bitwarden.

Usage:
    bitwarden-alfred-workflow [<query>]
    bitwarden-alfred-workflow -conf [<query>]
    bitwarden-alfred-workflow -auth [<query>]
    bitwarden-alfred-workflow -setsfaconfig [<setting>]
    bitwarden-alfred-workflow -sfa [<query>]
    bitwarden-alfred-workflow -open [<query>]
    bitwarden-alfred-workflow -lock
    bitwarden-alfred-workflow -unlock
    bitwarden-alfred-workflow -login
    bitwarden-alfred-workflow -logout
    bitwarden-alfred-workflow -sync [--force|--last]
    bitwarden-alfred-workflow -h|-help

Options:
`)
		// TODO: complete this list

		cli.PrintDefaults()
	}
}

func BitwardenAuthChecks() (loginErr error, unlockErr error) {
	args := fmt.Sprintf("%s login --quiet --check", BwExec)
	if wf.Debug() {
		args = fmt.Sprintf("%s login --check", BwExec)
	}
	_, loginErr = runCmd(args, NOT_LOGGED_IN_MSG)
	if wf.Debug() {
		if loginErr != nil {
			log.Println("[ERROR] ==> ", loginErr)
		}
	}

	noQuiet := "--quiet"
	if wf.Debug() {
		noQuiet = ""
	}
	token, err := alfred.GetToken(wf)
	if err != nil {
		args = fmt.Sprintf("%s unlock %s --check", BwExec, noQuiet)
	} else {
		args = fmt.Sprintf("%s unlock %s --check --session %s", BwExec, noQuiet, token)
	}
	_, unlockErr = runCmd(args, NOT_UNLOCKED_MSG)
	if wf.Debug() {
		if unlockErr != nil {
			log.Println("[ERROR] ==> ", unlockErr)
		}
	}
	return
}

// Filter configuration in Alfred
func runConfig() {

	// prevent Alfred from re-ordering results
	reordering := alfred.GetReorderingDisabled(wf)
	if opts.Query == "" || reordering {
		wf.Configure(aw.SuppressUIDs(true))
	}

	// get current email
	log.Println("Getting email from config.")
	email := alfred.GetEmail(wf)
	server := alfred.GetServer(wf)

	log.Printf("filtering config %q ...", opts.Query)

	wf.NewItem("Enter your Bitwarden Email").
		Subtitle("Configure your Bitwarden login email").
		UID("email").
		Valid(true).
		Icon(iconEmailAt).
		Var("action", "-setconfigs").
		Var("action2", "email").
		Var("notification", fmt.Sprintf("Set Email to: \n%s", opts.Query)).
		Var("title", "Set Email").
		Var("subtitle", fmt.Sprintf("Currently set to: %q (remove \"email\" from the beginning if exist)", email)).
		Arg(opts.Query)

	wf.NewItem("Set Server URL").
		Subtitle("Configure your Bitwarden Server URL (Only for selfhosted Bitwarden needed)").
		UID("server").
		Valid(true).
		Icon(iconServer).
		Var("action", "-setconfigs").
		Var("action2", "server").
		Var("notification", fmt.Sprintf("Set Server to: \n%s", opts.Query)).
		Var("title", "Set Server").
		Var("subtitle", fmt.Sprintf("Currently set to: %q", server)).
		Arg(opts.Query)

	wf.NewItem("Enable or disable 2FA").
		Subtitle("Configure Bitwarden to use or not use 2 Factor Authentication").
		UID("sfa").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-sfa").
		Var("action2", "-id ON/OFF")

	wf.NewItem("Set the 2FA method").
		Subtitle("Configure which 2 Factor Authentication Method you use").
		UID("sfamode").
		Valid(true).
		Icon(iconUserClock).
		Var("action", "-sfa").
		Var("action2", "-id Use")

	if wf.UpdateAvailable() {
		wf.NewItem("Workflow Update Available").
			Subtitle("↩ or ⇥ to install update").
			Valid(false).
			UID("update").
			Autocomplete("workflow:update").
			Icon(iconUpdateAvailable)
	} else {
		wf.NewItem("Workflow Is Up To Date").
			Subtitle("↩ or ⇥ to check for update now").
			Valid(false).
			UID("update").
			Autocomplete("workflow:update").
			Icon(iconUpdateOK)
	}

	wf.NewItem("Delete Workflow cache").
		Subtitle("↩ or ⇥ to clean cached items and icons").
		Valid(false).
		UID("delcache").
		Autocomplete("workflow:delcache").
		Icon(aw.IconTrash)

	wf.NewItem("View Help File").
		Subtitle("Open workflow help in your browser").
		Arg("README.html").
		UID("help").
		Valid(true).
		Icon(iconHelp).
		Var("action", "-open")

	wf.NewItem("Report Issue").
		Subtitle("Open workflow issue tracker in your browser").
		Arg(issueTrackerURL).
		UID("issue").
		Valid(true).
		Icon(iconIssue).
		Var("action", "-open")

	wf.NewItem("Visit Forum Thread").
		Subtitle("Open workflow thread on alfredforum.com in your browser").
		Arg(forumThreadURL).
		UID("forum").
		Valid(true).
		Icon(iconLink).
		Var("action", "-open")

	wf.NewItem("Sync Bitwarden Secrets").
		Subtitle("Sync cached keys of Bitwarden secrets (values, the real secrets, are not cached)").
		Valid(true).
		UID("sync").
		Icon(iconReload).
		Var("action", "-sync").
		Var("action2", "-force").
		Var("notification", "Syncing Bitwarden secrets cache…").
		Arg("-background")

	wf.NewItem("Download/Update Favicon for URLs").
		Subtitle("Downloads favicons for URLs").
		Valid(true).
		UID("icons").
		Icon(iconReload).
		Var("action", "-icons").
		Var("notification", "Downloading Favicons for URLs").
		Arg("-background")

	addRefreshCacheItem()

	wf.NewItem("Get date of last Bitwarden secret sync").
		Subtitle("Show the date when the last sync happened with the Bitwarden server.").
		Valid(true).
		UID("sync").
		Icon(iconCalDay).
		Var("action", "-sync").
		Var("notification", "Getting last sync date.").
		Arg("-last")

	if opts.Query != "" {
		wf.Filter(opts.Query)
	}

	wf.WarnEmpty("No Config Found", "Try a different query?")
	wf.SendFeedback()
}

// Open path/URL
func runOpen() {
	wf.Configure(aw.TextErrors(true))

	var args []string
	args = append(args, opts.Query)

	cmd := exec.Command("/usr/bin/open", args...)
	if _, err := util.RunCmd(cmd); err != nil {
		wf.Fatalf("/usr/bin/open %q: %v", opts.Query, err)
	}
}

// Filter auth config in Alfred
func runAuth() {

	// prevent Alfred from re-ordering results
	if opts.Query == "" {
		wf.Configure(aw.SuppressUIDs(true))
	}
	email, sfa, sfaMode, _ := getConfigs(wf)
	if !sfa {
		sfaMode = -1
	}

	log.Printf("filtering auth config %q ...", opts.Query)

	wf.NewItem("Login to Bitwarden").
		Subtitle("↩ or ⇥ to login now").
		Valid(true).
		UID("login").
		Icon(iconOn).
		Var("action", "-login").
		Arg("login", email, fmt.Sprintf("%d", sfaMode), map2faMode(sfaMode))

	wf.NewItem("Logout").
		Subtitle("Logout from Bitwarden").
		UID("logout").
		Valid(true).
		Icon(iconOff).
		Var("action", "-logout")

	wf.NewItem("Unlock").
		Subtitle("Unlock Bitwarden").
		UID("unlock").
		Valid(true).
		Icon(iconOn).
		Var("action", "-unlock").
		Arg("unlock", email)

	wf.NewItem("Lock").
		Subtitle("Lock Bitwarden").
		UID("lock").
		Valid(true).
		Icon(iconOff).
		Var("action", "-lock")

	if opts.Query != "" {
		wf.Filter(opts.Query)
	}

	wf.WarnEmpty("No Auth Config Found", "Try a different query?")
	wf.SendFeedback()
}

// Logout from Bitwarden
func runSetConfigs() {
	wf.Configure(aw.TextErrors(true))

	if cli.NFlag() > 0 {
		var err error
		mode := cli.Arg(0)
		value := cli.Arg(1)
		switch mode {
		case "email":
			err = alfred.SetEmail(wf, value)
		case "server":
			if value == "" {
				value = "https://bitwarden.com"
			}
			if cli.NArg() > 2 {
				value = cli.Arg(1)
				for i := 2; i < cli.NArg(); i++ {
					value = fmt.Sprintf("%s %s", value, cli.Arg(i))
				}
			}
			command := fmt.Sprintf("%s config server %s", BwExec, value)
			message := fmt.Sprintf("Unable to set Bitwarden server %s", value)
			_, err := runCmd(command, message)

			if err != nil {
				wf.FatalError(err)
			}
			err = alfred.SetServer(wf, value)
			if err != nil {
				wf.FatalError(err)
			}
		case "2fa":
			err = alfred.SetSfa(wf, value)
		case "2famode":
			err = alfred.SetSfaMode(wf, value)
			if err != nil {
				wf.FatalError(err)
			}
			sfaModeValue, err := strconv.Atoi(value)
			if err != nil {
				log.Println(err)
			}
			sfamode := map2faMode(sfaModeValue)
			fmt.Printf("DONE: Set %s to \n%s", mode, sfamode)
			searchAlfred(BWCONF_KEYWORD)
			return
		}
		if err != nil {
			wf.FatalError(err)
		}
		fmt.Printf("DONE: Set %s to: \n%s", mode, value)
		searchAlfred(BWCONF_KEYWORD)
	}
}

// Logout from Bitwarden
func runSfa() {
	wf.Configure(aw.TextErrors(true))

	sfamode := map2faMode(alfred.GetSfaMode(wf))
	sfa := alfred.GetSfa(wf)

	if opts.Id == "Use" {
		wf.NewItem("Use U2F (untested)").
			Subtitle(fmt.Sprintf("Currently set to: %q", sfamode)).
			UID("u2f").
			Valid(true).
			Icon(iconU2f).
			Var("notification", "2FA set to U2F").
			Var("action", "-setconfigs").
			Var("action2", "2famode").
			Arg("4")

		wf.NewItem("Use Yubikey (untested)").
			Subtitle(fmt.Sprintf("Currently set to: %q", sfamode)).
			UID("u2f").
			Valid(true).
			Icon(iconYubi).
			Var("notification", "2FA set to Yubikey").
			Var("action", "-setconfigs").
			Var("action2", "2famode").
			Arg("3")

		wf.NewItem("Use Duo (untested)").
			Subtitle(fmt.Sprintf("Currently set to: %q", sfamode)).
			UID("duo").
			Valid(true).
			Icon(iconDuo).
			Var("notification", "2FA set to Duo").
			Var("action", "-setconfigs").
			Var("action2", "2famode").
			Arg("2")

		wf.NewItem("Use Email").
			Subtitle(fmt.Sprintf("Currently set to: %q", sfamode)).
			UID("email").
			Valid(true).
			Icon(iconEmail).
			Var("notification", "2FA set to Email").
			Var("action", "-setconfigs").
			Var("action2", "2famode").
			Arg("1")

		wf.NewItem("Use Authenticator app").
			Subtitle(fmt.Sprintf("Currently set to: %q", sfamode)).
			UID("email").
			Valid(true).
			Icon(iconApp).
			Var("notification", "2FA set to Authenticator app").
			Var("action", "-setconfigs").
			Var("action2", "2famode").
			Arg("0")
	} else if opts.Id == "ON/OFF" {
		wf.NewItem("ON/OFF: Enable 2FA for Bitwarden").
			Subtitle(fmt.Sprintf("Currently set to: %t", sfa)).
			UID("sfaon").
			Valid(true).
			Icon(iconOn).
			Var("notification", "Enabled 2FA").
			Var("action", "-setconfigs").
			Var("action2", "2fa").
			Arg("true")

		wf.NewItem("ON/OFF: Disable 2FA for Bitwarden").
			Subtitle(fmt.Sprintf("Currently set to: %t", sfa)).
			UID("sfaoff").
			Valid(true).
			Icon(iconOff).
			Var("notification", "Disabled 2FA").
			Var("action", "-setconfigs").
			Var("action2", "2fa").
			Arg("false")
	}

	if opts.Query != "" {
		wf.Filter(opts.Query)
	}

	wf.SendFeedback()
}

// Filter Bitwarden secrets in Alfred
func runSearch(folderSearch bool, itemId string) {
	reordering := alfred.GetReorderingDisabled(wf)
	if reordering {
		wf.Configure(aw.SuppressUIDs(true))
	}

	maxResults := alfred.GetMaxResults
	wf.Configure(aw.MaxResults(maxResults(wf)))

	// Load data
	var items []Item
	var folders []Folder
	if wf.Cache.Exists(CACHE_NAME) && wf.Cache.Exists(FOLDER_CACHE_NAME) {
		data, err := Decrypt()
		if err != nil {
			log.Printf("Error decrypting data: %s", err)
		}
		if err := json.Unmarshal(data, &items); err != nil {
			log.Printf("Couldn't load the items cache, error: %s", err)
		}
		if err != nil {
			log.Printf("Couldn't load the items cache, error: %s", err)
		}
		err = wf.Cache.LoadJSON(FOLDER_CACHE_NAME, &folders)
		if err != nil {
			log.Printf("Couldn't load the folders cache, error: %s", err)
		}
	}

	// Check the sync cache, if it expired or doesn't exist do a sync.
	if wf.Cache.Expired(SYNC_CACH_NAME, syncMaxCacheAge) || !wf.Cache.Exists(SYNC_CACH_NAME) {
		if !wf.IsRunning("sync") {
			cmd := exec.Command(os.Args[0], "-sync", "-force")
			log.Println("Sync cmd: ", cmd)
			if err := wf.RunInBackground("sync", cmd); err != nil {
				wf.FatalError(err)
			}
			searchAlfred(BW_KEYWORD)
			return
		} else {
			log.Printf("Sync job already running.")
		}
		wf.NewItem("Refreshing Bitwarden cache…").
			Icon(ReloadIcon())
		wf.SendFeedback()
		return
	}

	// If the cache has expired, set Rerun (which tells Alfred to re-run the
	// workflow), and start the background update process if it isn't already
	// running.
	if wf.Cache.Expired(CACHE_NAME, maxCacheAge) || wf.Cache.Expired(FOLDER_CACHE_NAME, maxCacheAge) {
		wf.Rerun(0.3)
		if !wf.IsRunning("cache") {
			email, sfa, sfaMode, _ := getConfigs(wf)
			if !sfa {
				sfaMode = -1
			}

			// check if logged in or locked first
			loginErr, unlockErr := BitwardenAuthChecks()
			if loginErr != nil {
				wf.Rerun(0)
				wf.NewItem("Login to Bitwarden").
					Subtitle("↩ or ⇥ to login now").
					Valid(true).
					UID("login").
					Icon(iconOn).
					Var("action", "-login").
					Arg("login", email, fmt.Sprintf("%d", sfaMode), map2faMode(sfaMode))
				wf.NewWarningItem("Not logged in to Bitwarden.", "Need to login first.")
				wf.SendFeedback()
				return
			}
			if unlockErr != nil {
				wf.Rerun(0)
				wf.NewItem("Unlock Bitwarden").
					Subtitle("↩ or ⇥ to unlock now").
					Valid(true).
					UID("unlock").
					Icon(iconOn).
					Var("action", "-unlock").
					Arg("unlock", email)
				wf.NewWarningItem("Bitwarden is locked.", "Need to unlock first.")
				wf.SendFeedback()
				return
			}
			// now we can run the background job
			cmd := exec.Command(os.Args[0], "-cache")
			if err := wf.RunInBackground("cache", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Printf("Cache job already running.")
		}
		//if len(items)+len(folders) == 0 {
		wf.NewItem("Refreshing Bitwarden cache…").
			Icon(ReloadIcon())
		wf.SendFeedback()
		return
		//}
	}

	// If iconcache enabled and the cache is expired (or doesn't exist)
	if alfred.GetIconCacheEnabled(wf) && (wf.Data.Expired(ICON_CACHE_NAME, iconMaxCacheAge) || !wf.Data.Exists(ICON_CACHE_NAME)) {
		getIcon(wf)
	}

	if folderSearch && itemId == "" {
		runSearchFolder(items, folders)
	}

	autoFetchCache := false
	if wf.Cache.Expired(AUTO_FETCH_CACHE, autoFetchIconMaxCacheAge) || !wf.Cache.Exists(AUTO_FETCH_CACHE) {
		autoFetchCache = true
		err := wf.Cache.Store(AUTO_FETCH_CACHE, []byte(string("auto-fetch-cache")))
		if err != nil {
			log.Println(err)
		}
	}

	if itemId != "" && !folderSearch {
		log.Printf(`showing items for id "%s" ...`, itemId)
		// Add item to for itemId
		for _, item := range items {
			if item.Id == itemId {
				addItemDetails(item, opts.Previous, autoFetchCache)

				if opts.Query != "" {
					log.Printf(`searching for "%s" ...`, opts.Query)
					res := wf.Filter(opts.Query)
					for _, r := range res {
						log.Printf("[search] %0.2f %#v", r.Score, r.SortKey)
					}
				}
				wf.SendFeedback()
				return
			}
		}
	}

	if folderSearch && itemId != "" {
		log.Printf(`searching in folder with id "%s" ...`, itemId)
		// Add item to search folders
		wf.NewItem("Back to folder search.").
			Subtitle("Go back.").Valid(true).
			UID("").
			Icon(iconFolder).
			Var("action", "-search").
			Arg(fmt.Sprintf("%s -folder %s", BW_KEYWORD, opts.Previous))

		for _, item := range items {
			if item.FolderId == itemId {
				addItemsToWorkflow(item, autoFetchCache)
			}
			if itemId == "null" {
				if item.FolderId == "" {
					addItemsToWorkflow(item, autoFetchCache)
				}
			}
		}
	}

	if len(items) == 0 && len(folders) == 0 {
		addRefreshCacheItem()
		wf.WarnEmpty("No Secrets Found", "Try a different query or refresh the cache manually.")
	}

	if !folderSearch && itemId == "" {
		// Add item to search folders
		wf.NewItem("Search Folders").
			Subtitle("Find folders and secrets in them.").Valid(true).
			UID("").
			Icon(iconFolder).
			Var("action", "-search").
			Arg(fmt.Sprintf("%s -folder ", BW_KEYWORD))

		log.Printf("Number of items %d", len(items))
		for _, item := range items {
			addItemsToWorkflow(item, autoFetchCache)
		}
	}

	if opts.Query != "" {
		log.Printf(`searching for "%s" ...`, opts.Query)
		res := wf.Filter(opts.Query)
		for _, r := range res {
			log.Printf("[search] %0.2f %#v", r.Score, r.SortKey)
		}
	}
	wf.SendFeedback()
}

// Filter Bitwarden secrets in Alfred
func runSearchFolder(items []Item, folders []Folder) {
	if opts.Query != "" {
		log.Printf(`searching for "%s" ...`, opts.Query)
	}

	wf.NewItem("Back to normal search.").
		Subtitle("Go back one level to the normal search").Valid(true).
		UID("back").
		Icon(iconFolder).
		Var("action", "-search").Arg(BW_KEYWORD)

	log.Printf("Number of folders %d", len(folders))
	for _, folder := range folders {
		itemCount := getItemsInFolderCount(folder.Id, items)
		id := "null"
		if folder.Id != "" {
			id = folder.Id
		}
		if opts.Query != "" {
			wf.NewItem(folder.Name).
				Subtitle(fmt.Sprintf("Number of items: %d", itemCount)).Valid(true).
				UID(folder.Id).
				Icon(iconFolderOpen).
				Var("action", "-folder").
				Var("action2", fmt.Sprintf("-id %s ", id)).
				Var("action3", fmt.Sprintf("-previous %s", opts.Query))
		} else {
			wf.NewItem(folder.Name).
				Subtitle(fmt.Sprintf("Number of items: %d", itemCount)).Valid(true).
				UID(folder.Id).
				Icon(iconFolderOpen).
				Var("action", "-folder").
				Var("action2", fmt.Sprintf("-id %s ", id))
		}
	}

	if opts.Query != "" {
		res := wf.Filter(opts.Query)
		for _, r := range res {
			log.Printf("[search] %0.2f %#v", r.Score, r.SortKey)
		}
	}

	if len(items) == 0 && len(folders) == 0 {
		addRefreshCacheItem()
		wf.WarnEmpty("No Secrets Found", "Try a different query or refresh the cache manually.")
	}
	wf.WarnEmpty("No Folders Found", "Try a different query.")
	wf.SendFeedback()
}
