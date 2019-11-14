package gogm

const loadMapField = "LoadMap"

type BaseNode struct {
	Id       int64   `json:"-" gogm:"name=id"`
	UUID     string  `json:"uuid" gogm:"pk;name=uuid"`
	LoadMap map[string]*RelationLoad `json:"-" gogm:"-"`
}

type RelationType int
const (
	Single RelationType = 0
	Multi RelationType = 1
)

type RelationLoad struct {
	Ids []int64 `json:"-"`
	RelationType RelationType `json:"-"`
}