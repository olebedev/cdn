package cdn

import (
	"github.com/go-martini/martini"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
)

type Config struct {
	DB       *mgo.Database
	Log      *log.Logger
	MaxSize  int
	TailOnly bool
	ShowInfo bool
}

var conf Config

func Cdn(c Config) func(r martini.Router) {
	conf = c
	// Default value
	if conf.MaxSize == 0 {
		conf.MaxSize = 1000
	}

	if conf.DB == nil {
		panic("Cdn: MongoDB connection not found.")
	}

	if conf.Log == nil {
		conf.Log = log.New(ioutil.Discard, "", 0)
	}

	return func(r martini.Router) {
		if conf.ShowInfo {
			r.Get("/:coll", getIndex)
			r.Get("/:coll/_stats", getStat)
		}
		r.Post("/:coll", post)
		r.Get("/:coll/:_id", get)
		r.Get("/:coll/:_id/:file", get)
	}
}
