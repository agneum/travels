package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	mgo "gopkg.in/mgo.v2"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type User struct {
	Id        uint32 `json:"id" bson:"_id"`
	Email     string `json:"email"`
	Firstname string `json:"first_name"`
	Lastname  string `json:"last_name"`
	Gender    string `json:"gender"`
	Birthdate uint32 `json:"birth_date"`
}

type Location struct {
	Id       uint32 `json:"id"`
	Place    string `json:"place"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Distance uint32 `json:"distance"`
}

type Visit struct {
	Id        uint32 `json:"id"`
	Location  uint32 `json:"location"`
	User      uint32 `json:"user"`
	VisitedAt uint32 `json:"visited_at"`
	Mark      uint8  `json:"mark"`
}

func ResponseWithJSON(ctx *fasthttp.RequestCtx, json []byte, code int) {
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(code)
	ctx.SetBody(json)
}

func main() {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)
	ensureIndex(session)

	router := fasthttprouter.New()
	router.Handle("GET", "/users/:id", GetUser(session))
	router.Handle("GET", "/locations/:id", GetLocation(session))
	router.Handle("GET", "/visits/:id", GetVisit(session))
	router.Handle("POST", "/users/new", CreateUser(session))

	log.Fatal(fasthttp.ListenAndServe(":8084", router.Handler))
}

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	c := session.DB("travels").C("users")

	index := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     false,
	}

	err := c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func GetUser(s *mgo.Session) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		session := s.Copy()
		defer session.Close()

		var user User
		c := session.DB("travels").C("users")

		userId, err := parseIdParameter(ctx.UserValue("id"))
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		err = c.FindId(userId).One(&user)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		data, err := json.Marshal(user)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		ResponseWithJSON(ctx, data, http.StatusOK)
	}
}

func CreateUser(s *mgo.Session) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		session := s.Copy()
		defer session.Close()

		var user User
		err := json.Unmarshal(ctx.Request.Body(), &user)
		if err != nil {
			ResponseWithJSON(ctx, []byte("{}"), http.StatusBadRequest)
		}

		c := session.DB("travels").C("users")
		err = c.Insert(user)

		if err != nil {
			ResponseWithJSON(ctx, []byte("{}"), http.StatusBadRequest)
		}
	}
}

func GetLocation(s *mgo.Session) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		session := s.Copy()
		defer session.Close()

		var location Location
		c := session.DB("travels").C("locations")

		locationId, err := parseIdParameter(ctx.UserValue("id"))
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		err = c.FindId(locationId).One(&location)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		data, err := json.Marshal(location)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		ResponseWithJSON(ctx, data, http.StatusOK)
	}
}

func GetVisit(s *mgo.Session) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		session := s.Copy()
		defer session.Close()

		var visit Visit
		c := session.DB("travels").C("visits")

		visitId, err := parseIdParameter(ctx.UserValue("id"))
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		err = c.FindId(visitId).One(&visit)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		data, err := json.Marshal(visit)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return
		}

		ResponseWithJSON(ctx, data, http.StatusOK)
	}
}

func parseIdParameter(parameter interface{}) (id uint64, err error) {

	stringID, ok := parameter.(string)
	if !ok {
		return
	}

	id, err = strconv.ParseUint(stringID, 10, 32)
	if err != nil {
		return
	}

	return id, nil
}
