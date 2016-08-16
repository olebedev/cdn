package cdn

import mgo "gopkg.in/mgo.v2"

type Config struct {
	MaxSize  int
	TailOnly bool
	ShowInfo bool
	Db       *mgo.Database
}
