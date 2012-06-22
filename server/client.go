package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

const (
	serviceUrl = "http://transport.opendata.ch/v1/locations"
)

func QueryLocations(query url.Values) (string, error) {
	u, err := url.Parse(serviceUrl)

	if err != nil {
		log.Fatal(err)
	}

	u.RawQuery = query.Encode()

	res, err := http.Get(u.String())

	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	json, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	return string(json), nil
}
