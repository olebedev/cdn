package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	mgo "gopkg.in/mgo.v2"

	"github.com/gorilla/mux"
	"github.com/olebedev/cdn/lib"
	"github.com/olebedev/config"
)

var conf, _ = config.ParseYaml(`
debug: false
port: 5000
maxSize: 1000
showInfo: true
tailOnly: false
mongo:
  url: localhost
`)

func main() {
	conf.Env().Flag()
	_c, _ := config.RenderYaml(conf.Root)
	fmt.Println("start with config:\n", _c, "\n")

	fmt.Println(mgo.ParseURL(conf.UString("mongo.url")))
	session, err := mgo.Dial(conf.UString("mongo.url"))
	if err != nil {
		panic(err)
	}

	r := mux.NewRouter()
	app := cdn.Config{
		MaxSize:  conf.UInt("maxSize"),
		ShowInfo: conf.UBool("showInfo"),
		TailOnly: conf.UBool("tailOnly"),
		Db:       session.DB(""),
	}

	if app.ShowInfo {
		r.HandleFunc("/{coll}", app.GetIndex).Methods("GET")
		r.HandleFunc("/{coll}/_stats", app.GetStat).Methods("GET")
	}

	r.HandleFunc("/{coll}", app.Post).Methods("POST")
	r.HandleFunc("/{coll}/{_id}", app.Get).Methods("GET")
	r.HandleFunc("/{coll}/{_id}/{file}", app.Get).Methods("GET")

	log.Println("Server started at :" + conf.UString("port", "5000"))
	_err := http.ListenAndServe(":"+conf.UString("port", "5000"), r)
	if _err != nil {
		log.Printf("\x1B[31mServer exit with error: %s\x1B[39m\n", _err)
		os.Exit(1)
	}
}
