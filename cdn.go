package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"labix.org/v2/mgo"
)

// *** Configuration
type Config struct {
	Port      int64
	Prefix    string
	MongoUri  string
	MongoName string
	DB        mgo.Database
	MaxSize   int
}

var conf = Config{
	Port:      5000,
	Prefix:    "",
	MongoUri:  "localhost",
	MongoName: "gotest",
	MaxSize:   1000,
}

func main() {
	// write pid file
	savePid()

	// get config vars from flags or ENV
	if os.Getenv("PORT") != "" {
		port, _ := strconv.ParseInt(os.Getenv("PORT"), 10, 16)
		conf.Port = port
	}
	if os.Getenv("MONGO_URI") != "" {
		conf.MongoUri = os.Getenv("MONGO_URI")
	}
	if os.Getenv("MONGO_NAME") != "" {
		conf.MongoName = os.Getenv("MONGO_NAME")
	}

	if os.Getenv("PREFIX") != "" {
		conf.Prefix = os.Getenv("PREFIX")
	}

	port := flag.Int64("port", conf.Port,
		fmt.Sprintf("Specify a TCP/IP port number. Default is `%d`.", conf.Port))
	prefix := flag.String("prefix", conf.Prefix,
		fmt.Sprintf("Specify URI prefix. Default is `%s`.", conf.Prefix))
	mongouri := flag.String("mongouri", conf.MongoUri,
		fmt.Sprintf("Specify MONGOURI. Default is `%s`.", conf.MongoUri))
	mongoname := flag.String("mongoname", conf.MongoName,
		fmt.Sprintf("Specify MONGONAME. Default is `%s`.", conf.MongoName))

	flag.Parse()

	conf.Port = *port
	conf.Prefix = *prefix
	conf.MongoUri = *mongouri
	conf.MongoName = *mongoname

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
