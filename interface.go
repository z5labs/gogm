package gogm

type iGogm interface {
	GetLabels() []string
}

type IVertex interface {
	iGogm
}

type IEdge interface {
	iGogm
	GetStartNode() IVertex
	SetStartNode(v IVertex) error

	GetEndNode() IVertex
	SetEndNode(v IVertex) error
}
