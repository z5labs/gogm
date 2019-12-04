package gogm

import (
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
	"github.com/stretchr/testify/require"
)

func testDelete(req *require.Assertions) {
	conn, err := driverPool.Open(driver.ReadWriteMode)
	if err != nil {
		req.Nil(err)
	}
	defer driverPool.Reclaim(conn)

	del := a{
		BaseNode: BaseNode{
			Id:   0,
			UUID: "5334ee8c-6231-40fd-83e5-16c8016ccde6",
		},
	}

	err = deleteNode(conn, &del)
	req.Nil(err)
}
