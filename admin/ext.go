package admin

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mholt/caddy/app"
	"github.com/mholt/caddy/config/setup"
	"github.com/mholt/caddy/middleware/extensions"
)

func init() {
	router.GET("/:addr/ext", auth(extensionsGet))
	router.POST("/:addr/ext", auth(extensionsCreate))
	router.DELETE("/:addr/ext", auth(extensionsDelete))

	router.PUT("/:addr/ext/extensions", auth(extensionsSet))
	router.POST("/:addr/ext/extensions", auth(extensionsAdd))
	router.DELETE("/:addr/ext/extensions", auth(extensionsRemove))
}

// extensionsGet serializes the ext middleware out to the client to view.
func extensionsGet(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ext := getExt(w, r, p)
	if ext == nil {
		return
	}
	respondJSON(w, r, ext, http.StatusOK)
}

// extensionsCreate creates a new ext middleware and chains it into a virtualhost.
func extensionsCreate(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	status, err, ok := createMiddleware(r, p, "ext", setup.Ext)
	if !ok {
		handleError(w, r, status, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// extensionsDelete deletes the ext middleware from a virtualhost.
func extensionsDelete(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	status, err, ok := deleteMiddleware(p, "ext")
	if !ok {
		handleError(w, r, status, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// extensionsSet sets the list of extensions to the list in the body.
// Syntax:
// [".ext1", ".ext2", ...]
func extensionsSet(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ext := getExt(w, r, p)
	if ext == nil {
		return
	}

	var extList []string
	err := json.NewDecoder(r.Body).Decode(&extList)
	if err != nil {
		handleError(w, r, http.StatusBadRequest, err)
		return
	}

	app.ServersMutex.Lock()
	ext.Extensions = extList
	app.ServersMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

// extensionsAdd adds extensions.
// Syntax:
// [".ext1", ".ext2", ...]
func extensionsAdd(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ext := getExt(w, r, p)
	if ext == nil {
		return
	}

	var extList []string
	err := json.NewDecoder(r.Body).Decode(&extList)
	if err != nil {
		handleError(w, r, http.StatusBadRequest, err)
		return
	}

	app.ServersMutex.Lock()
	ext.Extensions = append(ext.Extensions, extList...)
	app.ServersMutex.Unlock()

	w.WriteHeader(http.StatusCreated)
}

// extensionsRemove deletes extensions.
// Syntax:
// [".ext1", ".ext2", ...]
func extensionsRemove(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ext := getExt(w, r, p)
	if ext == nil {
		return
	}

	var extList []string
	err := json.NewDecoder(r.Body).Decode(&extList)
	if err != nil {
		handleError(w, r, http.StatusBadRequest, err)
		return
	}

	app.ServersMutex.Lock()
	for i, extension := range ext.Extensions {
		for _, extension2 := range extList {
			if extension == extension2 {
				ext.Extensions = append(ext.Extensions[:i], ext.Extensions[i+1:]...)
				break
			}
		}
	}
	app.ServersMutex.Unlock()

	w.WriteHeader(http.StatusOK)
}

// getExt gets the extensions middleware asked for by the request.
// This function DOES handle errors if they occur, in which case return value is nil.
func getExt(w http.ResponseWriter, r *http.Request, p httprouter.Params) *extensions.Ext {
	vh := virtualHost(p.ByName("addr"))
	if vh == nil {
		handleError(w, r, http.StatusNotFound, nil)
		return nil
	}
	ext, ok := vh.Config.HandlerMap["ext"].(*extensions.Ext)
	if ext == nil {
		handleError(w, r, http.StatusNotFound, nil)
		return nil
	}
	if !ok {
		handleError(w, r, http.StatusInternalServerError, errors.New("Not ext middleware"))
		return nil
	}
	return ext
}
