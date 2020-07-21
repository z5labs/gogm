package main

import (
	"fmt"
	"github.com/neo4j/neo4j-go-driver/neo4j"
)

func main() {
	configForNeo4j40 := func(conf *neo4j.Config) {
		conf.Encrypted = false
	}

	driver, err := neo4j.NewDriver("bolt://0.0.0.0:7687", neo4j.BasicAuth("neo4j", "password", ""), configForNeo4j40)
	if err != nil {
		panic(err)
	}

	// handle driver lifetime based on your application lifetime requirements
	// driver's lifetime is usually bound by the application lifetime, which usually implies one driver instance per application
	defer driver.Close()

	// For multidatabase support, set sessionConfig.DatabaseName to requested database
	sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
	session, err := driver.NewSession(sessionConfig)
	if err != nil {
		panic(err)
	}
	defer session.Close()

	result, err := session.Run("CREATE (n:Item { id: $id, name: $name }) RETURN n", map[string]interface{}{
		"id":   1,
		"name": "Item 1",
	})
	if err != nil {
		panic(err)
	}

	for result.Next() {

		fmt.Printf("Created Item with Id = '%d' and Name = '%s'\n", result.Record().GetByIndex(0).(int64), result.Record().GetByIndex(1).(string))
	}
}
