package service

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	xmlUri  = "http://fahrplan.sbb.ch/bin/extxml.exe"
	jsonUri = "http://fahrplan.sbb.ch/bin/query.exe/dny"
)

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

func queryXml(cont *SbbRequest) (*http.Response, error) {
	var body = new(bytes.Buffer)
	var encoder = xml.NewEncoder(body)

	body.WriteString(xml.Header)

	if err := encoder.Encode(cont); err != nil {
		return nil, err
	}

	if err := encoder.Flush(); err != nil {
		return nil, err
	}

	log.Println("Requesting:\n", body)

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

/***************************************************************
 ****************** Location ***********************************/

type LocValReq struct {
	XMLName    xml.Name `xml:"LocValReq"`
	Id         string   `xml:"id,attr"`
	SearchMode string   `xml:"sMode,attr"`
	ReqLoc     struct {
		Match string `xml:"match,attr"`
		Type  string `xml:"type,attr"`
	}
}

type LocValRes struct {
	XMLName   xml.Name  `xml:"ResC"`
	Pois      []Station `xml:"LocValRes>Poi"`
	Stations  []Station `xml:"LocValRes>Station"`
	Addresses []Station `xml:"LocValRes>Address"`
}

func NewLocationQuery(location string) (*LocValRes, error) {
	var reqCont = &LocValReq{
		Id:         "station",
		SearchMode: "1",
	}

	reqCont.ReqLoc.Match = location
	reqCont.ReqLoc.Type = "ALLTYPE"

	var cont = &SbbRequest{
		Lang:     "en",
		Prod:     "iPhone3.1",
		Ver:      "2.3",
		AccessId: "MJXZ841ZfsmqqmSymWhBPy5dMNoqoGsHInHbWJQ5PTUZOJ1rLTkn8vVZOZDFfSe",
		Content:  reqCont,
	}

	var res, err = queryXml(cont)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	var decoder = xml.NewDecoder(res.Body)
	var stations LocValRes

	if err := decoder.Decode(&stations); err != nil {
		return nil, err
	}

	return &stations, nil
}

/***************************************************************
 ****************** StationBoard *******************************/

type STBRes struct {
	XMLName     xml.Name     `xml:"ResC"`
	JourneyList []STBJourney `xml:"STBRes>JourneyList>STBJourney"`
}

type STBJourney struct {
	JHandle              JHandle
	Stop                 BasicStop          `xml:"MainStop>BasicStop"`
	JourneyAttributeList []JourneyAttribute `xml:"JourneyAttributeList>JourneyAttribute"`
}

type JHandle struct {
	Tnr   string `xml:"tNr,attr"`
	Puic  string `xml:"puic,attr"`
	Cycle string `xml:"cycle,attr"`
}

type BasicStop struct {
	Station     Station
	Dep         StopTime
	Arr         StopTime
	Capacity1st int `xml:"StopPrognosis>Capacity1st"`
	Capacity2nd int `xml:"StopPrognosis>Capacity2nd"`
}

type StopTime struct {
	TimeRaw  string `xml:"Time"`
	Platform int    `xml:"Platform>Text"`
}

func (t StopTime) Time() time.Time {
	var time, _ = time.Parse("15:04", t.TimeRaw)
	return time
}

type JourneyAttribute struct {
	From     int    `xml:"from,attr"`
	To       int    `xml:"to,attr"`
	AttrType string `xml:"Attribute>type,attr"`
	VarType  string `xml:"Attribute>AttributeVariant>type,attr"`
	Text     string `xml:"Attribute>AttributeVariant>Text"`
}

type DateType string

const (
	Date_Type_Departure DateType = "DEP"
	Date_Type_Arrival   DateType = "ARR"
)

type STBReq struct {
	XMLName     xml.Name `xml:"STBReq"`
	DateType    DateType `xml:"dateType,attr"`
	MaxJourneys int      `xml:"maxJourneys,attr"`
	Time        string

	Period struct {
		DateBegin struct {
			Date string
		}
		DateEnd struct {
			Date string
		}
	}

	TableStation struct {
		ExternalId string `xml:"externalId,attr"`
	}

	ProductFilter struct {
		ProductFilter string `xml:"ProductFilter,attr"`
	}
}

const (
	Trans_Tramway = 1 << (6 + iota)
	Trans_Arz_Ext
	Trans_Cableway
	Trans_Bus
	Trans_S_Sn_R
	Trans_Ship
	Trans_Re_D
	Trans_Ir
	Trans_Ec_Ic
	Trans_Ice_Tgv_Rj
	Trans_All = 1<<16 - 1
)

var ErrLocationNotFound = errors.New("Did not find location to query stationboard")

func NewStationBoardQuery(location string) (interface{}, error) {
	var locations, err = NewLocationQuery(location)

	if err != nil {
		return nil, err
	}

	if len(locations.Stations) == 0 {
		return nil, ErrLocationNotFound
	}

	var loc = locations.Stations[0]

	var reqCont = &STBReq{
		DateType:    Date_Type_Departure,
		MaxJourneys: 40,
		Time:        time.Now().Format("15:04"),
	}

	reqCont.Period.DateBegin.Date = time.Now().Format("2006-01-02")
	reqCont.Period.DateBegin.Date = time.Now().Format("2006-01-02")

	reqCont.TableStation.ExternalId = loc.ExternalId

	reqCont.ProductFilter.ProductFilter = fmt.Sprintf("%b", Trans_All)

	var cont = &SbbRequest{
		Lang:     "en",
		Prod:     "iPhone3.1",
		Ver:      "2.3",
		AccessId: "MJXZ841ZfsmqqmSymWhBPy5dMNoqoGsHInHbWJQ5PTUZOJ1rLTkn8vVZOZDFfSe",
		Content:  reqCont,
	}

	res, err := queryXml(cont)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	//a, _ := ioutil.ReadAll(res.Body)
	//return string(a), nil

	var decoder = xml.NewDecoder(res.Body)
	decoder.CharsetReader = CharsetReader
	var response STBRes

	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}

	return response.JourneyList, nil
}

/* DevHttp
{"requests":[{"id":"6F9983D2-8070-4B69-B7A2-D848277E9EB6", "name":"sbb-board", "request":{"protocol":"HTTP", "method":{"name":"POST", "request-body":true}, "url":"fahrplan.sbb.ch/bin/extxml.exe", "body-type":"Text", "content":"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<ReqC lang=\"en\" prod=\"iPhone3.1\" ver=\"2.3\" accessId=\"MJXZ841ZfsmqqmSymWhBPy5dMNoqoGsHInHbWJQ5PTUZOJ1rLTkn8vVZOZDFfSe\">\n<STBReq dateType=\"DEP\" maxJourneys=\"40\">\n<Time>16:42</Time>\n<Period><DateBegin><Date>2012-07-01</Date></DateBegin><DateEnd><Date>2012-07-01</Date></DateEnd></Period>\n<TableStation externalId=\"008509003#85\"></TableStation>\n</STBReq></ReqC>", "headers-type":"Form", "url-assist":false}, "modified":"Sun, 1 Jul 2012 16:25:28 +0200"},{"id":"86427DAF-70AB-4D32-B6DE-963A05480F22", "name":"sbb-location", "request":{"protocol":"HTTP", "method":{"name":"POST", "request-body":true}, "url":"fahrplan.sbb.ch/bin/extxml.exe", "body-type":"Text", "content":"<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<ReqC lang=\"en\" prod=\"iPhone3.1\" ver=\"2.3\" accessId=\"MJXZ841ZfsmqqmSymWhBPy5dMNoqoGsHInHbWJQ5PTUZOJ1rLTkn8vVZOZDFfSe\"><LocValReq id=\"station\" sMode=\"1\"><ReqLoc match=\"maienf\" type=\"ALLTYPE\"></ReqLoc></LocValReq></ReqC>", "headers-type":"Form", "url-assist":false}, "modified":"Sun, 1 Jul 2012 15:21:16 +0200"}], "version":1, "modified":"Sun, 1 Jul 2012 16:25:28 +0200"}
*/
