package gogm

import (
	dsl "github.com/mindstand/go-cypherdsl"
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
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

	conn, err := driverPool.Open(driver.ReadWriteMode)
	if err != nil {
		require.Nil(t, err)
	}
	defer conn.Close()
	require.Nil(t, err)

	err = dropAllIndexesAndConstraints()
	require.Nil(t, err)

	constraintRows, err := dsl.QB().WithNeo(conn).Cypher("CALL db.constraints").Query(nil)
	require.Nil(t, err)

	found, _, err := constraintRows.All()
	require.Nil(t, err)

	require.Equal(t, 0, len(found))

	indexRows, err := dsl.QB().WithNeo(conn).Cypher("CALL db.indexes()").Query(nil)
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
	conn, err := driverPool.Open(driver.ReadWriteMode)
	if err != nil {
		require.Nil(t, err)
	}
	defer conn.Close()
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