package gogm

import "reflect"

type Vertex struct{
	Id string `json:"-" gogm:"name=id"`
	UUID string `json:"uuid" gogm:"pk;name=uuid"`
}

type Edge struct{
	Id string `json:"-" gogm:"name=id"`
}

type EdgeConfig struct {
	Type string
	Def *reflect.Value
	StartNode int64
	EndNode int64
}
