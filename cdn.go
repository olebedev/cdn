package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	mgo "gopkg.in/mgo.v2"

	"github.com/go-martini/martini"
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
	r := martini.NewRouter()
	m := martini.New()
	if conf.UBool("debug") {
		m.Use(martini.Logger())
	}
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)

	fmt.Println(mgo.ParseURL(conf.UString("mongo.url")))
	session, err := mgo.Dial(conf.UString("mongo.url"))
	if err != nil {
		panic(err)
	}
	// session.SetMode(mgo.Monotonic, true)
	db := session.DB("cdn")

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
