#CDN over MongoDb GridFs

This utility can be used as stand alone _Content Delivery Network_, using _MongoDB GridFs_ as backend file storage. It can be builded from source code or installed as already compiled binaries.  
 
Also it can be used as a thin file storage library for you projects, based on [martini](https://github.com/go-martini/martini) framework. For example, when you using one of the cloud platforms like _Heroku_, with ephemeral file system for application's instances and  you have to save user's data.

## Features

- on the fly crop and resize for `image/png` and `image/jpeg` _mimetypes_
- cache strategy, based at _HTTP Last-Modified_ header
- additional metadata for each files & aggregated statistic for it
- forced _HTTP Content-Disposition_ header with file name, for download links(only if flag is specified, see below)
- buckets(_MongoDB_ collections) as separation level
- file listings for buckets, queried by metadata or without

## Examples
Let's assume that the process is already running and listening to default `http://localhost:5000`. 
 
#### Uploading
~~~ bash
$ curl -F field=@./books.jpg http://localhost:5000/example
{
  "error": null,
  "field":"/example/5364d634952b829316000001/books.jpg"
}
~~~
Uploading with metadata is realy simple - you only should specify it as _GET_ parameters for uploading _URL_:
~~~ bash
$ curl -F field=@./books.jpg "http://localhost:5000/example?userId=1&some_another_data=useful"
...
~~~

#### Getting
As expected, the _URL_ for getting is:  
~~~ bash
http://localhost:5000/example/5364d634952b829316000001/books.jpg 
~~~
or   
~~~ bash
http://localhost:5000/example/5364d634952b829316000001
~~~  
That means that the filename is not necessary.  
For forsed downloading specify `dl` _GET_ parameter:
~~~ bash
http://localhost:5000/example/5364d634952b829316000001/books.jpg?dl
~~~  
In this case the file will not be previewed in the browser.

#### Crop and Resize images
> This works only for files with mimetypes `image/png` & `image/jpeg`!
> In another cases it feature will be ignored.

Specify _GET_ parameters `crop` or `resize` for _URL_. Crop example:  
~~~ bash
http://localhost:5000/example/5364d634952b829316000001/books.jpg?crop=500
~~~  
The value should contain one or two(separadet by one non-digit character) integer as width and height in pixels. If height is not specified, it will be used width value. For example, value `crop=500`  will be interpreted as `crop=500x500`.  

`resize` parameter works the same way.

#### Aggregation and the listing of files

To get storage usage information in bytes, based on saved metadata, _GET_ it like this:
~~~ bash
$ http://localhost:5000/example/_stats?userId=1
{
  "fileSize": 204789
}
~~~  
If metadata is not specified, it will be received usage information for whole bucket.

To get the listing of files, based on saved metadata, _GET_ it like this:
~~~ bash
$ http://localhost:5000/example?userId=1
[
  "/5364d634952b829316000001/books.jpg"
]
~~~  
If metadata is not specified, it will be received the listing of files for whole bucket.

## Usage 
As library for [martini](https://github.com/go-martini/martini) framework.

~~~ bash
$ go get github.com/olebedev/cdn
~~~

Simple `server.go` file:

~~~ go
package main

import (
	"log"
	"net/http"
	"os"
	"github.com/go-martini/martini"
	"github.com/olebedev/cdn/lib"
	"labix.org/v2/mgo"
)

// Acceess handler
func Access(res http.ResponseWriter, req *http.Request) {
	// check session or something like this
}

// Set prefix for current collection name, of course, if need it
// It useful when you store another data in one database
func AutoPrefix(params martini.Params) {
	// 'coll' - parameter name, which is used
	params["coll"] = "cdn." + params["coll"]
	// Ok. Now, cdn will work with this prefix for all collections
}

func main() {
	m := martini.Classic()

	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("cdn")
	m.Map(db)

	logger := log.New(os.Stdout, "\x1B[36m[cdn] >>\x1B[39m ", 0)
	m.Map(logger)

	m.Group("/uploads", 
		cdn.Cdn(cdn.Config{
			// Maximum width or height with pixels to crop or resize
			// Useful to high performance
			MaxSize:  1000,
			// Show statictics and the listing of files
			ShowInfo: true,
			// If true it send URL without collection name, like this:
			// {"field":"/5364d634952b829316000001/books.jpg", "error": null}
			TailOnly: true,
		}),
		// Access logic here
		Access,
		// On the fly prefix for collection
		AutoPrefix,
	)

	logger.Println("Server started at :3000")
	m.Run()
}
~~~
Let's start it!
~~~ bash
$ go run server.go
[cdn] >> Server started at :3000
~~~

That's all. Now you have started CDN at `http://localhost:3000/uploads/`.

## Installation as stand alone

If you want to build it from sources:
~~~ bash
$ go get github.com/olebedev/cdn
~~~

If you don't now what is _Golang_, check [releases](https://github.com/olebedev/cdn/releases) page and download binaries for your platform. Untar it and type this:  
~~~ bash
$ ./cdn --help
Usage of ./cdn:
  -maxSize="1000": 
  -mongo.name="cdn": 
  -mongo.uri="localhost": 
  -port="5000": 
  -showInfo="true": 
  -tailOnly="false":
~~~


##### TODO:

- handler for 206 HTTP Status for large file strimming
- cache(save to GridFs croppped & resized image files)


