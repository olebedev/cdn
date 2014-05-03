package cdn

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-martini/martini"
	"labix.org/v2/mgo/bson"
)

func get(w http.ResponseWriter, req *http.Request, vars martini.Params) {
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
	conf.Log.Printf("GET %s/%s (%s) ", vars["coll"], vars["_id"], tt.Format(time.RFC822))

	if h := req.Header.Get("If-Modified-Since"); h == tt.Format(time.RFC822) {
		w.WriteHeader(http.StatusNotModified)
		w.Write([]byte("304 Not Modified"))
		conf.Log.Println("304")
		return
	}

	// get file
	file, err := conf.DB.GridFS(vars["coll"]).OpenId(_id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}

	req.ParseForm()
	w.Header().Add("Accept-Ranges", "bytes")
	w.Header().Add("ETag", vars["_id"])
	w.Header().Add("Date", file.UploadDate().Format(time.RFC822))
	w.Header().Add("Last-Modified", file.UploadDate().Format(time.RFC822))
	w.Header().Add("Cache-Control", "public, max-age=31536000")
	w.Header().Add("Content-Type", file.ContentType())
	// w.Header().Add("Content-Length", conf.Log.Sprintf("%d", file.Size()))
	_, dl := req.Form["dl"]
	if (file.ContentType() == "application/octet-stream") || dl {
		w.Header().Add("Content-Disposition", "attachment; filename='"+file.Name()+"'")
	}

	// check to crop/resize
	cr, isCrop := req.Form["crop"]
	rsz, isResize := req.Form["resize"]

	isIn := ^in([]string{"image/png", "image/jpeg"}, file.ContentType()) != 0

	if isCrop && isIn && cr != nil {
		parsed := parseParams(cr[0])
		if parsed != nil {
			err = crop(w, file, parsed)
			conf.Log.Println("croped for:", parsed)
			if err != nil {
				conf.Log.Println("GET err:", err.Error())
			}
			return
		}
	} else if isResize && isIn && rsz != nil {
		parsed := parseParams(rsz[0])
		if parsed != nil {
			err = resize(w, file, parsed)
			conf.Log.Println("resized for:", parsed)
			if err != nil {
				conf.Log.Println("GET err:", err.Error())
			}
			return
		}
	} else {
		conf.Log.Println("as is")
		io.Copy(w, file)
	}
}

func getStat(w http.ResponseWriter, req *http.Request, vars martini.Params) {
	req.ParseForm()

	// if len(req.Form) < 1 {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	w.Write([]byte("Сriteria not found!"))
	// 	return
	// }

	// make query & ensure index for meta
	keys := []string{}
	q := bson.M{}
	for k, v := range req.Form {
		key := fmt.Sprintf("metadata.%s", k)
		value := strings.Join(v, "")
		if len(value) > 0 {
			keys = append(keys, key)
			q[key] = value
		}
	}

	// async ensure index
	go func() {
		if len(keys) > 0 {
			conf.DB.C(vars["coll"] + ".files").EnsureIndexKey(keys...)
		}
	}()

	pipe := conf.DB.C(vars["coll"] + ".files").Pipe([]bson.M{
		{"$match": q},
		{"$group": bson.M{
			"_id":      nil,
			"fileSize": bson.M{"$sum": "$length"},
			// "files":    bson.M{"$push": "$_id"},
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

func getIndex(w http.ResponseWriter, req *http.Request, vars martini.Params) {
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
			conf.DB.C(vars["coll"] + ".files").EnsureIndexKey(keys...)
		}
	}()

	result := []Doc{}
	conf.DB.C(vars["coll"] + ".files").Find(q).All(&result)
	names := make([]string, len(result))
	for i, _ := range names {
		names[i] = result[i].Join()
	}
	w.WriteHeader(http.StatusOK)
	bytes, _ := json.Marshal(names)
	w.Write(bytes)
}
