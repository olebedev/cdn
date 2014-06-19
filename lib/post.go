package cdn

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/go-martini/martini"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

func post(w http.ResponseWriter, req *http.Request, vars martini.Params, db *mgo.Database) {
	formFile, formHead, err := req.FormFile("field")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}
	defer formFile.Close()

	//remove any directory names in the filename
	//START: work around IE sending full filepath and manually get filename
	itemHead := formHead.Header["Content-Disposition"][0]
	lookfor := "filename=\""
	fileIndex := strings.Index(itemHead, lookfor)

	if fileIndex < 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}

	filename := itemHead[fileIndex+len(lookfor):]
	filename = filename[:strings.Index(filename, "\"")]

	slashIndex := strings.LastIndex(filename, "\\")
	if slashIndex > 0 {
		filename = filename[slashIndex+1:]
	}

	slashIndex = strings.LastIndex(filename, "/")
	if slashIndex > 0 {
		filename = filename[slashIndex+1:]
	}
	//END: work around IE sending full filepath

	// GridFs actions
	file, err := db.GridFS(vars["coll"]).Create(filename)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}
	defer file.Close()

	io.Copy(file, formFile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}

	b := make([]byte, 512)
	formFile.Seek(0, 0)
	formFile.Read(b)

	file.SetContentType(http.DetectContentType(b))
	file.SetMeta(req.Form)
	err = file.Close()

	_id, _ := file.Id().(bson.ObjectId)

	// json response
	field := "/" + _id.Hex() + "/" + filename
	if !conf.TailOnly {
		field = "/" + vars["coll"] + field
	}

	bytes, _ := json.Marshal(map[string]interface{}{
		"error": err,
		"field": field,
	})
	w.Write(bytes)
}
