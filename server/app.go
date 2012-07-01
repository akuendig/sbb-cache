package main

import (
	"fmt"
	"io"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"./service"
)

func main() {
	http.HandleFunc("/", locationAlternative)
	http.HandleFunc("/location", location)

	err := http.ListenAndServe(":9000", nil)
	//err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func location(w http.ResponseWriter, req *http.Request) {
	var query = req.URL.Query()
	var json, err = GetCached(query)

	switch err {
	case nil:
		io.WriteString(w, json)

	case mgo.ErrNotFound:
		json, err = QueryLocations(query)

		if err != nil {
			log.Println("QueryLocations:", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		var loc = NewLocation(query)
		loc.Json = json

		SetCached(loc)

		io.WriteString(w, json)

	default:
		log.Println("GetCached:", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func locationAlternative(w http.ResponseWriter, req *http.Request) {
	var query = req.URL.Query()
	var q, err = service.NewStationBoardQuery(query.Get("s"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "%+v", q)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, `The only service endpoint is /location.
Refer to <a href='http://transport.opendata.ch/#locations'>TransportAPI</a> for more information.`)
}
