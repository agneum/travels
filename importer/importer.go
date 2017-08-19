package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	mgo "gopkg.in/mgo.v2"
)

const dataPath = "/tmp/data"

func main() {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	importData(session)

}

func importData(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		log.Printf("%+v\n", err.Error())
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) != ".json" {
			continue
		}
		err := importFile(session, f.Name())
		if err != nil {
			log.Printf("%+v\n", err.Error())
		}
		fmt.Println(f.Name())
	}
}

func importFile(s *mgo.Session, filename string) error {
	raw, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", dataPath, filename))
	if err != nil {
		return err
	}

	splittedString := strings.Split(filename, "_")
	collection := splittedString[0]

	var importData map[string][]interface{}
	err = json.Unmarshal(raw, &importData)
	if err != nil {
		return err
	}

	users := s.DB("travels").C(collection)
	err = users.Insert(importData[collection]...)

	return err
}
