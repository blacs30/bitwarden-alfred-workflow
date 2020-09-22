// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import (
	"fmt"
	aw "github.com/deanishe/awgo"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func addItemDetails(item Item, autoFetchCache bool) {
	wf.Configure(aw.SuppressUIDs(true))
	wf.NewItem("Back to normal search.").
		Subtitle("Go back one level to the normal search").Valid(true).
		Icon(iconLevelUp).
		Var("action", "-search").
		Arg(conf.BwKeyword).
		Var("notification", "")
	//item.Name
	wf.NewItem(fmt.Sprintf("Detail view for: %s", item.Name)).
		Subtitle("").Valid(false).
		Icon(iconInfoCircle)
	wf.NewItem("Item Id").
		Subtitle(fmt.Sprintf("%q", item.Id)).
		Arg(item.Id).
		Icon(iconInfoCircle).
		Var("notification", fmt.Sprintf("Copied Item Id:\n%q", item.Id)).
		Var("action", "output").Valid(true)
	// item.OrganiztionId
	if conf.EmptyDetailResults || item.OrganizationId != "" {
		wf.NewItem("Organization Id").
			Subtitle(fmt.Sprintf("%q", item.OrganizationId)).
			Arg(item.OrganizationId).
			Icon(iconOrg).
			Var("notification", fmt.Sprintf("Copied Organization Id:\n%q", item.OrganizationId)).
			Var("action", "output").Valid(true)
	}
	// item.FolderId
	if conf.EmptyDetailResults || item.FolderId != "" {
		wf.NewItem("Folder Id").
			Subtitle(fmt.Sprintf("%q", item.FolderId)).
			Arg(item.FolderId).
			Icon(iconFolderOpen).
			Var("notification", fmt.Sprintf("Copied Folder Id:\n%q", item.FolderId)).
			Var("action", "output").Valid(true)
	}
	wf.NewItem("Type").
		Subtitle(fmt.Sprintf("%s (%d)", typeName(item.Type), item.Type)).
		Arg(fmt.Sprintf("%s (%d)", typeName(item.Type), item.Type)).
		Icon(iconList).
		Var("notification", fmt.Sprintf("Copied Item Type:\n%s (%d)", typeName(item.Type), item.Type)).
		Var("action", "output").Valid(true)
	if (conf.EmptyDetailResults && item.Type != 2) || (item.Type != 2 && item.Notes != "") {
		wf.NewItem("Note").
			Subtitle(item.Notes).
			Arg(item.Notes).
			Icon(iconNote).
			Var("notification", fmt.Sprintf("Copied Note:\n%q", item.Notes)).
			Var("action", "output").Valid(true)
	} else if (item.Type == 2 && conf.EmptyDetailResults) || (item.Type == 2 && item.Notes != "") {
		wf.NewItem("Note").
			Subtitle(fmt.Sprintf("Secure note: %s", item.Notes)).
			Icon(iconNote).
			Var("notification", "Copy Note").
			Var("action", "-getitem").
			Var("action2", fmt.Sprintf("-id %s", item.Id)).
			Arg("notes").Valid(true) // used as jsonpath
	}
	if conf.EmptyDetailResults || item.Favorite {
		wf.NewItem("Favorite").
			Subtitle(fmt.Sprintf("%q", strconv.FormatBool(item.Favorite))).
			Arg(strconv.FormatBool(item.Favorite)).
			Icon(iconStar).
			Var("notification", fmt.Sprintf("Copied Favorite:\n%q", strconv.FormatBool(item.Favorite))).
			Var("action", "output").Valid(true)
	}
	// item.Fields
	if len(item.Fields) > 0 {
		for k, field := range item.Fields {
			counter := k + 1
			// it's a secret type so we need to fetch the secret from Bitwarden
			if field.Type == 1 {
				wf.NewItem(fmt.Sprintf("[Field %d] %s", counter, field.Name)).
					Subtitle(fmt.Sprintf("%q", field.Value)).
					Icon(iconBars).
					Var("notification", fmt.Sprintf("Copy secret field:\n%s", field.Name)).
					Var("action", "-getitem").
					Var("action2", fmt.Sprintf("-id %s", item.Id)).
					Arg(fmt.Sprintf("fields[%d].value", k)). // used as jsonpath
					Valid(true)
			} else {
				wf.NewItem(fmt.Sprintf("[Field %d] %s", counter, field.Name)).
					Subtitle(fmt.Sprintf("%q", field.Value)).
					Arg(field.Value).
					Icon(iconBars).
					Var("notification", fmt.Sprintf("Copied field:\n%q", field.Name)).
					Var("action", "output").Valid(true)
			}
		}
	}
	// item.Attachments
	if len(item.Attachments) > 0 {
		for k, att := range item.Attachments {
			counter := k + 1
			// it's a secret type so we need to fetch the secret from Bitwarden
			wf.NewItem(fmt.Sprintf("[Attachment %d] %s", counter, att.FileName)).
				Subtitle(fmt.Sprintf("↩ or ⇥ save Attachment to %s, size %s", conf.OutputFolder, att.SizeName)).
				Icon(iconPaperClip).
				Valid(true).
				Var("notification", fmt.Sprintf("Save attachment to :\n%s%s", conf.OutputFolder, att.FileName)).
				Var("action", "-getitem").
				Var("action2", fmt.Sprintf("-attachment %s", att.Id)).
				Var("action3", fmt.Sprintf("-id %s", item.Id))
		}
	}
	// item.CollectionIds
	if conf.EmptyDetailResults || len(item.CollectionIds) > 0 {
		wf.NewItem("Collection IDs").
			Subtitle(fmt.Sprintf("%q", strings.Join(item.CollectionIds, ","))).
			Arg(fmt.Sprintf("%q", strings.Join(item.CollectionIds, ","))).
			Icon(iconBoxes).
			Var("notification", fmt.Sprintf("Copied Collections IDs:\n%q", fmt.Sprintf("%q", strings.Join(item.CollectionIds, ",")))).
			Var("action", "output").Valid(true)

	}
	// item.RevisionDate
	if conf.EmptyDetailResults || fmt.Sprint(item.RevisionDate) != "" {
		wf.NewItem("Revision Date").
			Subtitle(fmt.Sprintf("%q", item.RevisionDate)).
			Arg(fmt.Sprint(item.RevisionDate)).
			Icon(iconCalDay).
			Var("notification", fmt.Sprintf("Copied RevisionDate:\n%q", item.RevisionDate)).
			Var("action", "output").Valid(true)
	}
	// specifc items for login type
	// item.Type 1
	if item.Type == 1 {
		// get icons from cache
		icon := iconLink
		if len(item.Login.Uris) > 0 && conf.IconCacheEnabled {
			iconPath := fmt.Sprintf("%s/%s/%s.png", wf.DataDir(), "urlicon", item.Id)
			if _, err := os.Stat(iconPath); err != nil {
				log.Println("Couldn't load the cached icon, error: ", err)
				if autoFetchCache {
					log.Println("Getting icons.")
					runGetIcons(item.Login.Uris[0].Uri, item.Id)
				}
			}
			icon = &aw.Icon{Value: iconPath}
		}

		// item.Login.Username
		if conf.EmptyDetailResults || item.Login.Username != "" {
			wf.NewItem("Username").
				Subtitle(fmt.Sprintf("%q", item.Login.Username)).
				Valid(true).
				Arg(item.Login.Username).
				Icon(iconUser).
				Var("action", "output").Valid(true).
				Var("notification", fmt.Sprintf("Copied Username:\n%q", item.Login.Username))
		}
		// item.Login.Uris[*].Uri
		if len(item.Login.Uris) > 0 {
			for k, uri := range item.Login.Uris {
				counter := k + 1
				wf.NewItem(fmt.Sprintf("Url %d", counter)).
					Subtitle(fmt.Sprintf("%q", uri.Uri)).
					Valid(true).
					Arg(uri.Uri).
					Icon(icon).
					Var("action", "-open").Valid(true).
					Var("notification", "")
			}
		}
		// item.Login.Password
		if conf.EmptyDetailResults || item.Login.Password != "" {
			wf.NewItem("Password").
				Subtitle(fmt.Sprintf("%q", item.Login.Password)).
				Valid(true).
				Icon(iconPassword).
				Var("notification", fmt.Sprintf("Copy Password for user:\n%s", item.Login.Username)).
				Var("action", "-getitem").
				Var("action2", fmt.Sprintf("-id %s", item.Id)).
				Arg("login.password") // used as jsonpath
		}
		// TOTP
		if item.Login.Totp != "" {
			wf.NewItem("TOTP").
				Subtitle(fmt.Sprintf("%q", item.Login.Totp)).
				Valid(true).
				Icon(iconUserClock).
				Var("notification", fmt.Sprintf("Copy TOTP for user:\n%s", item.Login.Username)).
				Var("action", "-getitem").
				Var("action2", "-totp").
				Var("action3", fmt.Sprintf("-id %s", item.Id))
		}
		// Password Revision Date
		// check if the set value matches the initial value of time, then we know the passwordRevisionDate hasn't been set by Bitwarden
		d1 := time.Date(0001, 01, 01, 00, 00, 00, 00, time.UTC)
		datesEqual := d1.Equal(item.Login.PasswordRevisionDate)
		if !datesEqual {
			wf.NewItem("Password Revision Date").
				Subtitle(fmt.Sprintf("%q", item.Login.PasswordRevisionDate)).
				Valid(true).
				Icon(iconDate).
				Arg(fmt.Sprint(item.Login.PasswordRevisionDate)).
				Var("action", "output").Valid(true).
				Var("notification", fmt.Sprintf("Copied Password Revision Date:\n%q", fmt.Sprint(item.Login.PasswordRevisionDate)))
		}
	} else if item.Type == 3 {
		if conf.EmptyDetailResults || item.Card.CardHolderName != "" {
			wf.NewItem("Card Holder Name").
				Subtitle(fmt.Sprintf("%q", item.Card.CardHolderName)).
				Valid(true).
				Icon(iconUser).
				Arg(item.Card.CardHolderName).
				Var("notification", fmt.Sprintf("Copied Card Holder Name:\n%s", item.Card.CardHolderName)).
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Card.Number != "" {
			wf.NewItem("Card Number").
				Subtitle(fmt.Sprintf("%q", item.Card.Number)).
				Valid(true).
				Icon(iconCreditCard).
				Var("notification", fmt.Sprintf("Copy Card Number:\n%s", item.Card.Number)).
				Var("action", "-getitem").
				Var("action2", fmt.Sprintf("-id %s", item.Id)).
				Arg("card.number")
		}
		if conf.EmptyDetailResults || item.Card.Code != "" {
			wf.NewItem("Card Security Code").
				Subtitle(fmt.Sprintf("%q", item.Card.Code)).
				Valid(true).
				Icon(iconPassword).
				Var("notification", "Copy Card Security Code.").
				Var("action", "-getitem").
				Var("action2", fmt.Sprintf("-id %s", item.Id)).
				Arg("card.code")
		}
		if conf.EmptyDetailResults || item.Card.Brand != "" {
			wf.NewItem("Card Brand").
				Subtitle(fmt.Sprintf("%q", item.Card.Brand)).
				Valid(true).
				Icon(iconCreditCardRegular).
				Arg(item.Card.Brand).
				Var("notification", fmt.Sprintf("Copied Card Brand:\n%s", item.Card.Brand)).
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Card.ExpMonth != "" {
			wf.NewItem("Expiry Month").
				Subtitle(fmt.Sprintf("%q", item.Card.ExpMonth)).
				Valid(true).
				Icon(iconDate).
				Arg(item.Card.ExpMonth).
				Var("notification", fmt.Sprintf("Copied Card Expiry Month:\n%s", item.Card.ExpMonth)).
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Card.ExpYear != "" {
			wf.NewItem("Expiry Year").
				Subtitle(fmt.Sprintf("%q", item.Card.ExpYear)).
				Valid(true).
				Icon(iconDate).
				Arg(item.Card.ExpYear).
				Var("notification", fmt.Sprintf("Copied Card Expiry Year:\n%s", item.Card.ExpYear)).
				Var("action", "output")
		}
	} else if item.Type == 4 {
		if conf.EmptyDetailResults || item.Identity.Title != "" {
			wf.NewItem("Title").
				Subtitle(fmt.Sprintf("%q", item.Identity.Title)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.Title).
				Var("notification", "Copied title.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.FirstName != "" {
			wf.NewItem("Firstname").
				Subtitle(fmt.Sprintf("%q", item.Identity.FirstName)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.FirstName).
				Var("notification", "Copied fistname.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.MiddleName != "" {
			wf.NewItem("Middlename").
				Subtitle(fmt.Sprintf("%q", item.Identity.MiddleName)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.MiddleName).
				Var("notification", "Copied middlename.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.LastName != "" {
			wf.NewItem("Lastname").
				Subtitle(fmt.Sprintf("%q", item.Identity.LastName)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.LastName).
				Var("notification", "Copied lastname.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Address1 != "" {
			wf.NewItem("Address1").
				Subtitle(fmt.Sprintf("%q", item.Identity.Address1)).
				Valid(true).
				Icon(iconHome).
				Arg(item.Identity.Address1).
				Var("notification", "Copied address1.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Address2 != "" {
			wf.NewItem("Address2").
				Subtitle(fmt.Sprintf("%q", item.Identity.Address2)).
				Valid(true).
				Icon(iconHome).
				Arg(item.Identity.Address2).
				Var("notification", "Copied address2.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Address3 != "" {
			wf.NewItem("Address3").
				Subtitle(fmt.Sprintf("%q", item.Identity.Address3)).
				Valid(true).
				Icon(iconHome).
				Arg(item.Identity.Address3).
				Var("notification", "Copied address3.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.City != "" {
			wf.NewItem("City").
				Subtitle(fmt.Sprintf("%q", item.Identity.City)).
				Valid(true).
				Icon(iconCity).
				Arg(item.Identity.City).
				Var("notification", "Copied city.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.State != "" {
			wf.NewItem("State").
				Subtitle(fmt.Sprintf("%q", item.Identity.State)).
				Valid(true).
				Icon(iconMap).
				Arg(item.Identity.State).
				Var("notification", "Copied state.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.PostalCode != "" {
			wf.NewItem("Postal Code").
				Subtitle(fmt.Sprintf("%q", item.Identity.PostalCode)).
				Valid(true).
				Icon(iconMap).
				Arg(item.Identity.PostalCode).
				Var("notification", "Copied postal code.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Country != "" {
			wf.NewItem("Country").
				Subtitle(fmt.Sprintf("%q", item.Identity.Country)).
				Valid(true).
				Icon(iconMap).
				Arg(item.Identity.Country).
				Var("notification", "Copied country.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Company != "" {
			wf.NewItem("Company").
				Subtitle(fmt.Sprintf("%q", item.Identity.Company)).
				Valid(true).
				Icon(iconOrg).
				Arg(item.Identity.Company).
				Var("notification", "Copied company.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Email != "" {
			wf.NewItem("Email").
				Subtitle(fmt.Sprintf("%q", item.Identity.Email)).
				Valid(true).
				Icon(iconEmailAt).
				Arg(item.Identity.Email).
				Var("notification", "Copied email.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Phone != "" {
			wf.NewItem("Phone").
				Subtitle(fmt.Sprintf("%q", item.Identity.Phone)).
				Valid(true).
				Icon(iconPhone).
				Arg(item.Identity.Phone).
				Var("notification", "Copied phone.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Ssn != "" {
			wf.NewItem("Social Security Number").
				Subtitle(fmt.Sprintf("%q", item.Identity.Ssn)).
				Valid(true).
				Icon(iconIdCard).
				Arg(item.Identity.Ssn).
				Var("notification", "Copied Social Security Number.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.Username != "" {
			wf.NewItem("Username").
				Subtitle(fmt.Sprintf("%q", item.Identity.Username)).
				Valid(true).
				Icon(iconUser).
				Arg(item.Identity.Username).
				Var("notification", "Copied Username.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.PassportNumber != "" {
			wf.NewItem("Passport Number").
				Subtitle(fmt.Sprintf("%q", item.Identity.PassportNumber)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.PassportNumber).
				Var("notification", "Copied Passport Number.").
				Var("action", "output")
		}
		if conf.EmptyDetailResults || item.Identity.LicenseNumber != "" {
			wf.NewItem("License Number").
				Subtitle(fmt.Sprintf("%q", item.Identity.LicenseNumber)).
				Valid(true).
				Icon(iconIdBatch).
				Arg(item.Identity.LicenseNumber).
				Var("notification", "Copied License Number.").
				Var("action", "output")
		}
	}
}

func addItemsToWorkflow(item Item, autoFetchCache bool) {
	if item.Type == 1 {
		icon := iconLink
		if len(item.Login.Uris) > 0 && conf.IconCacheEnabled {
			iconPath := fmt.Sprintf("%s/%s/%s.png", wf.DataDir(), "urlicon", item.Id)
			if _, err := os.Stat(iconPath); err != nil {
				log.Println("Couldn't load the cached icon, error: ", err)
				if autoFetchCache {
					log.Println("Getting icons.")
					runGetIcons(item.Login.Uris[0].Uri, item.Id)
				}
			} else {
				icon = &aw.Icon{Value: iconPath}
			}
		}
		// Get the emoji assigned to a modifier action
		totpEmoji, err := getTypeEmoji("totp")
		if err != nil {
			log.Fatal(err.Error())
		}
		totp := fmt.Sprintf("%s *TOTP, ", totpEmoji)
		if len(item.Login.Totp) == 0 {
			totp = ""
		}
		urlEmoji, err := getTypeEmoji("url")
		if err != nil {
			log.Fatal(err.Error())
		}
		url := fmt.Sprintf("%s URL, ", urlEmoji)
		if len(item.Login.Uris) < 1 {
			url = ""
		}
		itemModSet := getModifierActionRelations(item, "item1", icon, totp, url)
		log.Printf("Item1:\n%+v", itemModSet.Item1)
		it1 := wf.NewItem(itemModSet.Item1.NoMod.Content.Title).
			Subtitle(itemModSet.Item1.NoMod.Content.Subtitle).Valid(true).
			Arg(itemModSet.Item1.NoMod.Content.Arg).
			UID(item.Name).
			Var("notification", itemModSet.Item1.NoMod.Content.Notification).
			Var("action", itemModSet.Item1.NoMod.Content.Action).
			Var("action2", itemModSet.Item1.NoMod.Content.Action2).
			Var("action3", itemModSet.Item1.NoMod.Content.Action3).
			Arg(itemModSet.Item1.NoMod.Content.Arg).
			Icon(itemModSet.Item1.NoMod.Content.Icon)
		if itemModSet.Item1.Mod1.ModKey != nil {
			it1.NewModifier(itemModSet.Item1.Mod1.ModKey[0:]...).
				Subtitle(itemModSet.Item1.Mod1.Content.Subtitle).
				Arg(itemModSet.Item1.Mod1.Content.Arg).
				Var("notification", itemModSet.Item1.Mod1.Content.Notification).
				Var("action", itemModSet.Item1.Mod1.Content.Action).
				Var("action2", itemModSet.Item1.Mod1.Content.Action2).
				Var("action3", itemModSet.Item1.Mod1.Content.Action3).
				Arg(itemModSet.Item1.Mod1.Content.Arg).
				Icon(itemModSet.Item1.Mod1.Content.Icon)
		}
		if itemModSet.Item1.Mod2.ModKey != nil {
			it1.NewModifier(itemModSet.Item1.Mod2.ModKey[0:]...).
				Subtitle(itemModSet.Item1.Mod2.Content.Subtitle).
				Arg(itemModSet.Item1.Mod2.Content.Arg).
				Var("notification", itemModSet.Item1.Mod2.Content.Notification).
				Var("action", itemModSet.Item1.Mod2.Content.Action).
				Var("action2", itemModSet.Item1.Mod2.Content.Action2).
				Var("action3", itemModSet.Item1.Mod2.Content.Action3).
				Arg(itemModSet.Item1.Mod2.Content.Arg).
				Icon(itemModSet.Item1.Mod2.Content.Icon)
		}
		if itemModSet.Item1.Mod3.ModKey != nil {
			it1.NewModifier(itemModSet.Item1.Mod3.ModKey[0:]...).
				Subtitle(itemModSet.Item1.Mod3.Content.Subtitle).
				Arg(itemModSet.Item1.Mod3.Content.Arg).
				Var("notification", itemModSet.Item1.Mod3.Content.Notification).
				Var("action", itemModSet.Item1.Mod3.Content.Action).
				Var("action2", itemModSet.Item1.Mod3.Content.Action2).
				Var("action3", itemModSet.Item1.Mod3.Content.Action3).
				Arg(itemModSet.Item1.Mod3.Content.Arg).
				Icon(itemModSet.Item1.Mod3.Content.Icon)
		}
		if itemModSet.Item1.Mod4.ModKey != nil {
			it1.NewModifier(itemModSet.Item1.Mod4.ModKey[0:]...).
				Subtitle(itemModSet.Item1.Mod4.Content.Subtitle).
				Arg(itemModSet.Item1.Mod4.Content.Arg).
				Var("notification", itemModSet.Item1.Mod4.Content.Notification).
				Var("action", itemModSet.Item1.Mod4.Content.Action).
				Var("action2", itemModSet.Item1.Mod4.Content.Action2).
				Var("action3", itemModSet.Item1.Mod4.Content.Action3).
				Arg(itemModSet.Item1.Mod4.Content.Arg).
				Icon(itemModSet.Item1.Mod4.Content.Icon)
		}
	} else if item.Type == 2 {
		itemModSet := getModifierActionRelations(item, "item2", nil, "", "")
		log.Printf("Item2:\n%+v", itemModSet.Item2)
		it2 := wf.NewItem(itemModSet.Item2.NoMod.Content.Title).
			Subtitle(itemModSet.Item2.NoMod.Content.Subtitle).Valid(true).
			Arg(itemModSet.Item2.NoMod.Content.Arg).
			UID(item.Name).
			Var("notification", itemModSet.Item2.NoMod.Content.Notification).
			Var("action", itemModSet.Item2.NoMod.Content.Action).
			Var("action2", itemModSet.Item2.NoMod.Content.Action2).
			Var("action3", itemModSet.Item2.NoMod.Content.Action3).
			Arg(itemModSet.Item2.NoMod.Content.Arg).
			Icon(itemModSet.Item2.NoMod.Content.Icon)
		if itemModSet.Item2.Mod1.ModKey != nil {
			it2.NewModifier(itemModSet.Item2.Mod1.ModKey[0:]...).
				Subtitle(itemModSet.Item2.Mod1.Content.Subtitle).
				Arg(itemModSet.Item2.Mod1.Content.Arg).
				Var("notification", itemModSet.Item2.Mod1.Content.Notification).
				Var("action", itemModSet.Item2.Mod1.Content.Action).
				Var("action2", itemModSet.Item2.Mod1.Content.Action2).
				Var("action3", itemModSet.Item2.Mod1.Content.Action3).
				Arg(itemModSet.Item2.Mod1.Content.Arg).
				Icon(itemModSet.Item2.Mod1.Content.Icon)
		}
		if itemModSet.Item2.Mod2.ModKey != nil {
			it2.NewModifier(itemModSet.Item2.Mod2.ModKey[0:]...).
				Subtitle(itemModSet.Item2.Mod2.Content.Subtitle).
				Arg(itemModSet.Item2.Mod2.Content.Arg).
				Var("notification", itemModSet.Item2.Mod2.Content.Notification).
				Var("action", itemModSet.Item2.Mod2.Content.Action).
				Var("action2", itemModSet.Item2.Mod2.Content.Action2).
				Var("action3", itemModSet.Item2.Mod2.Content.Action3).
				Arg(itemModSet.Item2.Mod2.Content.Arg).
				Icon(itemModSet.Item2.Mod2.Content.Icon)
		}
		if itemModSet.Item2.Mod3.ModKey != nil {
			it2.NewModifier(itemModSet.Item2.Mod3.ModKey[0:]...).
				Subtitle(itemModSet.Item2.Mod3.Content.Subtitle).
				Arg(itemModSet.Item2.Mod3.Content.Arg).
				Var("notification", itemModSet.Item2.Mod3.Content.Notification).
				Var("action", itemModSet.Item2.Mod3.Content.Action).
				Var("action2", itemModSet.Item2.Mod3.Content.Action2).
				Var("action3", itemModSet.Item2.Mod3.Content.Action3).
				Arg(itemModSet.Item2.Mod3.Content.Arg).
				Icon(itemModSet.Item2.Mod3.Content.Icon)
		}
		if itemModSet.Item2.Mod4.ModKey != nil {
			it2.NewModifier(itemModSet.Item2.Mod4.ModKey[0:]...).
				Subtitle(itemModSet.Item2.Mod4.Content.Subtitle).
				Arg(itemModSet.Item2.Mod4.Content.Arg).
				Var("notification", itemModSet.Item2.Mod4.Content.Notification).
				Var("action", itemModSet.Item2.Mod4.Content.Action).
				Var("action2", itemModSet.Item2.Mod4.Content.Action2).
				Var("action3", itemModSet.Item2.Mod4.Content.Action3).
				Arg(itemModSet.Item2.Mod4.Content.Arg).
				Icon(itemModSet.Item2.Mod4.Content.Icon)
		}
	} else if item.Type == 3 {
		itemModSet := getModifierActionRelations(item, "item3", nil, "", "")
		log.Printf("Item3:\n%+v", itemModSet.Item3)
		it3 := wf.NewItem(itemModSet.Item3.NoMod.Content.Title).
			Valid(true).
			Subtitle(itemModSet.Item3.NoMod.Content.Subtitle).
			Arg(itemModSet.Item3.NoMod.Content.Arg).
			UID(item.Name).
			Var("notification", itemModSet.Item3.NoMod.Content.Notification).
			Var("action", itemModSet.Item3.NoMod.Content.Action).
			Var("action2", itemModSet.Item3.NoMod.Content.Action2).
			Var("action3", itemModSet.Item3.NoMod.Content.Action3).
			Arg(itemModSet.Item3.NoMod.Content.Arg).
			Icon(itemModSet.Item3.NoMod.Content.Icon)
		if itemModSet.Item3.Mod1.ModKey != nil {
			it3.NewModifier(itemModSet.Item3.Mod1.ModKey[0:]...).
				Subtitle(itemModSet.Item3.Mod1.Content.Subtitle).
				Arg(itemModSet.Item3.Mod1.Content.Arg).
				Var("notification", itemModSet.Item3.Mod1.Content.Notification).
				Var("action", itemModSet.Item3.Mod1.Content.Action).
				Var("action2", itemModSet.Item3.Mod1.Content.Action2).
				Var("action3", itemModSet.Item3.Mod1.Content.Action3).
				Arg(itemModSet.Item3.Mod1.Content.Arg).
				Icon(itemModSet.Item3.Mod1.Content.Icon)
		}
		if itemModSet.Item3.Mod2.ModKey != nil {
			it3.NewModifier(itemModSet.Item3.Mod2.ModKey[0:]...).
				Subtitle(itemModSet.Item3.Mod2.Content.Subtitle).
				Arg(itemModSet.Item3.Mod2.Content.Arg).
				Var("notification", itemModSet.Item3.Mod2.Content.Notification).
				Var("action", itemModSet.Item3.Mod2.Content.Action).
				Var("action2", itemModSet.Item3.Mod2.Content.Action2).
				Var("action3", itemModSet.Item3.Mod2.Content.Action3).
				Arg(itemModSet.Item3.Mod2.Content.Arg).
				Icon(itemModSet.Item3.Mod2.Content.Icon)
		}
		if itemModSet.Item3.Mod3.ModKey != nil {
			it3.NewModifier(itemModSet.Item3.Mod3.ModKey[0:]...).
				Subtitle(itemModSet.Item3.Mod3.Content.Subtitle).
				Arg(itemModSet.Item3.Mod3.Content.Arg).
				Var("notification", itemModSet.Item3.Mod3.Content.Notification).
				Var("action", itemModSet.Item3.Mod3.Content.Action).
				Var("action2", itemModSet.Item3.Mod3.Content.Action2).
				Var("action3", itemModSet.Item3.Mod3.Content.Action3).
				Arg(itemModSet.Item3.Mod3.Content.Arg).
				Icon(itemModSet.Item3.Mod3.Content.Icon)
		}
		if itemModSet.Item3.Mod4.ModKey != nil {
			it3.NewModifier(itemModSet.Item3.Mod4.ModKey[0:]...).
				Subtitle(itemModSet.Item3.Mod4.Content.Subtitle).
				Arg(itemModSet.Item3.Mod4.Content.Arg).
				Var("notification", itemModSet.Item3.Mod4.Content.Notification).
				Var("action", itemModSet.Item3.Mod4.Content.Action).
				Var("action2", itemModSet.Item3.Mod4.Content.Action2).
				Var("action3", itemModSet.Item3.Mod4.Content.Action3).
				Arg(itemModSet.Item3.Mod4.Content.Arg).
				Icon(itemModSet.Item3.Mod4.Content.Icon)
		}
	} else if item.Type == 4 {
		itemModSet := getModifierActionRelations(item, "item4", nil, "", "")
		log.Printf("Item4:\n%+v", itemModSet.Item3)
		it4 := wf.NewItem(itemModSet.Item4.NoMod.Content.Title).
			Subtitle(itemModSet.Item4.NoMod.Content.Subtitle).Valid(true).
			Arg(itemModSet.Item4.NoMod.Content.Arg).
			UID(item.Name).
			Var("notification", itemModSet.Item4.NoMod.Content.Notification).
			Var("action", itemModSet.Item4.NoMod.Content.Action).
			Var("action2", itemModSet.Item4.NoMod.Content.Action2).
			Var("action3", itemModSet.Item4.NoMod.Content.Action3).
			Arg(itemModSet.Item4.NoMod.Content.Arg).
			Icon(itemModSet.Item4.NoMod.Content.Icon)
		if itemModSet.Item4.Mod1.ModKey != nil {
			it4.NewModifier(itemModSet.Item4.Mod1.ModKey[0:]...).
				Subtitle(itemModSet.Item4.Mod1.Content.Subtitle).
				Arg(itemModSet.Item4.Mod1.Content.Arg).
				Var("notification", itemModSet.Item4.Mod1.Content.Notification).
				Var("action", itemModSet.Item4.Mod1.Content.Action).
				Var("action2", itemModSet.Item4.Mod1.Content.Action2).
				Var("action3", itemModSet.Item4.Mod1.Content.Action3).
				Arg(itemModSet.Item4.Mod1.Content.Arg).
				Icon(itemModSet.Item4.Mod1.Content.Icon)
		}
		if itemModSet.Item4.Mod2.ModKey != nil {
			it4.NewModifier(itemModSet.Item4.Mod2.ModKey[0:]...).
				Subtitle(itemModSet.Item4.Mod2.Content.Subtitle).
				Arg(itemModSet.Item4.Mod2.Content.Arg).
				Var("notification", itemModSet.Item4.Mod2.Content.Notification).
				Var("action", itemModSet.Item4.Mod2.Content.Action).
				Var("action2", itemModSet.Item4.Mod2.Content.Action2).
				Var("action3", itemModSet.Item4.Mod2.Content.Action3).
				Arg(itemModSet.Item4.Mod2.Content.Arg).
				Icon(itemModSet.Item4.Mod2.Content.Icon)
		}
		if itemModSet.Item4.Mod3.ModKey != nil {
			it4.NewModifier(itemModSet.Item4.Mod3.ModKey[0:]...).
				Subtitle(itemModSet.Item4.Mod3.Content.Subtitle).
				Arg(itemModSet.Item4.Mod3.Content.Arg).
				Var("notification", itemModSet.Item4.Mod3.Content.Notification).
				Var("action", itemModSet.Item4.Mod3.Content.Action).
				Var("action2", itemModSet.Item4.Mod3.Content.Action2).
				Var("action3", itemModSet.Item4.Mod3.Content.Action3).
				Arg(itemModSet.Item4.Mod3.Content.Arg).
				Icon(itemModSet.Item4.Mod3.Content.Icon)
		}
		if itemModSet.Item4.Mod4.ModKey != nil {
			it4.NewModifier(itemModSet.Item4.Mod4.ModKey[0:]...).
				Subtitle(itemModSet.Item4.Mod4.Content.Subtitle).
				Arg(itemModSet.Item4.Mod4.Content.Arg).
				Var("notification", itemModSet.Item4.Mod4.Content.Notification).
				Var("action", itemModSet.Item4.Mod4.Content.Action).
				Var("action2", itemModSet.Item4.Mod4.Content.Action2).
				Var("action3", itemModSet.Item4.Mod4.Content.Action3).
				Arg(itemModSet.Item4.Mod4.Content.Arg).
				Icon(itemModSet.Item4.Mod4.Content.Icon)
		}
	} else {
		log.Printf("New item, needs to be implemented.")
	}
}

func addRefreshCacheItem() {
	wf.NewItem("Refresh Bitwardens Secret Cache").
		Subtitle("Fill the cache with cleaned Bitwarden secrets (the real secrets, are not kept in the cached)").
		Valid(true).
		UID("cache").
		Icon(ReloadIcon()).
		Var("action", "-cache").
		Var("notification", "Refreshing Bitwarden Workflow cache").
		Arg("-background")
}
