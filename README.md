[![Go Report Card](https://goreportcard.com/badge/github.com/mindstand/gogm)](https://goreportcard.com/report/github.com/mindstand/gogm)
[![Actions Status](https://github.com/mindstand/gogm/workflows/Go/badge.svg)](https://github.com/mindstand/gogm/actions)
[![](https://godoc.org/github.com/nathany/looper?status.svg)](http://godoc.org/github.com/mindstand/gogm)
# GoGM Golang Object Graph Mapper

```
go get -u github.com/mindstand/gogm
```

#### Documentation updates will be coming periodically as this project matures

## Features
- Struct Mapping through the `gogm` struct decorator
- Full support for ACID transactions
- Underlying connection pooling
- Support for HA Clusters using `bolt+routing` through [MindStand's fork](https://github.com/mindstand/golang-neo4j-bolt-driver) of [@johnnadratowski's golang bolt driver](https://github.com/johnnadratowski/golang-neo4j-bolt-driver)
- Custom queries in addition to built in functionality
- Builder pattern cypher queries using [MindStand's cypher dsl package](https://github.com/mindstand/go-cypherdsl)
- CLI to generate link and unlink functions for gogm structs.

## Usage

### Struct Configuration

Decorators that can be used
- `name=<name>` -- used to set the field name that will show up in neo4j.
- `relationship=<edge_name>` -- used to set the name of the edge on that field.
- `direction=<INCOMING|OUTGOING|BOTH|NONE>` -- used to specify direction of that edge field.
- `time` -- used to mark field as a time representation. compatible with `time.Time` Customizable time format coming soon.
- `index` -- marks field to have an index applied to it.
- `unique` -- marks field to have unique constraint.
- `pk` -- marks field as a primary key. Can only have one pk, composite pk's are not supported.
- `properties` -- marks that field is using a map. GoGM only supports properties fields of `map[string]interface{}`
- `-` -- marks that field will be ignored by the ogm

#### Not on relationship member variables
All relationships must be defined as either a pointer to a struct or a slice of struct pointers `*SomeStruct` or `[]*SomeStruct`

Use `;` as delimiter between decorator tags.

Ex.

```go
type TdString string

type MyNeo4jObject struct {
  // provides required node fields
  gogm.BaseNode

  Field string `gogm:"name=field"`
  Props map[string]interface{} `gogm:"properties;name=props"` //note that this would show up as `props.<key>` in neo4j
  IgnoreMe bool `gogm="-"`
  UniqueTypeDef TdString `gogm:"name=unique_type_def"`
  Relation *SomeOtherStruct `gogm="relationship=SOME_STRUCT;direction=OUTGOING"`
  ManyRelation []*SomeStruct `gogm="relationship=MANY;direction=INCOMING"`
}

```

### GOGM Usage
```go
package main

import (
  "github.com/mindstand/gogm"
  "time"
)

type tdString string
type tdInt int

//structs for the example (can also be found in decoder_test.go)
type VertexA struct {
    // provides required node fields
	gogm.BaseNode

	TestField         string   `gogm:"name=test_field"`
	TestTypeDefString tdString `gogm:"name=test_type_def_string"`
	TestTypeDefInt    tdInt    `gogm:"name=test_type_def_int"`
	SingleA           *VertexB       `gogm:"direction=incoming;relationship=test_rel"`
	ManyA             []*VertexB     `gogm:"direction=incoming;relationship=testm2o"`
	MultiA            []*VertexB     `gogm:"direction=incoming;relationship=multib"`
	SingleSpecA       *EdgeC       `gogm:"direction=outgoing;relationship=special_single"`
	MultiSpecA        []*EdgeC     `gogm:"direction=outgoing;relationship=special_multi"`
}

type VertexB struct {
    // provides required node fields
	gogm.BaseNode

	TestField  string    `gogm:"name=test_field"`
	TestTime   time.Time `gogm:"time;name=test_time"`

	Single     *VertexA         `gogm:"direction=outgoing;relationship=test_rel"`
	ManyB      *VertexA         `gogm:"direction=incoming;relationship=testm2o"`
	Multi      []*VertexA       `gogm:"direction=outgoing;relationship=multib"`

	SingleSpec *EdgeC        `gogm:"direction=incoming;relationship=special_single"`
	MultiSpec  []*EdgeC      `gogm:"direction=incoming;relationship=special_multi"`
}

type EdgeC struct {
    // provides required node fields
	gogm.BaseNode

	Start *VertexA
	End   *VertexB
	Test  string `gogm:"name=test"`
}

func main(){
  config := gogm.Config{
    IndexStrategy: gogm.VALIDATE_INDEX, //other options are ASSERT_INDEX and IGNORE_INDEX
    PoolSize:      50,
    Port:          7687,
    IsCluster:     false, //tells it whether or not to use `bolt+routing`
    Host:          "0.0.0.0",
    Password:      "password",
    Username:      "neo4j",
  }
  
  err := gogm.Init(&config, &VertexA{}, &VertexB{}, &EdgeC{})
  if err != nil {
    panic(err)
  }
  
  //param is readonly, we're going to make stuff so we're going to do read write
  sess, err := gogm.NewSession(false) 
  if err != nil {
    panic(err)
  }
  
  //close the session
  defer sess.Close()
  
  aVal := &VertexA{
    TestField: "woo neo4j",
  }
  
  bVal := &VertexB{
    TestTime: time.Now().UTC(),
  }
  
  //set bi directional pointer
  bVal.Single = aVal
  aVal.SingleA = bVal
  
  err = sess.SaveDepth(&aVal, 2)
  if err != nil {
    panic(err)
  }
  
  //load the object we just made (save will set the uuid)
  var readin VertexA
  err = sess.Load(&readin, aVal.UUID)
  if err != nil {
    panic(err)
  }
  
  
}

```

### GoGM CLI

## CLI Installation
```
go get -u github.com/mindstand/gogm/cli/gogmcli
```

## CLI Usage
```
NAME:
   gogmcli - used for neo4j operations from gogm schema

USAGE:
   gogmcli [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   generate, g, gen  to generate link and unlink functions for nodes
   help, h           Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --debug, -d    execute in debug mode (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

## Inspiration
Inspiration came from the Java OGM implementation by Neo4j.

## Road Map
- Schema Migration
- Errors overhaul using go 1.13 error wrapping
- TLS Support
- Documentation (obviously)

## How you can help
- Report Bugs
- Fix bugs
- Contribute (refer to contribute.md)
