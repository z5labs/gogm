package gogm

import (
	"errors"
	"fmt"
	neo "github.com/johnnadratowski/golang-neo4j-bolt-driver"
	"github.com/johnnadratowski/golang-neo4j-bolt-driver/structures/graph"
	dsl "github.com/mindstand/go-cypherdsl"
	"reflect"
	"sync"
)
func DecodeNeoRows(rows neo.Rows, respObj interface{}) error{
	defer rows.Close()

	arr, err := dsl.RowsTo2DInterfaceArray(rows)
	if err != nil{
		return err
	}

	return decode(arr)
}

func decode(arr [][]interface{}) error{
	/*
		MATCH (n:OrganizationNode)
		WITH n
		MATCH (n)-[e*0..1]-(m)
		RETURN DISTINCT
			collect(extract(n in e | {StartNodeId: ID(startnode(n)), StartNodeType: labels(startnode(n)), EndNodeId: ID(endnode(n)), EndNode: labels(endnode(n)), Obj: n, Type: type(n)})) as Edges,
			collect(DISTINCT m) as Ends,
			collect(DISTINCT n) as Starts
	*/

	//                                        0               1          2
	//signature of returned array should be list of edges, list of ends, list of starts
	// length of 3

	if len(arr) != 3{
		return  fmt.Errorf("malformed response, invalid number of rows (%v != 3)", len(arr[0]))
	}

	p0 := len(arr[0])
	p1 := len(arr[1])
	p2 := len(arr[2])

	//setup vals
	nodeLookup := make(map[int64]*reflect.Value, p1+ p2)
	pks := make([]int64, 0, p2)
	rels := make([]EdgeConfig, 0, p0)

	var nErr error
	var eErr error
	var pErr error

	nodes := append(arr[1], arr[2])

	var wg sync.WaitGroup

	wg.Add(3)

	go convertAndMapNodes(nodes, nodeLookup, nErr, &wg)
	go getPks(arr[2], pks, pErr, &wg)
	go convertAndMapEdges(arr[0], rels, eErr, &wg)

	//wait for mapping to commence
	wg.Wait()

	return nil
}

func getPks(nodes []interface{}, pks []int64, err error, wg *sync.WaitGroup) {
	if nodes == nil || len(nodes) == 0{
		err = fmt.Errorf("nodes can not be nil or empty")
	}

	for i, node := range nodes{
		nodeConv, ok := node.(graph.Node)
		if !ok{
			err = fmt.Errorf("unable to cast node to type graph.Node")
			wg.Done()
			return
		}

		pks[i] = nodeConv.NodeIdentity
	}

	err = nil
	wg.Done()
}

func convertAndMapEdges(nodes []interface{}, rels []EdgeConfig, err error, wg *sync.WaitGroup){
	if nodes == nil{
		err = errors.New("edges can not be nil or empty")
		wg.Done()
		return
	}

	if len(nodes) == 0{
		wg.Done()
		return
	}

	for i, n := range nodes{
		if node, ok := n.(EdgeConfig); ok{
			rels[i] = node
		} else {
			err = fmt.Errorf("unknown type %s", reflect.TypeOf(n).Name())
		}
	}

	err = nil
	wg.Done()
}

func convertAndMapNodes(nodes []interface{}, lookup map[int64]*reflect.Value, err error, wg *sync.WaitGroup) {
	if nodes == nil || len(nodes) == 0{
		err = errors.New("nodes can not be nil or empty")
		wg.Done()
		return
	}

	if lookup == nil{
		err = errors.New("lookup can not be nil")
		wg.Done()
		return
	}

	for _, node := range nodes{
		boltNode, ok := node.(graph.Node)
		if !ok{
			err = errors.New("unable to convert bolt node to graph.Node")
			wg.Done()
			return
		}

		var val *reflect.Value

		val, err = convertNodeToValue(boltNode)
		if err != nil{
			wg.Done()
			return
		}

		lookup[boltNode.NodeIdentity] = val
	}

	err = nil
	wg.Done()
}

func convertNodeToValue(boltNode graph.Node) (*reflect.Value, error){
	var err error
	defer catchPanic(err)

	if boltNode.Labels == nil || len(boltNode.Labels) == 0{
		return nil, errors.New("boltNode has no labels")
	}

	typeConfig, ok := mappedTypes[boltNode.Labels[0]]
	if !ok{
		return nil, fmt.Errorf("can not find mapping for node with label %s", boltNode.Labels[0])
	}

	t := typeConfig.Type

	isPtr := false
	if typeConfig.Type.Kind() == reflect.Ptr{
		isPtr = true
		t= typeConfig.Type.Elem()
	}

	val := reflect.New(t)

	reflect.Indirect(val).FieldByName("Id").Set(reflect.ValueOf(boltNode.NodeIdentity))

	for field, fieldConfig := range typeConfig.Fields{
		if fieldConfig.Name == "id"{
			continue //id is handled above
		}

		raw, ok := boltNode.Properties[fieldConfig.Name]
		if !ok{
			return nil, fmt.Errorf("unrecognized field [%s]", fieldConfig.Name)
		}

		if raw == nil{
			continue //its already initialized to 0 value, no need to do anything
		} else {
			reflect.Indirect(val).FieldByName(field).Set(reflect.ValueOf(raw))
		}
	}

	//if its not a pointer, dereference it
	if !isPtr{
		retV := reflect.Indirect(val)
		return &retV, nil
	}

	return &val, err
}

func catchPanic(err error){
	if r := recover(); r != nil{
		err = fmt.Errorf("%v", r)
	}
}



