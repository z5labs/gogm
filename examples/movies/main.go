package main

import (
	"context"
	"github.com/mindstand/gogm/v2"
	"github.com/mindstand/gogm/v2/examples/movies/api"
	"github.com/mindstand/gogm/v2/examples/movies/domain"
	"net/http"
)

func main() {
	// define your configuration
	config := gogm.Config{
		Host:          "0.0.0.0",
		Port:          7687,
		Username:      "neo4j",
		LogLevel: "INFO",
		Password:      "changeme",
		PoolSize:      50,
		Encrypted:     false,
		IndexStrategy: gogm.IGNORE_INDEX,
	}

	// register all vertices and edges
	// this is so that GoGM doesn't have to do reflect processing of each edge in real time
	// use nil or gogm.DefaultPrimaryKeyStrategy if you only want graph ids
	// we are using the default key strategy since our vertices are using BaseNode
	_gogm, err := gogm.New(&config, gogm.DefaultPrimaryKeyStrategy, &domain.Movie{}, &domain.Person{}, &domain.ActedInEdge{})
	if err != nil {
		panic(err)
	}

	gogm.SetGlobalGogm(_gogm)

	// we're going to make stuff so we're going to do read write
	sess, err := _gogm.NewSessionV2(gogm.SessionConfig{AccessMode: gogm.AccessModeWrite})
	if err != nil {
		panic(err)
	}

	//close the session
	defer sess.Close()

	query := `
MATCH p=(movie:Movie {title:$favorite})<-[:ACTED_IN]-(actor)
RETURN p
`
	movie := &domain.Movie{}
	err = sess.Query(context.Background(), query, map[string]interface{}{"favorite": "The Matrix"}, movie)
	if err != nil {
		panic(err)
	}

	println(len(movie.Actors))

	r := api.GetMux(_gogm)
	err = http.ListenAndServe(":8081", r)
	if err != nil {
		panic(err)
	}
}