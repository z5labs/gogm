// Copyright (c) 2022 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gogm

import (
	"testing"

	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/stretchr/testify/require"
)

func TestSchemaLoadStrategyMany(t *testing.T) {
	req := require.New(t)

	// reusing structs from decode_test
	gogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(gogm)

	// test base case with no schema expansion
	cypher, err := SchemaLoadStrategyMany(gogm, "n", "a", 0, nil)
	req.Nil(err)
	cypherStr, err := cypher.ToCypher()
	req.Nil(err)
	req.Equal(cypherStr, "MATCH (n:a) RETURN n")

	// test base case with no schema expansion
	cypher, err = SchemaLoadStrategyMany(gogm, "n", "a", 0, dsl.C(&dsl.ConditionConfig{
		Name:              "n",
		ConditionOperator: dsl.EqualToOperator,
		Field:             "test_field",
		Check:             dsl.ParamString("$someParam"),
	}))
	req.Nil(err)
	cypherStr, err = cypher.ToCypher()
	req.Nil(err)
	req.Equal(cypherStr, "MATCH (n:a) WHERE n.test_field = $someParam RETURN n")

	// test more complex case with schema expansion
	cypher, err = SchemaLoadStrategyMany(gogm, "n", "a", 2, nil)
	req.Nil(err)
	req.NotNil(cypher)
	cypherStr, err = cypher.ToCypher()
	req.Nil(err)
	req.NotContains(cypherStr, ":c)", "Spec edge should not be treated as a node")
	req.Regexp("\\[[^\\(\\)\\[\\]]+:special[^\\(\\)\\[\\]]+]..\\([^\\(\\)\\[\\]]+:b\\)", cypherStr, "Spec edge rels should properly link to b")

	// test fail condition of non-existing label
	cypher, err = SchemaLoadStrategyMany(gogm, "n", "nonexisting", 2, nil)
	req.NotNil(err, "Should fail due to non-existing label")
	req.Nil(cypher)
}

func TestSchemaLoadStrategyOne(t *testing.T) {
	req := require.New(t)

	// reusing structs from decode_test
	gogm, err := getTestGogmWithDefaultStructs()
	req.Nil(err)
	req.NotNil(gogm)

	// test base case with no schema expansion
	cypher, err := SchemaLoadStrategyOne(gogm, "n", "a", "uuid", "uuid", false, 0, nil)
	req.Nil(err)
	cypherStr, err := cypher.ToCypher()
	req.Nil(err)
	req.Equal(cypherStr, "MATCH (n:a) WHERE n.uuid = $uuid RETURN n")
}
