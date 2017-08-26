package importer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	mgo "gopkg.in/mgo.v2"
)

const zipPath = "/tmp/data"
const dataPath = "/tmp/extract"

func Import() {
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	session.SetMode(mgo.Monotonic, true)

	err = unzip(fmt.Sprintf("%s/%s", zipPath, "data.zip"), dataPath)
	if err != nil {
		log.Fatal(err)
	}

	err = importData(session)
	if err != nil {
		log.Fatal(err)
	}

	err = ensureIndexes(session)
	if err != nil {
		log.Fatal(err)
	}
}

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err = os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func importData(s *mgo.Session) error {
	session := s.Copy()
	defer session.Close()

	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		return err
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

	return nil
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

	dataCollection := s.DB("travels").C(collection)
	err = dataCollection.Insert(importData[collection]...)

	return err
}

func ensureIndexes(s *mgo.Session) error {
	c := s.DB("travels").C("users")
	err := c.EnsureIndexKey("id")
	if err != nil {
		return err
	}

	c = s.DB("travels").C("locations")
	err = c.EnsureIndexKey("id")
	if err != nil {
		return err
	}

	c = s.DB("travels").C("visits")
	err = c.EnsureIndexKey("id")
	if err != nil {
		return err
	}

	err = c.EnsureIndex(mgo.Index{
		Key: []string{"user", "location"},
	})

	if err != nil {
		return err
	}

	err = c.EnsureIndex(mgo.Index{
		Key: []string{"location", "visited_at", "user"},
	})

	if err != nil {
		return err
	}

	return nil
}
