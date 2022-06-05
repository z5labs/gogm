package gogm

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSessionV2Impl_LoadAll(t *testing.T) {
	if testing.Short() {
		t.Skip()
		return
	}

	ctx := context.Background()
	req := require.New(t)

	container, err := setupNeo4jContainer(ctx, neo4)
	req.Nil(err)
	defer container.Terminate(ctx)

	config := container.GetGogmConfig()
	// this is ignore because index management is part of the test
	config.IndexStrategy = IGNORE_INDEX
	config.CAFileLocation = filepath.Join(container.CertDir, "ca-public.crt")

	gogm, err := New(config, UUIDPrimaryKeyStrategy, &a{}, &b{}, &c{}, &propTest{})
	req.Nil(err)
	req.NotNil(gogm)

	sess, err := gogm.NewSessionV2(SessionConfig{AccessMode: AccessModeWrite})
	req.Nil(err)
	req.NotNil(sess)

	objs := []*a{
		{
			PropsTest2: []string{"1"},
		},
		{
			PropsTest2: []string{"2"},
		},
	}

	// create 2 objects
	req.Nil(sess.ManagedTransaction(ctx, func(tx TransactionV2) error {
		err := tx.SaveDepth(ctx, objs[0], 0)
		if err != nil {
			return err
		}

		return tx.SaveDepth(ctx, objs[1], 0)
	}))

	var writeTo []*a
	req.Nil(sess.LoadAllDepth(ctx, &writeTo, 0))

	req.Nil(sess.Delete(ctx, writeTo[0]))
	req.Nil(sess.Delete(ctx, writeTo[1]))
}
