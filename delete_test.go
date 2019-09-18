package gogm

import (
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDelete(t *testing.T){
	if !testing.Short(){
		t.Skip()
		return
	}
	conn, err := driverPool.Open(driver.ReadWriteMode)
	if err != nil {
		require.Nil(t, err)
	}
	driverPool.Reclaim(conn)

	del := a{
		Id: 0,
		UUID: "5334ee8c-6231-40fd-83e5-16c8016ccde6",
	}

	err = deleteNode(conn, &del)
	require.Nil(t, err)
}
