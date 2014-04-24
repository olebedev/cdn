package main

import (
	"github.com/go-martini/martini"
	"github.com/olebedev/cdn/lib"
	"github.com/olebedev/config"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"os"
)

var conf, _ = config.ParseJson(`
{
  "port": 5000,
  "prefix": "",
  "maxSize": 1000,
  "mongo":{
    "uri": "localhost",
    "name": "cdn"
  }
}
`)

func main() {
	conf.Env().Flag()
	r := martini.NewRouter()
	m := martini.New()
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	session, err := mgo.Dial(conf.UString("mongo.uri"))
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	db := session.DB(conf.UString("mongo.name"))

	logger := log.New(os.Stdout, "\x1B[36m[cdn] >>\x1B[39m ", 0)

	fn := cdn.Cdn(cdn.Config{
		DB:     db,
		Prefix: conf.UString("prefix"),
		Log:    logger,
	})
	fn(r)

	logger.Println("Server started at http://localhost:" + conf.UString("port", "5000"))
	_err := http.ListenAndServe("localhost:"+conf.UString("port", "5000"), m)
	if _err != nil {
		logger.Printf("\x1B[31mServer exit with error: %s\x1B[39m\n", _err)
		os.Exit(1)
	}
}
