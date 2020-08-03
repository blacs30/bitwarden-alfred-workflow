// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"strings"
	"time"

	"github.com/blacs30/bitwarden-alfred-workflow/alfred"

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
	// DefaultLocateInterval is how often to run locate
	// TODO: create config object
//	DefaultSyncInterval = 24 * time.Hour
//
//	defaultConfig = `# How long to cache the list of secret keys (not the values) for.
//# default: 24h
//#
//# cache-age = "24h"
//`
)

var (
	//conf          *config
	maxCacheAge              = 1440 * time.Minute      // 1 day
	iconMaxCacheAge          = 30 * 1440 * time.Minute // 30 days
	autoFetchIconMaxCacheAge = 1440 * time.Minute      // 1 days
	syncMaxCacheAge          = 1440 * time.Minute      // 1 days
	outputFolder             = ""
	mod1                     []aw.ModKey
	mod1Emoji                string
	mod2                     []aw.ModKey
	mod2Emoji                string
	mod3                     []aw.ModKey
	mod3Emoji                string
	mod4                     []aw.ModKey
	mod4Emoji                string
)

//type config struct {
//    // From workflow environment variables
//    FindInterval time.Duration `env:"CACHE_AGE"`
//}

//// Load configuration file.
//func loadConfig() (*config, error) {
//
//    defer util.Timed(time.Now(), "load config")
//
//    // Load workflow variables
//    if err := wf.Config.To(conf); err != nil {
//        return nil, err
//    }
//
//    // Update depths
//    if conf.FindInterval == 0 {
//        conf.FindInterval = DefaultSyncInterval
//    }
//
//    return conf, nil
//}

func initConfig() {
	BWCONF_KEYWORD = alfred.GetBwConfKeyword(wf)
	BWAUTH_KEYWORD = alfred.GetBwauthKeyword(wf)
	BW_KEYWORD = alfred.GetBwKeyword(wf)
	BwExec = alfred.GetBwExec(wf)

	// get cache config
	cacheAge := alfred.GetCacheTimeout(wf)
	cacheAgeDuration := time.Duration(cacheAge)
	maxCacheAge = cacheAgeDuration * time.Minute
	iconCacheAge := alfred.GetIconCacheTimeout(wf)
	iconCacheAgeDuration := time.Duration(iconCacheAge)
	iconMaxCacheAge = iconCacheAgeDuration * time.Minute
	autoFetchIconCacheAge := alfred.GetAutoFetchIconCacheTimeout(wf)
	autoFetchIconCacheAgeDuration := time.Duration(autoFetchIconCacheAge)
	autoFetchIconMaxCacheAge = autoFetchIconCacheAgeDuration * time.Minute
	syncCacheAge := alfred.GetSyncCacheTimeout(wf)
	syncCacheAgeDuration := time.Duration(syncCacheAge)
	syncMaxCacheAge = syncCacheAgeDuration * time.Minute
	outputFolder = alfred.GetOutputFolder(wf)
	initModifiers()
}

func initModifiers() {
	Modifier_1 := alfred.GetMod1(wf)
	Modifier_2 := alfred.GetMod2(wf)
	Modifier_3 := alfred.GetMod3(wf)
	Modifier_4 := alfred.GetMod4(wf)
	mod1 = getModifierKey(Modifier_1)
	mod1Emoji = getModifierEmoji(Modifier_1)
	mod2 = getModifierKey(Modifier_2)
	mod2Emoji = getModifierEmoji(Modifier_2)
	mod3 = getModifierKey(Modifier_3)
	mod3Emoji = getModifierEmoji(Modifier_3)
	mod4 = getModifierKey(Modifier_4)
	mod4Emoji = getModifierEmoji(Modifier_4)
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
