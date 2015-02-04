package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-martini/martini"
	"github.com/olebedev/cdn/lib"
	"github.com/olebedev/config"
	"labix.org/v2/mgo"
)

var conf, _ = config.ParseYaml(`
debug: false
port: 5000
maxSize: 1000
showInfo: true
tailOnly: false
mongo:
  uri: localhost
mongodb:
  database: cdn
`)

func main() {
	conf.Env().Flag()
	r := martini.NewRouter()
	m := martini.New()
	if conf.UBool("debug") {
		m.Use(martini.Logger())
	}
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	session, err := mgo.Dial(conf.UString("mongo.uri"))
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	db := session.DB(conf.UString("mongodb.database"))

	logger := log.New(os.Stdout, "\x1B[36m[cdn] >>\x1B[39m ", 0)
	m.Map(logger)
	m.Map(db)

	r.Group("", cdn.Cdn(cdn.Config{
		MaxSize:  conf.UInt("maxSize"),
		ShowInfo: conf.UBool("showInfo"),
		TailOnly: conf.UBool("tailOnly"),
	}))

	logger.Println("Server started at :" + conf.UString("port", "5000"))
	_err := http.ListenAndServe(":"+conf.UString("port", "5000"), m)
	if _err != nil {
		logger.Printf("\x1B[31mServer exit with error: %s\x1B[39m\n", _err)
		os.Exit(1)
	}
}
