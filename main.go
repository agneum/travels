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
	router.Get(`/locations/<id:\d+>`, handlers.GetLocation(session))
	router.Get(`/visits/<id:\d+>`, handlers.GetVisit(session))
	router.Get(`/users/<id:\d+>/visits`, handlers.GetUserVisit(session))
	router.Get(`/locations/<id:\d+>/avg`, handlers.GetAverageMark(session))
	router.Post("/users/new", handlers.CreateUser(session))
	router.Post(`/users/<id:\d+>`, handlers.UpdateUser(session))

	panic(fasthttp.ListenAndServe(":80", router.HandleRequest))
}
