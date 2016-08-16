package cdn

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
)

const FORMAT = "Mon, 2 Jan 2006 15:04:05 GMT"

func (c *Config) Get(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db := c.Db
	// validate _id
	if !bson.IsObjectIdHex(vars["_id"]) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// define main variables
	_id := bson.ObjectIdHex(vars["_id"])
	query := db.C(vars["coll"] + ".files").FindId(_id)
	meta := bson.M{}
	err := query.One(&meta)

	// found file or not
	if err != nil {
		if err.Error() == "not found" {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		return
	}

	uploadDate := meta["uploadDate"].(time.Time)
	contentType := meta["contentType"].(string)
	fileName := meta["filename"].(string)

	req.ParseForm()
	head := w.Header()
	head.Add("Accept-Ranges", "bytes")
	head.Add("ETag", vars["_id"]+"+"+req.URL.RawQuery)
	head.Add("Date", uploadDate.Format(FORMAT))
	head.Add("Last-Modified", uploadDate.Format(FORMAT))
	// Expires after ten years :)
	head.Add("Expires", uploadDate.Add(87600*time.Hour).Format(FORMAT))
	head.Add("Cache-Control", "public, max-age=31536000")
	head.Add("Content-Type", contentType)
	if _, dl := req.Form["dl"]; (contentType == "application/octet-stream") || dl {
		head.Add("Content-Disposition", "attachment; filename='"+fileName+"'")
	}

	// already served
	if h := req.Header.Get("If-None-Match"); h == vars["_id"]+"+"+req.URL.RawQuery {
		w.WriteHeader(http.StatusNotModified)
		w.Write([]byte("304 Not Modified"))
		return
	}

	// get file
	file, err := db.GridFS(vars["coll"]).OpenId(_id)
	defer file.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	// check to crop/resize
	cr, isCrop := req.Form["crop"]
	rsz, isResize := req.Form["resize"]

	isIn := ^in([]string{"image/png", "image/jpeg"}, file.ContentType()) != 0

	if isCrop && isIn && cr != nil {
		parsed, _ := parseParams(cr[0])
		if parsed != nil {
			crop(w, file, c.MaxSize, parsed)
			return
		}
	} else if isResize && isIn && rsz != nil {
		parsed, _ := parseParams(rsz[0])
		if parsed != nil {
			resize(w, file, parsed)
			return
		}
	} else {
		io.Copy(w, file)
	}

}

func (c *Config) GetStat(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db := c.Db
	req.ParseForm()

	// make query & ensure index for meta
	keys := []string{}
	q := bson.M{}
	for k, v := range req.Form {
		key := "metadata." + k
		value := strings.Join(v, "")
		if len(value) > 0 {
			keys = append(keys, key)
			q[key] = value
		}
	}

	// async ensure index
	go func() {
		if len(keys) > 0 {
			db.C(vars["coll"] + ".files").EnsureIndexKey(keys...)
		}
	}()

	pipe := db.C(vars["coll"] + ".files").Pipe([]bson.M{
		{"$match": q},
		{"$group": bson.M{
			"_id":      nil,
			"fileSize": bson.M{"$sum": "$length"},
		}},
	})

	result := bson.M{}
	pipe.One(&result)
	delete(result, "_id")
	w.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(result)
	w.Write(bytes)
}

type Doc struct {
	Id       bson.ObjectId `bson:"_id"`
	Filename string        `bson:"filename"`
}

func (d Doc) Join() string {
	return "/" + d.Id.Hex() + "/" + d.Filename
}

func (c *Config) GetIndex(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db := c.Db
	req.ParseForm()
	// make query & ensure index for meta
	keys := []string{}
	q := bson.M{}
	for k, v := range req.Form {
		key := "metadata." + k
		value := strings.Join(v, "")
		if len(value) > 0 {
			keys = append(keys, key)
			q[key] = value
		}
	}

	// async ensure index
	go func() {
		if len(keys) > 0 {
			db.C(vars["coll"] + ".files").EnsureIndexKey(keys...)
		}
	}()

	result := []Doc{}
	db.C(vars["coll"] + ".files").Find(q).All(&result)
	names := make([]string, len(result))
	for i, _ := range names {
		names[i] = result[i].Join()
	}
	w.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(names)
	w.Write(bytes)
}
