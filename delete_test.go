package gogm

import (
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDelete(t *testing.T){
	if testing.Short(){
		t.Skip()
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

	del := a{
		Id: 0,
		UUID: "5334ee8c-6231-40fd-83e5-16c8016ccde6",
	}

	err = deleteNode(dsl.NewSession(), &del)
	require.Nil(t, err)
}
