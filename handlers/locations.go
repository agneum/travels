package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/agneum/travels/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

//easyjson:json
type Location struct {
	Id       uint32 `json:"id"`
	Place    string `json:"place"`
	Country  string `json:"country"`
	City     string `json:"city"`
	Distance uint32 `json:"distance"`
}

func CreateLocation(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		location := &Location{}
		err := location.UnmarshalJSON(ctx.Request.Body())

		if err != nil || location.Id == 0 {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("locations")
		err = c.Insert(location)

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		utils.ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
}

func UpdateLocation(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		locationId, err := utils.ParseIdParameter(ctx.Param("id"))
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		c := session.DB("travels").C("locations")
		count, err := c.Find(bson.M{"id": locationId}).Count()
		if err != nil || count == 0 {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		var location map[string]interface{}
		err = bson.UnmarshalJSON([]byte(ctx.Request.Body()), &location)

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		for _, v := range location {
			if v == nil {
				utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
				return nil
			}
		}

		err = c.Update(bson.M{"id": locationId}, bson.M{"$set": &location})

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		utils.ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
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

		data, err := location.MarshalJSON()
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
		coreFilters, err := getCoreFiltersForAverageMark(ctx)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		session := s.Copy()
		defer session.Close()

		l := session.DB("travels").C("locations")
		count, err := l.Find(bson.M{"id": coreFilters["location"]}).Count()
		if err != nil || count == 0 {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		userFilters, err := getUserFiltersForAverageMark(ctx)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(err.Error()), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("visits")

		pipeline := make([]bson.M, 0, 5)
		pipeline = append(pipeline, bson.M{"$match": coreFilters})

		if len(userFilters) > 0 {
			pipeline = append(pipeline, bson.M{
				"$lookup": bson.M{
					"from":         "users",
					"localField":   "user",
					"foreignField": "id",
					"as":           "user",
				},
			},
				bson.M{"$match": userFilters},
				bson.M{"$unwind": "$user"})
		}

		pipeline = append(pipeline, bson.M{"$group": bson.M{
			"_id": "$location",
			"avg": bson.M{"$avg": "$mark"},
		}})

		averageMark := bson.M{}

		err = c.Pipe(pipeline).One(&averageMark)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(fmt.Sprintf("{\"avg\":0.0}")), http.StatusOK)
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

func getUserFiltersForAverageMark(ctx *routing.Context) (map[string]interface{}, error) {
	userFilters := make(map[string]interface{}, 3)

	if gender := ctx.QueryArgs().Peek("gender"); len(gender) > 0 {
		g := string(gender)
		if g != "m" && g != "f" {
			return nil, errors.New("")
		}
		userFilters["user.gender"] = g
	}

	age := make(map[string]int64, 2)
	currentTime := time.Now()

	if fromAge := ctx.QueryArgs().Peek("fromAge"); len(fromAge) > 0 {
		date, err := strconv.Atoi(string(fromAge))
		if err != nil {
			return nil, err
		}

		age["$lt"] = currentTime.AddDate(-1*date, 0, 0).Unix()
	}

	if toAge := ctx.QueryArgs().Peek("toAge"); len(toAge) > 0 {
		date, err := strconv.Atoi(string(toAge))
		if err != nil {
			return nil, err
		}
		age["$gt"] = currentTime.AddDate(-1*date, 0, 0).Unix()
	}

	if len(age) > 0 {
		userFilters["user.birth_date"] = age
	}

	return userFilters, nil
}
