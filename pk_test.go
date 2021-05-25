package gogm

import (
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestPrimaryKeyStrategy_validate(t *testing.T) {
	cases := []struct {
		name string
		strategy PrimaryKeyStrategy
		shouldPass bool
	}{
		{
			name:     "Zero Test",
			strategy: PrimaryKeyStrategy{},
			shouldPass: false,
		},
		{
			name: "Valid",
			strategy: PrimaryKeyStrategy{
				StrategyName: "ValidIndex",
				DBName:       "id",
				Type:         reflect.TypeOf(""),
				GenIDFunc: func() interface{} {
					return ""
				},
			},
			shouldPass: true,
		},
		{
			name:     "Invalid",
			strategy: PrimaryKeyStrategy{
				StrategyName: "invalid",
				DBName:       "id",
				Type:         reflect.TypeOf(""),
				GenIDFunc: func() interface{} {
					return 1
				},
			},
			shouldPass: false,
		},
	}

	req := require.New(t)
	for _, c := range cases {
		t.Log("testing case [" + c.name + "]")
		err := c.strategy.validate()
		if c.shouldPass {
			req.Nil(err)
		} else {
			req.NotNil(err)
		}
	}
}