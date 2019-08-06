package gogm

import (
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestDropAllIndexesAndConstraints(t *testing.T){
	//requires connection
	if !testing.Short(){
		t.SkipNow()
		return
	}

	err := dsl.Init(&dsl.ConnectionConfig{
		Username: "neo4j",
		Password: "password",
		Host: "0.0.0.0",
		Port: 7687,
		PoolSize: 15,
	})
	require.Nil(t, err)

	err = dropAllIndexesAndConstraints()
	require.Nil(t, err)

	constraintRows, err := dsl.QB(true).Cypher("CALL db.constraints").Query(nil)
	require.Nil(t, err)

	found, _, err := constraintRows.All()
	require.Nil(t, err)

	require.Equal(t, 0, len(found))

	indexRows, err := dsl.QB(true).Cypher("CALL db.indexes()").Query(nil)
	require.Nil(t, err)

	iFound, _, err := indexRows.All()
	require.Nil(t, err)

	require.Equal(t, 0, len(iFound))
}

func TestIndexManagement(t *testing.T){
	//requires connection
	if !testing.Short(){
		t.SkipNow()
		return
	}

	req := require.New(t)

	//init
	err := dsl.Init(&dsl.ConnectionConfig{
		Username: "neo4j",
		Password: "password",
		Host: "0.0.0.0",
		Port: 7687,
		PoolSize: 15,
	})
	req.Nil(err)

	//delete everything
	req.Nil(dropAllIndexesAndConstraints())

	//setup structure
	mapp := toHashmapStructdecconf(map[string]structDecoratorConfig{
		"TEST1": {
			Label:"Test1",
			IsVertex: true,
			Fields: map[string]decoratorConfig{
				"UUID": {
					Name: "uuid",
					PrimaryKey: true,
					Type: reflect.TypeOf(""),
				},
				"IndexField": {
					Name: "index_field",
					Index: true,
					Type: reflect.TypeOf(1),
				},
				"UniqueField": {
					Name: "unique_field",
					Unique: true,
					Type: reflect.TypeOf(""),
				},
			},
		},
		"TEST2": {
			Label: "Test2",
			IsVertex: true,
			Fields: map[string]decoratorConfig{
				"UUID": {
					Name: "uuid",
					PrimaryKey: true,
					Type: reflect.TypeOf(""),
				},
				"IndexField1": {
					Name: "index_field1",
					Index: true,
					Type: reflect.TypeOf(1),
				},
				"UniqueField1": {
					Name: "unique_field1",
					Unique: true,
					Type: reflect.TypeOf(""),
				},
			},
		},
	})

	//create stuff
	req.Nil(createAllIndexesAndConstraints(mapp))

	t.Log("created indices and constraints")

	//validate
	req.Nil(verifyAllIndexesAndConstraints(mapp))

	//clean up
	req.Nil(dropAllIndexesAndConstraints())
}