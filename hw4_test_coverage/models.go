package main

import "encoding/xml"

type Dataset struct {
	XMLName xml.Name `xml:"root"`
	Row     []Row    `xml:"row"`
}

type Row struct {
	Id        int    `xml:"id"`
	Age       int    `xml:"age"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Gender    string `xml:"gender"`
	About     string `xml:"about"`
}
