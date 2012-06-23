package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/location", location)

	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)

	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func location(w http.ResponseWriter, req *http.Request) {
	var query = req.URL.Query()

	json, err := GetCached(query)

	if err != nil {
		log.Println("GetCached:", err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if json == "" {
		json, err = QueryLocations(query)

		if err != nil {
			log.Println("QueryLocations:", err.Error())
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		var loc = NewLocation(query)
		loc.Json = json

		SetCached(loc)

		io.WriteString(w, json)
	} else {
		io.WriteString(w, json)
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, `The only service endpoint is /location.
Refer to <a href='http://transport.opendata.ch/#locations'>TransportAPI</a> for more information.`)
}
