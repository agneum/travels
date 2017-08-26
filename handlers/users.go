package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/agneum/travels/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Id        uint32 `json:"id"`
	Email     string `json:"email"`
	Firstname string `json:"first_name" bson:"first_name"`
	Lastname  string `json:"last_name" bson:"last_name"`
	Gender    string `json:"gender"`
	Birthdate int32  `json:"birth_date" bson:"birth_date"`
}

func GetUser(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		var user User
		c := session.DB("travels").C("users")

		userId, err := utils.ParseIdParameter(ctx.Param("id"))
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		err = c.Find(bson.M{"id": userId}).One(&user)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		data, err := json.Marshal(user)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		utils.ResponseWithJSON(ctx, data, http.StatusOK)
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
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("users")
		err = c.Insert(user)

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		utils.ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
}

func UpdateUser(s *mgo.Session) func(ctx *routing.Context) error {
	return func(ctx *routing.Context) error {
		session := s.Copy()
		defer session.Close()

		userId, err := utils.ParseIdParameter(ctx.Param("id"))
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusNotFound)
			return nil
		}

		var user interface{}
		err = bson.UnmarshalJSON([]byte(ctx.Request.Body()), &user)
		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		c := session.DB("travels").C("users")
		err = c.Update(bson.M{"_id": userId}, bson.M{"$set": &user})

		if err != nil {
			utils.ResponseWithJSON(ctx, []byte(""), http.StatusBadRequest)
			return nil
		}

		utils.ResponseWithJSON(ctx, []byte("{}"), http.StatusOK)
		return nil
	}
}
