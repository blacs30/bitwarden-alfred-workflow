package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	aw "github.com/deanishe/awgo"
	"github.com/jpillora/go-tld"
	"github.com/jychri/tilde"
)

func popuplateCacheItems(items []Item) {
	var cacheItems []Item
	for _, item := range items {
		var tempItem Item
		tempItem.Object = item.Object
		tempItem.Id = item.Id
		tempItem.OrganizationId = item.OrganizationId
		tempItem.FolderId = item.FolderId
		tempItem.Type = item.Type
		tempItem.Name = item.Name
		tempItem.Favorite = item.Favorite
		tempItem.CollectionIds = item.CollectionIds
		tempItem.RevisionDate = item.RevisionDate

		// special cases because we don't want to cache secrets
		if item.Type == 2 {
			noteValue := ""
			if item.Notes != "" {
				noteValue = "hidden"
			}
			tempItem.Notes = noteValue
		} else {
			tempItem.Notes = item.Notes
		}
		shortNumber := item.Card.Number
		if item.Card.Number != "" {
			shortNumber = fmt.Sprintf("*%s", shortNumber[len(shortNumber)-4:])
		}
		codeValue := "hidden"
		if item.Card.Code == "" {
			codeValue = ""
		}
		tempItem.Card = CardInfo{
			CardHolderName: item.Card.CardHolderName,
			Brand:          item.Card.Brand,
			Number:         shortNumber,
			ExpMonth:       item.Card.ExpMonth,
			ExpYear:        item.Card.ExpYear,
			Code:           codeValue,
		}
		tempItem.SecureNote = item.SecureNote
		passwordValue := "hidden"
		if item.Login.Password == "" {
			passwordValue = ""
		}
		totpValue := "hidden"
		if item.Login.Totp == "" {
			totpValue = ""
		}
		tempItem.Login = Login{
			Uris:                 item.Login.Uris,
			Username:             item.Login.Username,
			Password:             passwordValue,
			Totp:                 totpValue,
			PasswordRevisionDate: item.Login.PasswordRevisionDate,
		}
		tempItem.Identity = Identity{
			Title:          item.Identity.Title,
			FirstName:      item.Identity.FirstName,
			MiddleName:     item.Identity.MiddleName,
			LastName:       item.Identity.LastName,
			Address1:       item.Identity.Address1,
			Address2:       item.Identity.Address2,
			Address3:       item.Identity.Address3,
			City:           item.Identity.City,
			State:          item.Identity.State,
			PostalCode:     item.Identity.PostalCode,
			Country:        item.Identity.Country,
			Company:        item.Identity.Company,
			Email:          item.Identity.Email,
			Phone:          item.Identity.Email,
			Ssn:            item.Identity.Ssn,
			Username:       item.Identity.Username,
			PassportNumber: item.Identity.PassportNumber,
			LicenseNumber:  item.Identity.LicenseNumber,
		}
		var tempFields []Field
		for _, field := range item.Fields {
			if field.Type == 1 {
				valueContent := "hidden"
				if field.Value == "" {
					valueContent = ""
				}
				tempFields = append(tempFields, Field{
					Name:  field.Name,
					Value: valueContent,
					Type:  field.Type,
				})
			} else {
				tempFields = append(tempFields, Field{
					Name:  field.Name,
					Value: field.Value,
					Type:  field.Type,
				})
			}
		}
		tempItem.Fields = tempFields

		// handling attchements slice here
		var tempAttachments []Attachments
		for _, att := range item.Attachments {
			tempAttachments = append(tempAttachments, Attachments{
				Id:       att.Id,
				FileName: att.FileName,
				Size:     att.Size,
				SizeName: att.SizeName,
				Url:      att.Url,
			})
		}
		tempItem.Attachments = tempAttachments

		// last step: appending cached items
		cacheItems = append(cacheItems, tempItem)
	}

	data, err := json.Marshal(cacheItems)
	if err != nil {
		log.Println(err)
	}
	Encrypt(data)

	if conf.IconCacheEnabled && (wf.Data.Expired(ICON_CACHE_NAME, conf.IconMaxCacheAge) || !wf.Data.Exists(ICON_CACHE_NAME)) {
		getIcon(wf)
	}
}

func getIcon(workflow *aw.Workflow) {
	if !wf.IsRunning("icons") {
		// start job
		cmd := exec.Command(os.Args[0], "-icons")
		log.Println("geticon cmd: ", cmd)
		if err := wf.RunInBackground("icons", cmd); err != nil {
			log.Println(err)
			wf.FatalError(err)
		}
		log.Println("Started job icons: ", wf.IsRunning("icons"))
		return
	} else {
		log.Printf("Download icons job already running.")
	}
}

func popuplateCacheFolders(folders []Folder) {
	var cacheFolders []Folder
	for _, folder := range folders {
		var tempFolder Folder
		tempFolder.Name = folder.Name
		tempFolder.Object = folder.Object
		tempFolder.Id = folder.Id
		cacheFolders = append(cacheFolders, tempFolder)
	}

	err := wf.Cache.StoreJSON(FOLDER_CACHE_NAME, cacheFolders)
	if err != nil {
		log.Println(err)
	}
}

func DownloadIcon(urlMap map[string]string, outputFolder string) {
	//get https://icons.duckduckgo.com/ip3/maersk-analytics.atlassian.net.ico
	//fullUrlFile = fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s", urlString)
	for id, url := range urlMap {
		if !strings.HasPrefix(url, "http") {
			url = fmt.Sprintf("http://%s", url)
		}
		u, err := tld.Parse(url)
		if err != nil {
			log.Println("Error parsing url: ", url, "error: ", err)
			continue
		}
		if u.Host != "" {
			url = u.Host
			url = fmt.Sprintf("https://icons.duckduckgo.com/ip3/%s.ico", url)
		}
		filePath := fmt.Sprintf("%s/%s.png", outputFolder, id)
		if _, err := os.Stat(filePath); err == nil {
			log.Print("Skip downloading icon for url, most likely exists already: ", url)
		} else {
			err = DownloadFile(filePath, url)
			if err != nil {
				log.Print("Download icon error: ", err)
			}
		}
	}
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	log.Println("Dowloading url: ", url, " to file: ", filepath)

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func runGetIcons(url string, id string) {
	log.Println("Background?", opts.Background)
	if opts.Background {
		if !wf.IsRunning("icons") {
			cmd := exec.Command(os.Args[0], "-icons")
			log.Println("Download Icon cmd: ", cmd)
			if err := wf.RunInBackground("icons", cmd); err != nil {
				wf.FatalError(err)
			}
		} else {
			log.Printf("Download icons job already running.")
		}
		searchAlfred(conf.BwKeyword)
		return
	}

	urlIdMap := make(map[string]string)
	if url == "" && id == "" {
		// Load data
		var items []Item
		if wf.Cache.Exists(CACHE_NAME) {
			data, err := Decrypt()
			if err != nil {
				log.Printf("Error decrypting data: %s", err)
			}
			if err := json.Unmarshal(data, &items); err != nil {
				log.Printf("Couldn't load the items cache, error: %s", err)
			}
			for _, item := range items {
				if item.Type == 1 {
					if len(item.Login.Uris) > 0 {
						urlIdMap[item.Id] = item.Login.Uris[0].Uri
					}
				}
			}
		}
	} else {
		urlIdMap[id] = url
	}

	// Marshal the map into a JSON string.
	jsonMarshall, err := json.Marshal(urlIdMap)
	if err != nil {
		log.Println(err.Error())
		return
	}
	jsonStr := string(jsonMarshall)

	outputFolder := tilde.Abs(fmt.Sprintf("%s/urlicon", wf.DataDir()))
	err = os.MkdirAll(outputFolder, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
	DownloadIcon(urlIdMap, outputFolder)
	err = wf.Data.StoreJSON(ICON_CACHE_NAME, jsonStr)
	if err != nil {
		log.Println(err)
	}
	log.Println("Finished downloading icons.")
}
