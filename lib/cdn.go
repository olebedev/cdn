package cdn

import (
	"github.com/go-martini/martini"
	"io/ioutil"
	"labix.org/v2/mgo"
	"log"
)

type Config struct {
	Prefix  string
	MaxSize int
	DB      *mgo.Database
	Log     *log.Logger
}

var conf Config

func Cdn(c Config) func(r martini.Router) {
	conf = c
	// Default value
	if conf.MaxSize == 0 {
		conf.MaxSize = 1000
	}

	if conf.DB == nil {
		panic("cdn: MongoDB instance not found.")
	}

	if conf.Log == nil {
		conf.Log = log.New(ioutil.Discard, "", 0)
	}

	return func(r martini.Router) {
		r.Post("/:coll", post)
		r.Get("/:coll/stats-for", getStat)
		r.Get("/:coll/:_id", get)
		r.Get("/:coll/:_id/:file", get)
	}
}
