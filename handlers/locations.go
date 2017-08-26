package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/agneum/travels/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Location struct {
	Id       uint32 `json:"id"`
	Place    string `json:"place"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Distance uint32 `json:"distance"`
}

func GetLocation(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var location Location
		c := session.DB("travels").C("locations")

		locationId, err := utils.ParseIdParameter(ctx.Param("id"))
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		err = c.Find(bson.M{"id": locationId}).One(&location)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		data, err := json.Marshal(location)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		utils.ResponseWithJSON(ctx, data, http.StatusOK)
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
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
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
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		utils.ResponseWithJSON(ctx, []byte(fmt.Sprintf("{\"avg\":%.5f}", averageMark["avg"])), http.StatusOK)
		return nil
	}
}

func getCoreFiltersForAverageMark(ctx *routing.Context) (map[string]interface{}, error) {

	coreFilters := make(map[string]interface{}, 3)

	locationId, err := utils.ParseIdParameter(ctx.Param("id"))
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
