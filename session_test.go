package gogm

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSession(t *testing.T) {
	if !testing.Short() {
		t.Skip()
		return
	}

	req := require.New(t)

	conf := Config{
		Username:      "neo4j",
		Password:      "password",
		Host:          "0.0.0.0",
		Port:          7687,
		PoolSize:      15,
		IndexStrategy: VALIDATE_INDEX,
	}

	req.Nil(Init(&conf, &a{}, &b{}, &c{}))

	req.EqualValues(3, mappedTypes.Len())

	sess, err := NewSession(true)
	req.NotNil(err)

	var stuffs []a

	req.Nil(sess.LoadAll(&stuffs))

	req.EqualValues(1, len(stuffs))
}
