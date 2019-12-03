package gogm

const loadMapField = "LoadMap"

type BaseNode struct {
	Id   int64  `json:"-" gogm:"name=id"`
	UUID string `json:"uuid" gogm:"pk;name=uuid"`
	// field -- relations
	LoadMap map[string]*RelationConfig `json:"-" gogm:"-"`
}

type RelationType int

const (
	Single RelationType = 0
	Multi  RelationType = 1
)

type RelationConfig struct {
	Ids          []int64      `json:"-" gomg:"-"`
	//used to replace for new nodes
	UUIDs []string `json:"-"`
	RelationType RelationType `json:"-"  gomg:"-"`
}
