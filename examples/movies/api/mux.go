package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mindstand/gogm/v2"
	"github.com/mindstand/gogm/v2/examples/movies/domain"
	"net/http"
)

func GetMux(ogm *gogm.Gogm) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/actor/{name}", ActorHandler).Methods("GET")
	r.HandleFunc("/movie/{name}", GetMovieHandler(ogm)).Methods("GET")

	return r
}

// uses global gogm
func ActorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	if name == "" {
		writeResponse(http.StatusBadRequest, map[string]interface{}{
			"error": "name is empty",
		}, w)
		return
	}

	sess, err := gogm.G().NewSessionV2(gogm.SessionConfig{AccessMode: gogm.AccessModeRead})
	if err != nil {
		writeResponse(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		}, w)
		return
	}

	defer sess.Close()

	// grab any relations they have to the movie
	query := "match p=(:Person{name:$name})-->(:Movie) return p"
	var person domain.Person
	err = sess.Query(context.Background(), query, map[string]interface{}{
		"name": name,
	}, &person)

	if err != nil {
		writeResponse(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		}, w)
		return
	}

	writeResponse(http.StatusOK, person, w)
}

// uses passed gogm
func GetMovieHandler(ogm *gogm.Gogm) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		if name == "" {
			writeResponse(http.StatusBadRequest, map[string]interface{}{
				"error": "name is empty",
			}, w)
			return
		}

		sess, err := ogm.NewSessionV2(gogm.SessionConfig{AccessMode: gogm.AccessModeRead})
		if err != nil {
			writeResponse(http.StatusInternalServerError, map[string]interface{}{
				"error": err.Error(),
			}, w)
			return
		}

		defer sess.Close()

		var movie domain.Movie
		err = sess.Query(context.Background(), "MATCH p=(director)-[:DIRECTED]->(movie:Movie {title:$name})<-[:ACTED_IN]-(actor) RETURN p", map[string]interface{}{
			"name": name,
		}, &movie)
		if err != nil {
			writeResponse(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			}, w)
			return
		}

		writeResponse(http.StatusOK, movie, w)
	}
}

func writeResponse(code int, obj interface{}, w http.ResponseWriter) error {
	w.WriteHeader(code)
	jsb, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to convert object to json, %w", err)
	}

	_, err = w.Write(jsb)
	return err
}
