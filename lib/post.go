package cdn

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
)

func (c *Config) Post(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	db := c.Db
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

	// json response
	field := "/" + file.Id().(bson.ObjectId).Hex() + "/" + filename
	if !c.TailOnly {
		field = "/" + vars["coll"] + field
	}

	bytes, _ := json.Marshal(map[string]interface{}{
		"error": err,
		"field": field,
	})
	w.Write(bytes)
}
