package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"labix.org/v2/mgo/bson"
)

func get(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	// validate _id
	d, e := hex.DecodeString(vars["_id"])
	if e != nil || len(d) != 12 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request. Error: " + e.Error()))
		return
	}

	// define main variables
	_id := bson.ObjectIdHex(vars["_id"])
	query := conf.DB.C(vars["coll"] + ".files").FindId(_id)
	meta := bson.M{}
	err := query.One(&meta)

	// found file or not
	if err != nil {
		if err.Error() == "not found" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("404 Not Found"))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("400 Bad Request. Error: " + err.Error()))
		}
		return
	}

	// check cache headers & response
	tt := meta["uploadDate"].(time.Time)
	fmt.Printf("GET /%s/%s (%s)\n", vars["coll"], vars["_id"], tt.Format(time.RFC822))

	if h := req.Header.Get("If-Modified-Since"); h == tt.Format(time.RFC822) {
		w.WriteHeader(http.StatusNotModified)
		w.Write([]byte("304 Not Modified"))
		return
	}

	// get file
	file, err := conf.DB.GridFS(vars["coll"]).OpenId(_id)
	if err != nil {
		log.Fatal(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}

	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("ETag", vars["_id"])
	w.Header().Add("Date", file.UploadDate().Format(time.RFC822))
	w.Header().Add("Last-Modified", file.UploadDate().Format(time.RFC822))
	w.Header().Add("Cache-Control", "public, max-age=31536000")
	w.Header().Add("Content-Type", file.ContentType())
	// w.Header().Add("Content-Length", fmt.Sprintf("%d", file.Size()))
	if file.ContentType() == "application/octet-stream" {
		w.Header().Add("Content-Disposition", "attachment; filename='"+file.Name()+"'")
	}

	// check to crop/resize
	req.ParseForm()
	crop, isCrop := req.Form["crop"]
	resize, isResize := req.Form["resize"]

	isIn := ^in([]string{"image/png", "image/jpeg"}, file.ContentType()) != 0

	if isCrop && isIn && crop != nil {
		parsed := parseParams(crop[0])
		if parsed != nil {
			fmt.Println("crop", parsed)
			err = Crop(w, file, parsed)
			if err != nil {
				fmt.Println("GET err:", err.Error())
			}
			return
		}
	} else if isResize && isIn && resize != nil {
		parsed := parseParams(resize[0])
		if parsed != nil {
			fmt.Println("resize", parsed)
			err = Resize(w, file, parsed)
			if err != nil {
				fmt.Println("GET err:", err.Error())
			}
			return
		}
	} else {
		io.Copy(w, file)
	}
}
