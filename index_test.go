package gogm

import (
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
	"github.com/stretchr/testify/require"
	"reflect"
)

func testIndexManagement(req *require.Assertions) {
	//init
	conn, err := driverPool.Open(driver.ReadWriteMode)
	req.Nil(err)

	defer driverPool.Reclaim(conn)
	req.Nil(err)

	//delete everything
	req.Nil(dropAllIndexesAndConstraints())

	//setup structure
	mapp := toHashmapStructdecconf(map[string]structDecoratorConfig{
		"TEST1": {
			Label:    "Test1",
			IsVertex: true,
			Fields: map[string]decoratorConfig{
				"UUID": {
					Name:       "uuid",
					PrimaryKey: true,
					Type:       reflect.TypeOf(""),
				},
				"IndexField": {
					Name:  "index_field",
					Index: true,
					Type:  reflect.TypeOf(1),
				},
				"UniqueField": {
					Name:   "unique_field",
					Unique: true,
					Type:   reflect.TypeOf(""),
				},
			},
		},
		"TEST2": {
			Label:    "Test2",
			IsVertex: true,
			Fields: map[string]decoratorConfig{
				"UUID": {
					Name:       "uuid",
					PrimaryKey: true,
					Type:       reflect.TypeOf(""),
				},
				"IndexField1": {
					Name:  "index_field1",
					Index: true,
					Type:  reflect.TypeOf(1),
				},
				"UniqueField1": {
					Name:   "unique_field1",
					Unique: true,
					Type:   reflect.TypeOf(""),
				},
			},
		},
	})

	//create stuff
	req.Nil(createAllIndexesAndConstraints(mapp))

	log.Println("created indices and constraints")

	//validate
	req.Nil(verifyAllIndexesAndConstraints(mapp))

	//clean up
	req.Nil(dropAllIndexesAndConstraints())
}
