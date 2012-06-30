package main

import (
	"bytes"
	"encoding/xml"
	"net/http"
)

const (
	xmlUri  = "http://fahrplan.sbb.ch/bin/extxml.exe"
	jsonUri = "http://fahrplan.sbb.ch/bin/query.exe/dny"
)

type LocationRequest struct {
	XMLName    xml.Name `xml:"LocValReq"`
	Id         string   `xml:"id,attr"`
	SearchMode string   `xml:"sMode,attr"`
	ReqLoc     ReqLoc
}

type ReqLoc struct {
	Match string `xml:"match,attr"`
	Type  string `xml:"type,attr"`
}

type Station struct {
	Name              string `xml:"name,attr"`
	X                 int    `xml:"x,attr"`
	Y                 int    `xml:"y,attr"`
	Type              string `xml:"type,attr"`
	ExternalId        string `xml:"externalId,attr"`
	ExternalStationNr string `xml:"externalStationNr,attr"`
	Puic              string `xml:"puic,attr"`
	ProdClass         string `xml:"prodClass,attr"`
	Urlname           string `xml:"urlname,attr"`
}

type SbbRequest struct {
	XMLName  xml.Name `xml:"ReqC"`
	Lang     string   `xml:"lang,attr"`
	Prod     string   `xml:"prod,attr"`
	Ver      string   `xml:"ver,attr"`
	AccessId string   `xml:"accessId,attr"`
	Content  interface{}
}

type LocationResponse struct {
	XMLName  xml.Name  `xml:"ResC"`
	Stations []Station `xml:"LocValRes>Station"`
}

func NewCoordinateQuery(x, y int) {

}

func requestLocationQuery(location string) (*http.Response, error) {
	var cont = &SbbRequest{
		Lang:     "en",
		Prod:     "iPhone3.1",
		Ver:      "2.3",
		AccessId: "MJXZ841ZfsmqqmSymWhBPy5dMNoqoGsHInHbWJQ5PTUZOJ1rLTkn8vVZOZDFfSe",
		Content: LocationRequest{
			Id:         "station",
			SearchMode: "1",
			ReqLoc: ReqLoc{
				Match: location,
				Type:  "ALLTYPE",
			},
		},
	}

	var body = new(bytes.Buffer)
	var encoder = xml.NewEncoder(body)

	body.WriteString(xml.Header)

	if err := encoder.Encode(cont); err != nil {
		return nil, err
	}

	if err := encoder.Flush(); err != nil {
		return nil, err
	}

	var client = new(http.Client)
	var req, err = http.NewRequest("POST", xmlUri, body)

	if err != nil {
		return nil, err
	}

	//req.Header.Add("User-Agent", "SBBMobile/4.2 CFNetwork/485.13.9 Darwin/11.0.0")
	req.Header.Add("Accept", "application/xml")
	req.Header.Add("Content-Type", "application/xml")

	return client.Do(req)
}

func NewLocationQuery(location string) ([]Station, error) {
	var res, err = requestLocationQuery(location)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var decoder = xml.NewDecoder(res.Body)
	var stations LocationResponse

	if err := decoder.Decode(&stations); err != nil {
		return nil, err
	}

	return stations.Stations, nil
}
