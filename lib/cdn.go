package cdn

import (
	"github.com/go-martini/martini"
)

type Config struct {
	MaxSize  int
	TailOnly bool
	ShowInfo bool
}

var conf Config

func Cdn(c Config) func(r martini.Router) {
	conf = c
	if conf.MaxSize == 0 {
		conf.MaxSize = 1000
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
