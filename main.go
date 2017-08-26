package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/agneum/travels/importer"
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

func init() {
	importer.Import()
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
	router.Get(`/users/<id:\d+>/visits`, GetUserVisit(session))
	router.Get(`/locations/<id:\d+>/avg`, GetAverageMark(session))
	router.Post("/users/new", CreateUser(session))
	router.Post(`/users/<id:\d+>`, UpdateUser(session))

	panic(fasthttp.ListenAndServe(":80", router.HandleRequest))
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

func GetUserVisit(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		c := session.DB("travels").C("visits")

		coreFilters, err := getCoreFiltersForUserVisits(ctx)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		locationFilters, err := getLocationFiltersForUserVisits(ctx)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		pipeline := []bson.M{
			bson.M{"$match": coreFilters},
			bson.M{
				"$lookup": bson.M{
					"from":         "locations",
					"localField":   "location",
					"foreignField": "id",
					"as":           "location",
				},
			},
			bson.M{"$match": locationFilters},
			bson.M{"$unwind": "$location"},
			bson.M{"$sort": bson.M{"visited_at": 1}},
			bson.M{"$project": bson.M{
				"_id":        0,
				"mark":       1,
				"visited_at": 1,
				"place":      "$location.place",
			}},
		}

		visits := []bson.M{}

		err = c.Pipe(pipeline).All(&visits)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		response := make(map[string][]bson.M, 1)
		response["visits"] = visits
		data, err := json.Marshal(response)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		ResponseWithJSON(ctx, data, http.StatusOK)
		return nil
	}
}

func GetAverageMark(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		c := session.DB("travels").C("visits")

		coreFilters, err := getCoreFiltersForAverageMark(ctx)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		pipeline := []bson.M{
			bson.M{"$match": coreFilters},
			bson.M{"$group": bson.M{
				"_id": "$location",
				"avg": bson.M{"$avg": "$mark"},
			}},
			bson.M{"$project": bson.M{
				"_id": 0,
				"avg": 1,
			}},
		}

		averageMark := bson.M{}

		err = c.Pipe(pipeline).One(&averageMark)
		if err != nil {
			ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		ResponseWithJSON(ctx, []byte(fmt.Sprintf("{\"avg\":%.5f}", averageMark["avg"])), http.StatusOK)
		return nil
	}
}

func getCoreFiltersForAverageMark(ctx *routing.Context) (map[string]interface{}, error) {

	coreFilters := make(map[string]interface{}, 3)

	locationId, err := parseIdParameter(ctx.Param("id"))
	if err != nil {
		return nil, err
	}
	coreFilters["location"] = locationId

	visitedAt := make(map[string]int, 2)

	if fromDate := ctx.QueryArgs().Peek("fromDate"); len(fromDate) > 0 {
		date, err := strconv.Atoi(string(fromDate))
		if err != nil {
			return nil, err
		}
		visitedAt["$gt"] = date
	}

	if toDate := ctx.QueryArgs().Peek("toDate"); len(toDate) > 0 {
		date, err := strconv.Atoi(string(toDate))
		if err != nil {
			return nil, err
		}
		visitedAt["$lt"] = date
	}

	if len(visitedAt) > 0 {
		coreFilters["visited_at"] = visitedAt
	}

	return coreFilters, nil
}

func getCoreFiltersForUserVisits(ctx *routing.Context) (map[string]interface{}, error) {

	coreFilters := make(map[string]interface{}, 3)

	userId, err := parseIdParameter(ctx.Param("id"))
	if err != nil {
		return nil, err
	}
	coreFilters["user"] = userId

	visitedAt := make(map[string]int, 2)

	if fromDate := ctx.QueryArgs().Peek("fromDate"); len(fromDate) > 0 {
		date, err := strconv.Atoi(string(fromDate))
		if err != nil {
			return nil, err
		}
		visitedAt["$gt"] = date
	}

	if toDate := ctx.QueryArgs().Peek("toDate"); len(toDate) > 0 {
		date, err := strconv.Atoi(string(toDate))
		if err != nil {
			return nil, err
		}
		visitedAt["$lt"] = date
	}

	if len(visitedAt) > 0 {
		coreFilters["visited_at"] = visitedAt
	}

	return coreFilters, nil
}

func getLocationFiltersForUserVisits(ctx *routing.Context) (map[string]interface{}, error) {
	locationFilters := make(map[string]interface{}, 2)
	if distance := ctx.QueryArgs().Peek("toDistance"); len(distance) > 0 {
		dist, err := strconv.Atoi(string(distance))
		if err != nil {
			return nil, err
		}
		locationFilters["location.distance"] = bson.M{"$lt": dist}
	}

	if country := ctx.QueryArgs().Peek("country"); len(country) > 0 {
		locationFilters["location.country"] = bson.M{"$eq": string(country)}
	}

	return locationFilters, nil
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
