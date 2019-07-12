package gogm


type IVertex interface {
	GetLabels() []string
}

type IEdge interface {
	GetLabels() []string
	GetStartNode() IVertex
	SetStartNode(v IVertex) error

	GetEndNode() IVertex
	SetEndNode(v IVertex) error
}
