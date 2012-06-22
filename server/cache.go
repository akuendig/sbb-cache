package main

import (
	"labix.org/v2/mgo"
	"log"
	"net/url"
	"os"
	"strconv"
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
	X, Y  int    ",omitempty"
	Query string ",omitempty"
	Tpe   string ",omitempty"
	Json  string
}

func GetCached(query url.Values) (string, error) {
	var q = make(map[string]interface{})

	if val := query.Get("query"); len(val) > 0 {
		q["query"] = val
	}

	if x, y := query.Get("x"), query.Get("y"); len(x) > 0 && len(y) > 0 {
		q["x"] = x
		q["y"] = y
	}

	if t := query.Get("type"); len(t) > 0 {
		q["tpe"] = t
	}

	var result = &cachedLocation{}
	err := session.DB("heroku_app5462032").C("locations").Find(q).One(&result)

	if err != nil {
		return "", err
	}

	return result.Json, nil
}

func SetCached(loc *cachedLocation) error {
	return session.DB("heroku_app5462032").C("locations").Insert(loc)
}

func NewLocation(query url.Values) *cachedLocation {
	var loc = &cachedLocation{}

	if val := query.Get("query"); len(val) > 0 {
		loc.Query = val
	}

	if x, y := query.Get("x"), query.Get("y"); len(x) > 0 && len(y) > 0 {
		loc.X, _ = strconv.Atoi(x)
		loc.Y, _ = strconv.Atoi(y)
	}

	if t := query.Get("type"); len(t) > 0 {
		loc.Tpe = t
	}

	return loc
}
