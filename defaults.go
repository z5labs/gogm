package gogm

type BaseNode struct {
	Id       int64   `json:"-" gogm:"name=id"`
	UUID     string  `json:"uuid" gogm:"pk;name=uuid"`
}