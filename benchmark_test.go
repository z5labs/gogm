package gogm

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSessionV2Impl_LoadAll(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	req := require.New(t)
	conf := Config{
		Username:  "neo4j",
		Password:  "changeme",
		Host:      "0.0.0.0",
		IsCluster: false,
		Port:      7687,
		PoolSize:  15,
		// this is ignore because index management is part of the test
		IndexStrategy:             IGNORE_INDEX,
		EnableDriverLogs:          true,
		DefaultTransactionTimeout: 2 * time.Minute,
	}

	gogm, err := New(&conf, UUIDPrimaryKeyStrategy, &a{}, &b{}, &c{}, &propTest{})
	req.NotNil(gogm)
	req.Nil(err)

	sess, err := gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
	req.NotNil(sess)
	req.Nil(err)

	objs := []*a{
		{
			PropsTest2: []string{"1"},
		},
		{
			PropsTest2: []string{"2"},
		},
	}

	// create 2 objects
	req.Nil(sess.ManagedTransaction(context.Background(), func(tx TransactionV2) error {
		err := tx.SaveDepth(context.Background(), objs[0], 0)
		if err != nil {
			return err
		}

		return tx.SaveDepth(context.Background(), objs[1], 0)
	}))

	var writeTo []*a
	req.Nil(sess.LoadAllDepth(context.Background(), &writeTo, 0))

	req.Nil(sess.Delete(context.Background(), writeTo[0]))
	req.Nil(sess.Delete(context.Background(), writeTo[1]))
}
