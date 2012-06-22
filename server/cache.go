package main

import (
	"launchpad.net/mgo"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

var session *mgo.Session

func init() {
	var err error
	session, err = mgo.Dial(os.Getenv("MONGOLAB_URI"))

	if err != nil {
		log.Fatal(err)
	}
}

type cachedLocation struct {
	x, y    int
	query   string
	tpe     string
	json    string
	expires time.Time
}

func GetCached(query url.Values) (string, error) {
	var q map[string]interface{}

	if val := query.Get("query"); len(val) > 0 {
		q["query"] = val[0]
	}

	if x, y := query.Get("x"), query.Get("y"); len(x) > 0 && len(y) > 0 {
		q["x"] = x[0]
		q["y"] = y[0]
	}

	if t := query.Get("type"); len(t) > 0 {
		q["tpe"] = t[0]
	}

	var result = &cachedLocation{}
	err := session.DB("heroku_app5462032").C("locations").Find(q).Sort("-$natural").One(&result)

	if err != nil {
		return "", err
	}

	if result.expires.Before(time.Now()) {
		return "", nil
	}

	return result.json, nil
}

func SetCached(loc *cachedLocation) error {
	loc.expires = time.Now().Add(time.Minute)
	return session.DB("heroku_app5462032").C("locations").Insert(loc)
}

func NewLocation(query url.Values) *cachedLocation {
	var loc = &cachedLocation{}

	if val := query["query"]; len(val) > 0 {
		loc.query = val[0]
	}

	var x = query["x"]
	var y = query["y"]

	if len(x) > 0 && len(y) > 0 {
		loc.x, _ = strconv.Atoi(x[0])
		loc.y, _ = strconv.Atoi(y[0])
	}

	if t := query["type"]; len(t) > 0 {
		loc.tpe = t[0]
	}

	return loc
}
