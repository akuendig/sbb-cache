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

	ensureCollection()
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

func ensureCollection() {
	var db = session.DB("heroku_app5462032")
	var col = db.C("locations")
	var cols, err = db.CollectionNames()
	var colExists = false

	if err != nil {
		log.Fatal(err)
	}

	for _, name := range cols {
		if name == "locations" {
			colExists = true
			break
		}
	}

	if !colExists {
		err = col.Create(&mgo.CollectionInfo{
			DisableIdIndex: false,
			Capped:         true,
			MaxBytes:       60 * 1024 * 1024,
		})

		if err != nil {
			log.Fatal(err)
		}
	}

	err = col.EnsureIndex(mgo.Index{
		Key:        []string{"x"},
		Background: true, // Allow other connections to use a not fully build index
		Sparse:     true,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = col.EnsureIndex(mgo.Index{
		Key:        []string{"y"},
		Background: true, // Allow other connections to use a not fully build index
		Sparse:     true,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = col.EnsureIndex(mgo.Index{
		Key:        []string{"query"},
		Background: true, // Allow other connections to use a not fully build index
		Sparse:     true,
	})

	if err != nil {
		log.Fatal(err)
	}

	err = col.EnsureIndex(mgo.Index{
		Key:        []string{"tpe"},
		Background: true, // Allow other connections to use a not fully build index
		Sparse:     true,
	})

	if err != nil {
		log.Fatal(err)
	}
}
