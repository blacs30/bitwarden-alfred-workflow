// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"fmt"
	aw "github.com/deanishe/awgo"
	"log"
)

// Workflow icons
var (
	iconHelp              = &aw.Icon{Value: "icons/help.png"}
	iconIssue             = &aw.Icon{Value: "icons/issue.png"}
	iconLoading           = &aw.Icon{Value: "icons/loading.png"}
	iconReload            = &aw.Icon{Value: "icons/reload.png"}
	iconServer            = &aw.Icon{Value: "icons/server-solid.png"}
	iconUpdateAvailable   = &aw.Icon{Value: "icons/update-available.png"}
	iconUpdateOK          = &aw.Icon{Value: "icons/update-ok.png"}
	iconWarning           = &aw.Icon{Value: "icons/warning.png"}
	iconOn                = &aw.Icon{Value: "icons/on.png"}
	iconOff               = &aw.Icon{Value: "icons/off.png"}
	iconU2f               = &aw.Icon{Value: "icons/u2f.png"}
	iconApp               = &aw.Icon{Value: "icons/app.png"}
	iconEmail             = &aw.Icon{Value: "icons/email.png"}
	iconEmailAt           = &aw.Icon{Value: "icons/at-solid.png"}
	iconDuo               = &aw.Icon{Value: "icons/duo.png"}
	iconYubi              = &aw.Icon{Value: "icons/yubico.png"}
	iconCreditCard        = &aw.Icon{Value: "icons/credit-card-solid.png"}
	iconCreditCardRegular = &aw.Icon{Value: "icons/credit-card-regular.png"}
	iconLink              = &aw.Icon{Value: "icons/link-solid.png"}
	iconFolder            = &aw.Icon{Value: "icons/folder-solid.png"}
	iconFolderOpen        = &aw.Icon{Value: "icons/folder-open-solid.png"}
	iconLevelUp           = &aw.Icon{Value: "icons/level-up-alt-solid.png"}
	iconInfoCircle        = &aw.Icon{Value: "icons/info-circle-solid.png"}
	iconOrg               = &aw.Icon{Value: "icons/warehouse-solid.png"}
	iconHome              = &aw.Icon{Value: "icons/home-solid.png"}
	iconCity              = &aw.Icon{Value: "icons/city-solid.png"}
	iconMap               = &aw.Icon{Value: "icons/map-solid.png"}
	iconUser              = &aw.Icon{Value: "icons/user-solid.png"}
	iconPhone             = &aw.Icon{Value: "icons/phone-solid.png"}
	iconPassword          = &aw.Icon{Value: "icons/key-solid.png"}
	iconList              = &aw.Icon{Value: "icons/list-solid.png"}
	iconNote              = &aw.Icon{Value: "icons/sticky-note-solid.png"}
	iconStar              = &aw.Icon{Value: "icons/star-solid.png"}
	iconBars              = &aw.Icon{Value: "icons/bars-solid.png"}
	iconPaperClip         = &aw.Icon{Value: "icons/paperclip-solid.png"}
	iconBoxes             = &aw.Icon{Value: "icons/boxes-solid.png"}
	iconCalDay            = &aw.Icon{Value: "icons/calendar-day-solid.png"}
	iconUserClock         = &aw.Icon{Value: "icons/user-clock-solid.png"}
	iconDate              = &aw.Icon{Value: "icons/calendar-alt-solid.png"}
	iconIdCard            = &aw.Icon{Value: "icons/id-card-solid.png"}
	iconIdBatch           = &aw.Icon{Value: "icons/id-badge-solid.png"}
	//iconSettings          = &aw.Icon{Value: "icons/settings.png"}
	//iconURL               = &aw.Icon{Value: "icons/url.png"}
	//iconBitwarden         = &aw.Icon{Value: "icons/bitwarden.png"}
	//iconGlobe             = &aw.Icon{Value: "icons/globe-americas-solid.png"}
)

func init() {
	aw.IconWarning = iconWarning
}

// ReloadIcon returns a spinner icon. It rotates by 15 deg on every
// subsequent call. Use with wf.Reload(0.1) to implement an animated
// spinner.
func ReloadIcon() *aw.Icon {
	var (
		step    = 15
		max     = (45 / step) - 1
		current = wf.Config.GetInt("RELOAD_PROGRESS", 0)
		next    = current + 1
	)
	if next > max {
		next = 0
	}

	log.Printf("progress: current=%d, next=%d", current, next)

	wf.Var("RELOAD_PROGRESS", fmt.Sprintf("%d", next))

	if current == 0 {
		return iconLoading
	}

	return &aw.Icon{Value: fmt.Sprintf("icons/loading-%d.png", current*step)}
}
