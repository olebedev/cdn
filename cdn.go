package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/olebedev/config"
	"labix.org/v2/mgo"
)

type Settings struct {
	Port      int
	Prefix    string
	MongoUri  string
	MongoName string
	DB        mgo.Database
	MaxSize   int
}

var conf = Settings{}

var co, _ = config.ParseYaml(`
port: 5000
prefix: ""
mongo:
  uri: localhost
  name: cdn
maxSize: 1000
`)

func main() {
	// write pid file
	savePid()
	// parse system env and then args
	co.Env().Flag()

	conf.Port, _ = co.Int("port")
	conf.Prefix, _ = co.String("prefix")
	conf.MongoUri, _ = co.String("mongo.uri")
	conf.MongoName, _ = co.String("mongo.name")
	conf.MaxSize, _ = co.Int("maxSize")

	sess, err := mgo.Dial(conf.MongoUri)
	if err != nil {
		panic(err)
	}
	conf.DB = *sess.DB(conf.MongoName)

	router := new(mux.Router)
	router.HandleFunc(fmt.Sprintf("%s/{coll}", conf.Prefix), post).Methods("POST")
	router.HandleFunc(fmt.Sprintf("%s/{coll}/stats-for", conf.Prefix), getStat).Methods("GET")
	router.HandleFunc(fmt.Sprintf("%s/{coll}/{_id}", conf.Prefix), get).Methods("GET")

	fmt.Printf("Start listening http at %d port.\n", conf.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", conf.Port), router))
}
