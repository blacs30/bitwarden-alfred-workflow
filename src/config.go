// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blacs30/bitwarden-alfred-workflow/alfred"
	"github.com/kelseyhightower/envconfig"

	"log"

	aw "github.com/deanishe/awgo"
)

// Valid modifier keys used to specify alternate actions in Script Filters.
var (
	conf      config
	mod1      []string
	mod1Emoji string
	mod2      []string
	mod2Emoji string
	mod3      []string
	mod3Emoji string
	mod4      []string
	mod4Emoji string
	mod5      []string
	mod5Emoji string
	bwData    BwData
)

func loadBitwardenJSON() error {
	bwDataPath := conf.BwDataPath
	if bwDataPath == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		bwDataPath = fmt.Sprintf("%s/Library/Application Support/Bitwarden CLI/data.json", homedir)
		debugLog(fmt.Sprintf("bwDataPath is: %s", bwDataPath))
	}
	if err := loadDataFile(bwDataPath); err != nil {
		return err
	}
	return nil
}

func loadDataFile(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()
	byteData, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	bwConfigData, err := decodeBitwardenDataJson(byteData)
	if err != nil {
		return err
	}
	bwData = bwConfigData
	bwData.path = path
	return nil
}

func decodeBitwardenDataJson(byteData []byte) (BwData, error) {
	// unmarshal and decode towards an interface
	newBwData := BwData{}

	var parsed interface{}
	err := json.Unmarshal(byteData, &parsed)
	if err != nil {
		return newBwData, err
	}

	switch parsed.(type) {
	case map[string]interface{}:
		table := parsed.(map[string]interface{})

		if val, ok := table["userId"]; ok {
			newBwData.UserId = fmt.Sprintf("%s", val)
			if val == nil {
				newBwData.UserId = ""
			}
		} else if val == nil {
			newBwData.UserId = ""
		}
		if val, ok := table["activeUserId"]; ok {
			newBwData.ActiveUserId = fmt.Sprintf("%s", val)
			if val == nil {
				newBwData.ActiveUserId = ""
			}
		} else if val == nil {
			newBwData.ActiveUserId = ""
		}

		if newBwData.ActiveUserId != "" && newBwData.UserId == "" {
			// version 1.21.1 and newer
			newBwData.UserId = newBwData.ActiveUserId

			if val, ok := table[fmt.Sprintf("__PROTECTED__%s_user_auto", newBwData.UserId)]; ok {
				newBwData.ProtectedKey = fmt.Sprintf("%s", val)
			}
			if val, ok := table["global"]; ok {
				globalTable := val.(map[string]interface{})
				if globalVal, ok := globalTable["installedVersion"]; ok {
					newBwData.InstalledVersion = fmt.Sprintf("%s", globalVal)
					newBwData.Global.InstalledVersion = fmt.Sprintf("%s", globalVal)
				}
			}
			if val, ok := table[newBwData.UserId]; ok {
				userTable := val.(map[string]interface{})
				if keyVal, ok := userTable["keys"]; ok {
					if apiKeyVal, ok := keyVal.(map[string]interface{})["apiKeyClientSecret"]; ok {
						newBwData.Keys.ApiKeyClientSecret = fmt.Sprintf("%s", apiKeyVal)
					}
					if cryptSymKeyVal, ok := keyVal.(map[string]interface{})["cryptoSymmetricKey"]; ok {
						if val, ok := cryptSymKeyVal.(map[string]interface{})["encrypted"]; ok {
							newBwData.Keys.CryptoSymmetricKey.Encrypted = fmt.Sprintf("%s", val)
							newBwData.EncKey = newBwData.Keys.CryptoSymmetricKey.Encrypted
						}
					}
					// For Version >= 2023.8.2
					if masterKeyEncryptedUserKeyVal, ok := keyVal.(map[string]interface{})["masterKeyEncryptedUserKey"]; ok {
						newBwData.Keys.CryptoSymmetricKey.Encrypted = fmt.Sprintf("%s", masterKeyEncryptedUserKeyVal)
						newBwData.EncKey = newBwData.Keys.CryptoSymmetricKey.Encrypted
					}
					if privateKeyVal, ok := keyVal.(map[string]interface{})["privateKey"]; ok {
						if val, ok := privateKeyVal.(map[string]interface{})["encrypted"]; ok {
							newBwData.Keys.PrivateKey.Encrypted = fmt.Sprintf("%s", val)
						}
					}
				}
				if tokensVal, ok := userTable["tokens"]; ok {
					if val, ok := tokensVal.(map[string]interface{})["accessToken"]; ok {
						newBwData.Tokens.AccessToken = fmt.Sprintf("%s", val)
					}
				}
				if profileVal, ok := userTable["profile"]; ok {
					if val, ok := profileVal.(map[string]interface{})["everBeenUnlocked"]; ok {
						newBwData.Profile.EverBeenUnlocked, _ = strconv.ParseBool(fmt.Sprintf("%t", val))
					}
					if val, ok := profileVal.(map[string]interface{})["lastSync"]; ok {
						newBwData.Profile.LastSync = fmt.Sprintf("%s", val)
					}
					if val, ok := profileVal.(map[string]interface{})["email"]; ok {
						newBwData.Profile.Email = fmt.Sprintf("%s", val)
						newBwData.UserEmail = newBwData.Profile.Email
					}
					if val, ok := profileVal.(map[string]interface{})["userId"]; ok {
						newBwData.Profile.UserId = fmt.Sprintf("%s", val)
					}
					if val, ok := profileVal.(map[string]interface{})["kdfIterations"]; ok {
						kdfIteractionsFloat64, _ := strconv.ParseFloat(fmt.Sprintf("%f", val), 64)
						newBwData.KdfIterations = int64(kdfIteractionsFloat64)
						newBwData.Profile.KdfIterations = newBwData.KdfIterations
					}
					if val, ok := profileVal.(map[string]interface{})["kdfType"]; ok {
						kdfFloat64, _ := strconv.ParseFloat(fmt.Sprintf("%f", val), 64)
						newBwData.Kdf = int64(kdfFloat64)
						newBwData.Profile.KdfType = newBwData.Kdf
					}
				}
			}

		} else {
			// version 1.21.0 and earlier
			newBwData.InstalledVersion = fmt.Sprintf("%s", table["installedVersion"])
			newBwData.UserEmail = fmt.Sprintf("%s", table["userEmail"])
			newBwData.ProtectedKey = fmt.Sprintf("%s", table["__PROTECTED__key"])
			newBwData.EncKey = fmt.Sprintf("%s", table["encKey"])

			// kdfIterations is the only int/float value
			kdfIteractionsFloat64, _ := strconv.ParseFloat(fmt.Sprintf("%f", table["kdfIterations"]), 64)
			newBwData.KdfIterations = int64(kdfIteractionsFloat64)
			kdfFloat64, _ := strconv.ParseFloat(fmt.Sprintf("%f", table["kdf"]), 64)
			newBwData.Kdf = int64(kdfFloat64)
		}

	default:
		return newBwData, fmt.Errorf("type %T unexpected", parsed)
	}

	return newBwData, nil
}

func loadConfig() {
	// Load workflow vars
	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	// load the bitwarden data.json
	err = loadBitwardenJSON()
	if err != nil {
		log.Print(err.Error())
	}

	debugLog(fmt.Sprintf("BwData config is: %+v", bwData))

	conf.Email = alfred.GetEmail(wf, conf.Email, bwData.UserEmail)
	conf.OutputFolder = alfred.GetOutputFolder(wf, conf.OutputFolder)

	// Set a few cache timeout durations
	iconCacheAgeDuration := time.Duration(conf.IconCacheAge)
	conf.IconMaxCacheAge = iconCacheAgeDuration * time.Minute

	autoFetchIconCacheAgeDuration := time.Duration(conf.AutoFetchIconCacheAge)
	conf.AutoFetchIconMaxCacheAge = autoFetchIconCacheAgeDuration * time.Minute

	conf.BwauthKeyword = os.Getenv("bwauth_keyword")
	conf.BwconfKeyword = os.Getenv("bwconf_keyword")
	conf.BwKeyword = os.Getenv("bw_keyword")
	conf.BwfKeyword = os.Getenv("bwf_keyword")

	initModifiers()
}

func initModifiers() {
	mod1 = getModifierKey(conf.Mod1)
	mod1Emoji = getModifierEmoji(conf.Mod1)
	mod2 = getModifierKey(conf.Mod2)
	mod2Emoji = getModifierEmoji(conf.Mod2)
	mod3 = getModifierKey(conf.Mod3)
	mod3Emoji = getModifierEmoji(conf.Mod3)
	mod4 = getModifierKey(conf.Mod4)
	mod4Emoji = getModifierEmoji(conf.Mod4)
	mod5 = getModifierKey(conf.Mod5)
	mod5Emoji = getModifierEmoji(conf.Mod5)
}

func getModifierKey(keys string) []string {
	items := strings.Split(keys, ",")
	var collectKeys []string
	for _, item := range items {
		switch item {
		case "cmd":
			collectKeys = append(collectKeys, "cmd")
		case "alt":
			collectKeys = append(collectKeys, "alt")
		case "fn":
			collectKeys = append(collectKeys, "fn")
		case "opt":
			collectKeys = append(collectKeys, "alt")
		case "ctrl":
			collectKeys = append(collectKeys, "ctrl")
		case "shift":
			collectKeys = append(collectKeys, "shift")
		}
	}
	return collectKeys
}

func getModifierEmoji(keys string) string {
	items := strings.Split(keys, ",")
	var emojiSlice []string
	for _, item := range items {
		switch item {
		case "cmd":
			emojiSlice = append(emojiSlice, "⌘")
		case "alt":
			emojiSlice = append(emojiSlice, "⌥")
		case "fn":
			emojiSlice = append(emojiSlice, "fn")
		case "opt":
			emojiSlice = append(emojiSlice, "⌥")
		case "ctrl":
			emojiSlice = append(emojiSlice, "ˆ")
		case "shift":
			emojiSlice = append(emojiSlice, "⇧")
		}
	}
	emojiString := strings.Join(emojiSlice, "+")

	return emojiString
}

func getTypeEmoji(itemType string) (string, error) {
	modKeysMap := map[string]string{
		conf.NoModAction: "",
		conf.Mod1Action:  mod1Emoji,
		conf.Mod2Action:  mod2Emoji,
		conf.Mod3Action:  mod3Emoji,
		conf.Mod4Action:  mod4Emoji,
		conf.Mod5Action:  mod5Emoji,
	}
	for keys, emoji := range modKeysMap {
		splitKeys := strings.Split(keys, ",")
		for _, key := range splitKeys {
			key = strings.TrimSpace(key)
			if key == itemType {
				return emoji, nil
			}
		}
	}
	return "", fmt.Errorf("no matching key found for type: %s", itemType)
}

func getModifierActionRelations(itemModConfig itemsModifierActionRelationMap, item Item, itemType string, icon *aw.Icon, totp string, url string) {
	modModes := map[string]string{
		"nomod": conf.NoModAction,
		"mod1":  conf.Mod1Action,
		"mod2":  conf.Mod2Action,
		"mod3":  conf.Mod3Action,
		"mod4":  conf.Mod4Action,
		"mod5":  conf.Mod5Action,
	}
	for modMode, action := range modModes {
		setModAction(itemModConfig, item, itemType, modMode, action, icon, totp, url)
	}
}

func setModAction(itemConfig itemsModifierActionRelationMap, item Item, itemType string, modMode string, actionString string, icon *aw.Icon, totp string, url string) {
	// get emojis assigned to the modification key
	moreEmoji, err := getTypeEmoji("more")
	if err != nil {
		log.Fatal(err.Error())
	}
	codeEmoji, err := getTypeEmoji("code")
	if err != nil {
		log.Fatal(err.Error())
	}
	cardEmoji, err := getTypeEmoji("card")
	if err != nil {
		log.Fatal(err.Error())
	}
	passEmoji, err := getTypeEmoji("password")
	if err != nil {
		log.Fatal(err.Error())
	}
	userEmoji, err := getTypeEmoji("username")
	if err != nil {
		log.Fatal(err.Error())
	}
	webUiEmoji, err := getTypeEmoji("webui")
	if err != nil {
		log.Fatal(err.Error())
	}
	// urlEmoji, err := getTypeEmoji("url")
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	splitActions := strings.Split(actionString, ",")
	for _, action := range splitActions {
		action = strings.TrimSpace(action)
		if itemType == "item1" {
			title := item.Name
			if conf.TitleWithUser {
				title = fmt.Sprintf("%s - %s", item.Name, item.Login.Username)
			}

			var urlList string
			for _, url := range item.Login.Uris {
				urlList = fmt.Sprintf("%s - %s", urlList, url.Uri)
			}
			if conf.TitleWithUrls {
				title = fmt.Sprintf("%s - %s", title, urlList)
			}

			if action == "password" {
				subtitle := "Copy password"
				if modMode == "nomod" {
					subtitle = fmt.Sprintf("↩ or ⇥ copy Password, %s %s, %s %s %s Show more, %s open in WebUI", userEmoji, item.Login.Username, totp, url, moreEmoji, webUiEmoji)
				}
				modItem := modifierActionContent{
					Title:        title,
					Subtitle:     subtitle,
					Notification: fmt.Sprintf("Copy Password for user:\n%s", item.Login.Username),
					Action:       "-getitem",
					Action2:      fmt.Sprintf("-id %s", item.Id),
					Action3:      " ",
					Arg:          "login.password",
					Icon:         icon,
					ActionName:   action,
				}
				setItemMod(itemConfig, modItem, itemType, modMode)
			}
			if action == "username" {
				assignedIcon := iconUser
				subtitle := "Copy Username"
				if modMode == "nomod" {
					assignedIcon = icon
					subtitle = fmt.Sprintf("↩ or ⇥ copy Username, %s Password, %s %s %s Show more, %s open in WebUI", passEmoji, totp, url, moreEmoji, webUiEmoji)
				}
				modItem := modifierActionContent{
					Title:        title,
					Subtitle:     subtitle,
					Notification: fmt.Sprintf("Copy Username:\n%s", item.Login.Username),
					Action:       "output",
					Action2:      " ",
					Action3:      " ",
					Arg:          item.Login.Username,
					Icon:         assignedIcon,
					ActionName:   action,
				}
				setItemMod(itemConfig, modItem, itemType, modMode)
			}
			if action == "url" {
				if len(item.Login.Uris) == 0 {
					continue
				}
				assignedIcon := iconLink
				subtitle := "Open URL"
				loginUrlAction := "-open"
				notification := " "
				if !conf.OpenLoginUrl {
					subtitle = "Copy URL"
					loginUrlAction = "output"
					notification = fmt.Sprintf("Copy url for user:\n%s", item.Login.Username)
				}
				if modMode == "nomod" {
					assignedIcon = icon
					subtitle = fmt.Sprintf("↩ or ⇥ copy URL, %s Password, %s Username %s %s Show more, %s open in WebUI", passEmoji, userEmoji, totp, moreEmoji, webUiEmoji)
				}
				modItem := modifierActionContent{
					Title:        title,
					Subtitle:     subtitle,
					Notification: notification,
					Action:       loginUrlAction,
					Action2:      " ",
					Action3:      " ",
					Arg:          item.Login.Uris[0].Uri,
					Icon:         assignedIcon,
					ActionName:   action,
				}
				setItemMod(itemConfig, modItem, itemType, modMode)
			}
			if action == "totp" {
				if totp == "" {
					continue
				}
				assignedIcon := iconUserClock
				subtitle := "Copy TOTP"
				if modMode == "nomod" {
					assignedIcon = icon
					subtitle = fmt.Sprintf("↩ or ⇥ copy TOTP, %s Password, %s Username %s %s Show more", passEmoji, userEmoji, url, moreEmoji)
				}
				modItem := modifierActionContent{
					Title:        title,
					Subtitle:     subtitle,
					Notification: fmt.Sprintf("Copy TOTP for user:\n%s", item.Login.Username),
					Action:       "-getitem",
					Action2:      "-totp",
					Action3:      fmt.Sprintf("-id %s", item.Id),
					Arg:          " ",
					Icon:         assignedIcon,
					ActionName:   action,
				}
				setItemMod(itemConfig, modItem, itemType, modMode)
			}
		}
		if itemType == "item2" {
			modItem := modifierActionContent{
				Title:        item.Name,
				Subtitle:     fmt.Sprintf("↩ or ⇥ copy Note, %s show more", moreEmoji),
				Notification: "Copy Note",
				Action:       "-getitem",
				Action2:      fmt.Sprintf("-id %s", item.Id),
				Action3:      " ",
				Arg:          "notes",
				Icon:         iconNote,
				ActionName:   "",
			}
			setItemMod(itemConfig, modItem, itemType, "nomod")
		}
		if itemType == "item3" {
			title := item.Name
			if conf.TitleWithUser {
				title = fmt.Sprintf("%s - %s", item.Name, item.Card.Number)
			}

			var urlList string
			for _, url := range item.Login.Uris {
				urlList = fmt.Sprintf("%s - %s", urlList, url.Uri)
			}
			if conf.TitleWithUrls {
				title = fmt.Sprintf("%s - %s", title, urlList)
			}

			if action == "card" {
				subtitle := "Copy Card Number"
				if modMode == "nomod" {
					subtitle = fmt.Sprintf("%s, %s, ↩ or ⇥ copy Card Number, %s copy Security Code, %s show more, %s open in WebUI", item.Card.Brand, item.Card.Number, codeEmoji, moreEmoji, webUiEmoji)
				}
				modItem := modifierActionContent{
					Title:        title,
					Subtitle:     subtitle,
					Notification: fmt.Sprintf("Copied Card %s:\n%s", item.Card.Brand, item.Card.Number),
					Action:       "-getitem",
					Action2:      fmt.Sprintf("-id %s", item.Id),
					Action3:      " ",
					Arg:          "card.number",
					Icon:         iconCreditCard,
					ActionName:   action,
				}
				setItemMod(itemConfig, modItem, itemType, modMode)
			}
			if action == "code" {
				subtitle := "Copy card security code"
				if modMode == "nomod" {
					subtitle = fmt.Sprintf("%s, %s, ↩ or ⇥ copy Security Code, %s copy Card Number, %s show more, %s open in WebUI", item.Card.Brand, item.Card.Number, cardEmoji, moreEmoji, webUiEmoji)
				}
				assignedIcon := iconPassword
				if modMode == "nomod" {
					assignedIcon = iconCreditCard
				}
				modItem := modifierActionContent{
					Title:        title,
					Subtitle:     subtitle,
					Notification: "Copied Card Security Code",
					Action:       "-getitem",
					Action2:      fmt.Sprintf("-id %s", item.Id),
					Action3:      " ",
					Arg:          "card.code",
					Icon:         assignedIcon,
					ActionName:   action,
				}
				setItemMod(itemConfig, modItem, itemType, modMode)
			}
		}
		if itemType == "item4" {
			modItem := modifierActionContent{
				Title:        item.Name,
				Subtitle:     fmt.Sprintf("↩ or ⇥ copy name %s %s, %s show more", item.Identity.FirstName, item.Identity.LastName, moreEmoji),
				Notification: fmt.Sprintf("Copied Identity Name:\n%s %s", item.Identity.FirstName, item.Identity.LastName),
				Action:       "output",
				Action2:      " ",
				Action3:      " ",
				Arg:          " ",
				Icon:         iconIdBatch,
				ActionName:   "",
			}
			setItemMod(itemConfig, modItem, itemType, "nomod")
		}
		if action == "more" {
			modItem := modifierActionContent{
				Title:        item.Name,
				Subtitle:     "Show item",
				Notification: " ",
				Action:       fmt.Sprintf("-id %s", item.Id),
				Action2:      " ",
				Action3:      " ",
				Arg:          " ",
				Icon:         iconList,
				ActionName:   action,
			}
			setItemMod(itemConfig, modItem, itemType, modMode)
		}
		if action == "webui" {
			webUi := "https://vault.bitwarden.com"
			if len(conf.WebUiURL) > 0 {
				webUi = conf.WebUiURL
			}
			subtitle := "Open Bitwarden WebUI"
			if modMode == "nomod" {
				subtitle = fmt.Sprintf("↩ or ⇥ open in Web UI, %s Password, %s Username %s %s Show more", passEmoji, userEmoji, totp, moreEmoji)
			}
			modItem := modifierActionContent{
				Title:        item.Name,
				Subtitle:     subtitle,
				Notification: " ",
				Action:       "-open",
				Action2:      " ",
				Action3:      " ",
				Arg:          fmt.Sprintf("%s/#/vault?itemId=%s", webUi, item.Id),
				Icon:         iconBw,
				ActionName:   action,
			}
			setItemMod(itemConfig, modItem, itemType, modMode)
		}
	}
}

func setItemMod(itemConfig itemsModifierActionRelationMap, content modifierActionContent, itemType string, modMode string) {
	switch modMode {
	case "nomod":
		itemConfig[itemType][modMode] = modifierActionRelation{Keys: nil, Content: content}
	case "mod1":
		itemConfig[itemType][modMode] = modifierActionRelation{Keys: mod1, Content: content}
	case "mod2":
		itemConfig[itemType][modMode] = modifierActionRelation{Keys: mod2, Content: content}
	case "mod3":
		itemConfig[itemType][modMode] = modifierActionRelation{Keys: mod3, Content: content}
	case "mod4":
		itemConfig[itemType][modMode] = modifierActionRelation{Keys: mod4, Content: content}
	case "mod5":
		itemConfig[itemType][modMode] = modifierActionRelation{Keys: mod5, Content: content}
	}
}
