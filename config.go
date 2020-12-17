// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"
	"github.com/blacs30/bitwarden-alfred-workflow/alfred"
	"github.com/kelseyhightower/envconfig"
	"os"
	"strings"
	"time"

	"log"

	aw "github.com/deanishe/awgo"
)

// Valid modifier keys used to specify alternate actions in Script Filters.
const (
	ModCmd   aw.ModKey = "cmd"   // Alternate action for ⌘↩
	ModAlt   aw.ModKey = "alt"   // Alternate action for ⌥↩
	ModOpt   aw.ModKey = "alt"   // Synonym for ModAlt
	ModCtrl  aw.ModKey = "ctrl"  // Alternate action for ^↩
	ModShift aw.ModKey = "shift" // Alternate action for ⇧↩
	ModFn    aw.ModKey = "fn"    // Alternate action for fn↩
)

var (
	conf      config
	mod1      []aw.ModKey
	mod1Emoji string
	mod2      []aw.ModKey
	mod2Emoji string
	mod3      []aw.ModKey
	mod3Emoji string
	mod4      []aw.ModKey
	mod4Emoji string
	bwData    BwData
)

type modifierActionContent struct {
	Title        string
	Subtitle     string
	Notification string
	Action       string
	Action2      string
	Action3      string
	Arg          string
	Icon         *aw.Icon
	ActionName   string
}

type modifierActionRelation struct {
	ModKey  []aw.ModKey
	Content modifierActionContent
}

type itemActions struct {
	NoMod modifierActionRelation
	Mod1  modifierActionRelation
	Mod2  modifierActionRelation
	Mod3  modifierActionRelation
	Mod4  modifierActionRelation
}

type itemsModifierActionRelation struct {
	Item1 itemActions
	Item2 itemActions
	Item3 itemActions
	Item4 itemActions
	More  modifierActionRelation
}

type config struct {
	// From workflow environment variables
	AutoFetchIconCacheAge    int `default:"1440" split_words:"true"`
	AutoFetchIconMaxCacheAge time.Duration
	BwconfKeyword            string
	BwauthKeyword            string
	BwKeyword                string
	BwfKeyword               string
	BwExec                   string `split_words:"true"`
	// BwDataPath default is set in loadBitwardenJSON()
	BwDataPath         string `envconfig:"BW_DATA_PATH"`
	CacheAge           int    `default:"1440" split_words:"true"`
	Debug              bool   `envconfig:"DEBUG" default:"false"`
	Email              string
	EmptyDetailResults bool `default:"false" split_words:"true"`
	IconCacheAge       int  `default:"43200" split_words:"true"`
	IconCacheEnabled   bool `default:"true" split_words:"true"`
	IconMaxCacheAge    time.Duration
	MaxResults         int `default:"1000" split_words:"true"`
	MaxCacheAge        time.Duration
	Mod1               string `envconfig:"MODIFIER_1" default:"alt"`
	Mod1Action         string `envconfig:"MODIFIER_1_ACTION" default:"username,code"`
	Mod2               string `envconfig:"MODIFIER_2" default:"shift"`
	Mod2Action         string `envconfig:"MODIFIER_2_ACTION" default:"url"`
	Mod3               string `envconfig:"MODIFIER_3" default:"cmd"`
	Mod3Action         string `envconfig:"MODIFIER_3_ACTION" default:"totp"`
	Mod4               string `envconfig:"MODIFIER_4" default:"cmd,alt,ctrl"`
	Mod4Action         string `envconfig:"MODIFIER_4_ACTION" default:"more"`
	NoModAction        string `envconfig:"NO_MODIFIER_ACTION" default:"password,card"`
	OutputFolder       string `default:"" split_words:"true"`
	Path               string
	ReorderingDisabled bool   `default:"true" split_words:"true"`
	Server             string `envconfig:"SERVER_URL" default:"https://bitwarden.com"`
	Sfa                bool   `envconfig:"2FA_ENABLED" default:"true"`
	SfaMode            int    `envconfig:"2FA_MODE" default:"0"`
	SyncCacheAge       int    `default:"1440" split_words:"true"`
	SyncMaxCacheAge    time.Duration
	TitleWithUser      bool `envconfig:"TITLE_WITH_USER" default:"false"`
}

type BwData struct {
	path             string
	InstalledVersion string                 `json:"installedVersion"`
	UserEmail        string                 `json:"userEmail"`
	UserId           string                 `json:"userId"`
	AccessToken      string                 `json:"accessToken"`
	RefreshToken     string                 `json:"refreshToken"`
	ProtectedKey     string                 `json:"__PROTECTED__key"`
	KeyHash          string                 `json:"keyHash"`
	EncKey           string                 `json:"encKey"`
	EncPrivateKey    string                 `json:"encPrivateKey"`
	SecurityStamp    string                 `json:"securityStamp"`
	EncOrgKeys       map[string]interface{} `json:"encOrgKeys"`
	Kdf              int                    `json:"kdf"`
	KdfIterations    int                    `json:"kdfIterations"`
	Unused           map[string]interface{} `json:"-"`
}

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
	bwData.path = path
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(&bwData); err != nil {
		return err
	}
	return nil
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

	conf.Email = alfred.GetEmail(wf, conf.Email, bwData.UserEmail)
	conf.OutputFolder = alfred.GetOutputFolder(wf, conf.OutputFolder)

	// Set a few cache timeout durations
	setItemCacheAge := conf.CacheAge
	if conf.CacheAge < 30 && conf.CacheAge != 0 {
		setItemCacheAge = 30
	}
	cacheAgeDuration := time.Duration(setItemCacheAge)
	conf.MaxCacheAge = cacheAgeDuration * time.Minute

	iconCacheAgeDuration := time.Duration(conf.IconCacheAge)
	conf.IconMaxCacheAge = iconCacheAgeDuration * time.Minute

	autoFetchIconCacheAgeDuration := time.Duration(conf.AutoFetchIconCacheAge)
	conf.AutoFetchIconMaxCacheAge = autoFetchIconCacheAgeDuration * time.Minute

	// if SYNC_CACHE_AGE is lower than 30 but not 0 set to 30
	setSyncCacheAge := conf.SyncCacheAge
	if conf.SyncCacheAge < 30 && conf.SyncCacheAge != 0 {
		setSyncCacheAge = 30
	}
	syncCacheAgeDuration := time.Duration(setSyncCacheAge)
	conf.SyncMaxCacheAge = syncCacheAgeDuration * time.Minute

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
}

func getModifierKey(keys string) []aw.ModKey {
	items := strings.Split(keys, ",")
	var collectKeys []aw.ModKey
	for _, item := range items {
		switch item {
		case "cmd":
			collectKeys = append(collectKeys, ModCmd)
		case "alt":
			collectKeys = append(collectKeys, ModAlt)
		case "fn":
			collectKeys = append(collectKeys, ModFn)
		case "opt":
			collectKeys = append(collectKeys, ModOpt)
		case "ctrl":
			collectKeys = append(collectKeys, ModCtrl)
		case "shift":
			collectKeys = append(collectKeys, ModShift)
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

func getModifierActionRelations(item Item, itemType string, icon *aw.Icon, totp string, url string) itemsModifierActionRelation {
	var itemModConfig itemsModifierActionRelation
	setModAction(&itemModConfig, item, itemType, "nomod", conf.NoModAction, icon, totp, url)
	setModAction(&itemModConfig, item, itemType, "mod1", conf.Mod1Action, icon, totp, url)
	setModAction(&itemModConfig, item, itemType, "mod2", conf.Mod2Action, icon, totp, url)
	setModAction(&itemModConfig, item, itemType, "mod3", conf.Mod3Action, icon, totp, url)
	setModAction(&itemModConfig, item, itemType, "mod4", conf.Mod4Action, icon, totp, url)
	return itemModConfig
}

func setModAction(itemConfig *itemsModifierActionRelation, item Item, itemType string, modMode string, actionString string, icon *aw.Icon, totp string, url string) {
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
	splitActions := strings.Split(actionString, ",")
	for _, action := range splitActions {
		action = strings.TrimSpace(action)
		if itemType == "item1" {
			title := item.Name
			if conf.TitleWithUser {
				title = fmt.Sprintf("%s - %s", item.Name, item.Login.Username)
			}
			if action == "password" {
				subtitle := "Copy password"
				if modMode == "nomod" {
					subtitle = fmt.Sprintf("↩ or ⇥ copy Password, %s %s, %s %s %s Show more", userEmoji, item.Login.Username, totp, url, moreEmoji)
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
					subtitle = fmt.Sprintf("↩ or ⇥ copy Username, %s Password, %s %s %s Show more", passEmoji, totp, url, moreEmoji)
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
				subtitle := "Copy URL"
				if modMode == "nomod" {
					assignedIcon = icon
					subtitle = fmt.Sprintf("↩ or ⇥ copy URL, %s Password, %s Username %s %s Show more", passEmoji, userEmoji, totp, moreEmoji)
				}
				modItem := modifierActionContent{
					Title:        title,
					Subtitle:     subtitle,
					Notification: " ",
					Action:       "-open",
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
			if action == "card" {
				subtitle := "Copy Card Number"
				if modMode == "nomod" {
					subtitle = fmt.Sprintf("%s, %s, ↩ or ⇥ copy Card Number, %s copy Security Code, %s show more", item.Card.Brand, item.Card.Number, codeEmoji, moreEmoji)
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
					subtitle = fmt.Sprintf("%s, %s, ↩ or ⇥ copy Security Code, %s copy Card Number, %s show more", item.Card.Brand, item.Card.Number, cardEmoji, moreEmoji)
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
	}
}

func setItemMod(itemConfig *itemsModifierActionRelation, content modifierActionContent, itemType string, modMode string) {
	switch itemType {
	case "item1":
		switch modMode {
		case "nomod":
			itemConfig.Item1.NoMod = modifierActionRelation{ModKey: nil, Content: content}
		case "mod1":
			itemConfig.Item1.Mod1 = modifierActionRelation{ModKey: mod1, Content: content}
		case "mod2":
			itemConfig.Item1.Mod2 = modifierActionRelation{ModKey: mod2, Content: content}
		case "mod3":
			itemConfig.Item1.Mod3 = modifierActionRelation{ModKey: mod3, Content: content}
		case "mod4":
			itemConfig.Item1.Mod4 = modifierActionRelation{ModKey: mod4, Content: content}
		}
	case "item2":
		switch modMode {
		case "nomod":
			itemConfig.Item2.NoMod = modifierActionRelation{ModKey: nil, Content: content}
		case "mod1":
			itemConfig.Item2.Mod1 = modifierActionRelation{ModKey: mod1, Content: content}
		case "mod2":
			itemConfig.Item2.Mod2 = modifierActionRelation{ModKey: mod2, Content: content}
		case "mod3":
			itemConfig.Item2.Mod3 = modifierActionRelation{ModKey: mod3, Content: content}
		case "mod4":
			itemConfig.Item2.Mod4 = modifierActionRelation{ModKey: mod4, Content: content}
		}
	case "item3":
		switch modMode {
		case "nomod":
			itemConfig.Item3.NoMod = modifierActionRelation{ModKey: nil, Content: content}
		case "mod1":
			itemConfig.Item3.Mod1 = modifierActionRelation{ModKey: mod1, Content: content}
		case "mod2":
			itemConfig.Item3.Mod2 = modifierActionRelation{ModKey: mod2, Content: content}
		case "mod3":
			itemConfig.Item3.Mod3 = modifierActionRelation{ModKey: mod3, Content: content}
		case "mod4":
			itemConfig.Item3.Mod4 = modifierActionRelation{ModKey: mod4, Content: content}
		}
	case "item4":
		switch modMode {
		case "nomod":
			itemConfig.Item4.NoMod = modifierActionRelation{ModKey: nil, Content: content}
		case "mod1":
			itemConfig.Item4.Mod1 = modifierActionRelation{ModKey: mod1, Content: content}
		case "mod2":
			itemConfig.Item4.Mod2 = modifierActionRelation{ModKey: mod2, Content: content}
		case "mod3":
			itemConfig.Item4.Mod3 = modifierActionRelation{ModKey: mod3, Content: content}
		case "mod4":
			itemConfig.Item4.Mod4 = modifierActionRelation{ModKey: mod4, Content: content}
		}
	}
}
