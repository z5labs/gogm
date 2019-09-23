package gogm

type Vertex struct{
	Id string `json:"-" gogm:"name=id"`
	UUID string `json:"uuid" gogm:"pk;name=uuid"`
}

type Edge struct{
	Id string `json:"-" gogm:"name=id"`
}

type neoEdgeConfig struct {
	Id int64

	StartNodeId int64
	StartNodeType string

	EndNodeId int64
	EndNodeType string

	Obj  map[string]interface{}

	Type string
}
