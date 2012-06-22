package main

import (
	"labix.org/v2/mgo"
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
	X, Y    int    "/c"
	Query   string "/c"
	Tpe     string "/c"
	Json    string
	Expires time.Time
}

func GetCached(query url.Values) (string, error) {
	var q = make(map[string]interface{})

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

	if result.Expires.Before(time.Now()) {
		return "", nil
	}

	return result.Json, nil
}

func SetCached(loc *cachedLocation) error {
	loc.Expires = time.Now().Add(time.Minute)
	return session.DB("heroku_app5462032").C("locations").Insert(loc)
}

func NewLocation(query url.Values) *cachedLocation {
	var loc = &cachedLocation{}

	if val := query["query"]; len(val) > 0 {
		loc.Query = val[0]
	}

	var x = query["x"]
	var y = query["y"]

	if len(x) > 0 && len(y) > 0 {
		loc.X, _ = strconv.Atoi(x[0])
		loc.Y, _ = strconv.Atoi(y[0])
	}

	if t := query["type"]; len(t) > 0 {
		loc.Tpe = t[0]
	}

	return loc
}
