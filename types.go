// Copyright (c) 2020 Claas Lisowski <github@lisowski-development.com>
// MIT Licence - http://opensource.org/licenses/MIT

package main

import "time"

type Folder struct {
	Object string `json:"object"`
	Id     string `json:"id"`
	Name   string `json:"name"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

type Uri struct {
	Match int    `json:"match"`
	Uri   string `json:"uri"`
}

type Login struct {
	Uris                 []Uri     `json:"uris"`
	Username             string    `json:"username"`
	Password             string    `json:"password"`
	Totp                 string    `json:"totp"`
	PasswordRevisionDate time.Time `json:"passwordRevisionDate"`
}

type SecureNoteType struct {
	Type int `json:"type"`
}

type CardInfo struct {
	CardHolderName string `json:"cardholderName"`
	Brand          string `json:"brand"`
	Number         string `json:"number"`
	ExpMonth       string `json:"expMonth"`
	ExpYear        string `json:"expYear"`
	Code           string `json:"code"`
}

type Attachments struct {
	Id       string `json:"id"`
	FileName string `json:"fileName"`
	Size     string `json:"size"`
	SizeName string `json:"sizeName"`
	Url      string `json:"url"`
}

type Identity struct {
	Title          string `json:"title"`
	FirstName      string `json:"firstName"`
	MiddleName     string `json:"middleName"`
	LastName       string `json:"lastName"`
	Address1       string `json:"address1"`
	Address2       string `json:"address2"`
	Address3       string `json:"address3"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postalCode"`
	Country        string `json:"country"`
	Company        string `json:"company"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Ssn            string `json:"ssn"`
	Username       string `json:"username"`
	PassportNumber string `json:"passportNumber"`
	LicenseNumber  string `json:"licenseNumber"`
}

// 1: Login
// 2: SecureNote
// 3: Card
// 4: Identity
type Item struct {
	Object         string         `json:"object"`
	Id             string         `json:"id"`
	OrganizationId string         `json:"organizationId"`
	FolderId       string         `json:"folderId"`
	Type           int            `json:"type"`
	Name           string         `json:"name"`
	Notes          string         `json:"notes"`
	Favorite       bool           `json:"favorite"`
	Fields         []Field        `json:"fields"`
	Card           CardInfo       `json:"card,omitempty"`
	Login          Login          `json:"login,omitempty"`
	Identity       Identity       `json:"identity,omitempty"`
	SecureNote     SecureNoteType `json:"secureNote,omitempty"`
	CollectionIds  []string       `json:"collectionIds"`
	RevisionDate   time.Time      `json:"revisionDate"`
	Attachments    []Attachments  `json:"attachments,omitempty"`
}
