package main

import (
	mgo "gopkg.in/mgo.v2"

	"github.com/agneum/travels/handlers"
	"github.com/agneum/travels/importer"
	"github.com/qiangxue/fasthttp-routing"
	"github.com/valyala/fasthttp"
)

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
	router.Get(`/users/<id:\d+>`, handlers.GetUser(session))
	router.Get(`/users/<id:\d+>/visits`, handlers.GetUserVisit(session))
	router.Get(`/locations/<id:\d+>`, handlers.GetLocation(session))
	router.Get(`/locations/<id:\d+>/avg`, handlers.GetAverageMark(session))
	router.Get(`/visits/<id:\d+>`, handlers.GetVisit(session))
	router.Post(`/users/new`, handlers.CreateUser(session))
	router.Post(`/users/<id:\d+>`, handlers.UpdateUser(session))
	router.Post(`/locations/new`, handlers.CreateLocation(session))
	router.Post(`/locations/<id:\d+>`, handlers.UpdateLocation(session))
	router.Post(`/visits/new`, handlers.CreateVisit(session))
	router.Post(`/visits/<id:\d+>`, handlers.UpdateVisit(session))

	panic(fasthttp.ListenAndServe(":80", router.HandleRequest))
}
