package cdn

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-martini/martini"
	"labix.org/v2/mgo/bson"
)

func post(w http.ResponseWriter, req *http.Request, vars martini.Params) {
	formFile, formHead, err := req.FormFile("field")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 Bad Request"))
		return
	}
	defer formFile.Close()

	vars["coll"] = conf.Prefix + vars["coll"]

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
	file, err := conf.DB.GridFS(vars["coll"]).Create(filename)
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

	// guess mime type
	extIndex := strings.LastIndex(filename, ".")
	mimeType := "application/octet-stream"
	if extIndex > 0 {
		if t := mime.TypeByExtension(filename[extIndex:]); len(t) > 0 {
			mimeType = t
		}
	}

	file.SetContentType(mimeType)
	file.SetMeta(req.Form)
	file.Close()

	_id, _ := file.Id().(bson.ObjectId)

	// json response
	w.Write([]byte(
		fmt.Sprintf("{\"error\":null,\"data\":{\"field\":\"%s\"}}",
			fmt.Sprintf("%s/%s/%s/%s", conf.Prefix, vars["coll"], _id.Hex(), filename))))
}
