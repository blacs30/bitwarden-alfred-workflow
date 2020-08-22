// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"github.com/blacs30/bitwarden-alfred-workflow/alfred"
	"github.com/kelseyhightower/envconfig"
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
)

type config struct {
	// From workflow environment variables
	AutoFetchIconCacheAge    int `default:"1440" split_words:"true"`
	AutoFetchIconMaxCacheAge time.Duration
	BwconfKeyword            string `split_words:"true"`
	BwauthKeyword            string `split_words:"true"`
	BwKeyword                string `split_words:"true"`
	BwExec                   string `split_words:"true"`
	CacheAge                 int    `default:"1440" split_words:"true"`
	Email                    string
	EmptyDetailResults       bool `default:"false" split_words:"true"`
	IconCacheAge             int  `default:"43200" split_words:"true"`
	IconCacheEnabled         bool `default:"true" split_words:"true"`
	IconMaxCacheAge          time.Duration
	MaxResults               int `default:"1000" split_words:"true"`
	MaxCacheAge              time.Duration
	Mod1                     string `envconfig:"MODIFIER_1" default:"alt"`
	Mod2                     string `envconfig:"MODIFIER_2" default:"shift"`
	Mod3                     string `envconfig:"MODIFIER_3" default:"cmd"`
	Mod4                     string `envconfig:"MODIFIER_4" default:"cmd,alt,ctrl"`
	OutputFolder             string `default:"" split_words:"true"`
	ReorderingDisabled       bool   `default:"true" split_words:"true"`
	Server                   string `envconfig:"SERVER_URL" default:"https://bitwarden.com"`
	Sfa                      bool   `envconfig:"2FA_ENABLED" default:"true"`
	SfaMode                  int    `envconfig:"2FA_MODE" default:"0"`
	SyncCacheAge             int    `default:"1440" split_words:"true"`
	SyncMaxCacheAge          time.Duration
}

//// Load configuration file.
func loadConfig() {
	err := envconfig.Process("", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	conf.Email = alfred.GetEmail(wf, conf.Email)
	conf.OutputFolder = alfred.GetOutputFolder(wf, conf.OutputFolder)

	cacheAgeDuration := time.Duration(conf.CacheAge)
	conf.MaxCacheAge = cacheAgeDuration * time.Minute
	iconCacheAgeDuration := time.Duration(conf.IconCacheAge)
	conf.IconMaxCacheAge = iconCacheAgeDuration * time.Minute
	autoFetchIconCacheAgeDuration := time.Duration(conf.AutoFetchIconCacheAge)
	conf.AutoFetchIconMaxCacheAge = autoFetchIconCacheAgeDuration * time.Minute
	syncCacheAgeDuration := time.Duration(conf.SyncCacheAge)
	conf.SyncMaxCacheAge = syncCacheAgeDuration * time.Minute
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
