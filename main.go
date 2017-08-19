package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

type User struct {
	Id        uint32 `json:"id"`
	Email     string `json:"email"`
	Firstname string `json:"first_name" bson:"first_name"`
	Lastname  string `json:"last_name" bson:"last_name"`
	Gender    string `json:"gender"`
	Birthdate int32  `json:"birth_date" bson:"birth_date"`
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
	VisitedAt uint32 `json:"visited_at" bson:"visited_at"`
	Mark      uint8  `json:"mark"`
}

func ResponseWithJSON(ctx *routing.Context, json []byte, code int) {
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

	router := routing.New()
	router.Get(`/users/<id:\d+>`, GetUser(session))
	router.Get(`/locations/<id:\d+>`, GetLocation(session))
	router.Get(`/visits/<id:\d+>`, GetVisit(session))
	router.Post("/users/new", CreateUser(session))
	router.Post(`/users/<id:\d+>`, UpdateUser(session))

	panic(fasthttp.ListenAndServe(":8084", router.HandleRequest))
}

func GetUser(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var user User
		c := session.DB("travels").C("users")

		userId, err := parseIdParameter(ctx.Param("id"))
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		err = c.Find(bson.M{"id": userId}).One(&user)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		data, err := json.Marshal(user)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		ResponseWithJSON(ctx, data, http.StatusOK)
		return nil
	}
}

func CreateUser(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var user User
		err := json.Unmarshal(ctx.Request.Body(), &user)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("users")
		err = c.Insert(user)

		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
}

func UpdateUser(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		userId, err := parseIdParameter(ctx.Param("id"))
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		var user interface{}
		err = bson.UnmarshalJSON([]byte(ctx.Request.Body()), &user)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("users")
		err = c.Update(bson.M{"_id": userId}, bson.M{"$set": &user})

		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
}

func GetLocation(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var location Location
		c := session.DB("travels").C("locations")

		locationId, err := parseIdParameter(ctx.Param("id"))
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		err = c.Find(bson.M{"id": locationId}).One(&location)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		data, err := json.Marshal(location)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		ResponseWithJSON(ctx, data, http.StatusOK)
		return nil
	}
}

func GetVisit(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var visit Visit
		c := session.DB("travels").C("visits")

		visitId, err := parseIdParameter(ctx.Param("id"))
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		err = c.FindId(bson.M{"id": visitId}).One(&visit)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		data, err := json.Marshal(visit)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		ResponseWithJSON(ctx, data, http.StatusOK)
		return nil
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
