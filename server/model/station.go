package main

import (
	"encoding/xml"
)

type LocationRequest struct {
	XMLName    xml.Name `xml:"LocValReq"`
	Id         string   `xml:",attr"`
	SearchMode string   `xml:"sMode,attr"`
	ReqLoc     struct {
		Match string `xml:"match,attr"`
		Type  string `xml:"type,attr"`
	}
}

type Station struct {
	Name              string
	X                 int
	Y                 int
	Type              string
	ExternalId        string
	ExternalStationNr string
	Puic              string
	ProdClass         string
	Urlname           string
}
