package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/agneum/travels/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Visit struct {
	Id        uint32 `json:"id"`
	Location  uint32 `json:"location"`
	User      uint32 `json:"user"`
	VisitedAt uint32 `json:"visited_at" bson:"visited_at"`
	Mark      uint8  `json:"mark"`
}

func CreateVisit(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var visit Visit
		err := json.Unmarshal(ctx.Request.Body(), &visit)

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("visits")
		err = c.Insert(visit)

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		utils.ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
}

func UpdateVisit(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		visitId, err := utils.ParseIdParameter(ctx.Param("id"))
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		var visit map[string]interface{}
		err = bson.UnmarshalJSON([]byte(ctx.Request.Body()), &visit)

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("visits")
		err = c.Update(bson.M{"id": visitId}, bson.M{"$set": &visit})

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		utils.ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
}

func GetVisit(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var visit Visit
		c := session.DB("travels").C("visits")

		visitId, err := utils.ParseIdParameter(ctx.Param("id"))
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		err = c.Find(bson.M{"id": visitId}).One(&visit)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		data, err := json.Marshal(visit)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		utils.ResponseWithJSON(ctx, data, http.StatusOK)
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
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		locationFilters, err := getLocationFiltersForUserVisits(ctx)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
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
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		response := make(map[string][]bson.M, 1)
		response["visits"] = visits
		data, err := json.Marshal(response)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		utils.ResponseWithJSON(ctx, data, http.StatusOK)
		return nil
	}
}

func getCoreFiltersForUserVisits(ctx *routing.Context) (map[string]interface{}, error) {

	coreFilters := make(map[string]interface{}, 3)

	userId, err := utils.ParseIdParameter(ctx.Param("id"))
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
