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

	return decode(arr, respObj)
}

func decode(arr [][]interface{}, respObj interface{}) error{
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

	if respObj == nil {
		return errors.New("response object can not be nil")
	}

	rv := reflect.ValueOf(respObj)
	rt := reflect.TypeOf(respObj)

	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("invalid resp type")
	}

	if len(arr) != 3{
		return  fmt.Errorf("malformed response, invalid number of rows (%v != 3)", len(arr[0]))
	}

	p0 := len(arr[0])
	p1 := len(arr[1])
	p2 := len(arr[2])

	//setup vals
	nodeLookup := make(map[int64]*reflect.Value, p1+ p2)
	pks := make([]int64, 0, p2)
	rels := make([]neoEdgeConfig, 0, p0)

	//validate the type provided is compatible with return
	if p2 == 0{
		return errors.New("no primary node to return")
	}

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

	//build relationships
	for _, relationConfig := range rels{
		start, _, err := getValueAndConfig(relationConfig.StartNodeId, relationConfig.StartNodeType, nodeLookup)
		if err != nil {
			return err
		}

		end, _, err := getValueAndConfig(relationConfig.EndNodeId, relationConfig.EndNodeType, nodeLookup)
		if err != nil {
			return err
		}

		var internalEdgeConf decoratorConfig
		_, ok := mappedRelations.GetOrInsert(makeRelMapKey(relationConfig.StartNodeType, relationConfig.Type), &internalEdgeConf)
		if !ok {
			return errors.New("cannot find decorator config for key")
		}

		if internalEdgeConf.UsesEdgeNode {
			var typeConfig structDecoratorConfig

			it := internalEdgeConf.Type

			//get the actual type if its a slice
			if it.Kind() == reflect.Slice{
				it = it.Elem()
			}

			_, ok := mappedTypes.GetOrInsert(internalEdgeConf.Type.Name(), &typeConfig)// mappedTypes[boltNode.Labels[0]]
			if !ok{
				return fmt.Errorf("can not find mapping for node with label %s", internalEdgeConf.Type.Name())
			}

			//create value
			val, err := convertToValue(-1, typeConfig, relationConfig.Obj, internalEdgeConf.Type)
			if err != nil{
				return err
			}

			//can ensure that it implements proper interface if it made it this far
			res := val.MethodByName("SetStartNode").Call([]reflect.Value{*start})
			if res == nil || len(res) == 0 {
				return errors.New("invalid response")
			} else if !res[0].IsNil(){
				return res[0].Interface().(error)
			}

			res = val.MethodByName("SetEndNode").Call([]reflect.Value{*start})
			if res == nil || len(res) == 0 {
				return errors.New("invalid response")
			} else if !res[0].IsNil(){
				return res[0].Interface().(error)
			}

			//relate end-start
			if reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(*start, *val))
			} else {
				reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Set(*val)
			}

			//relate start-start
			if reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(*start, *val))
			} else {
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(*val)
			}
		} else {
			if reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(*end, *start))
			} else {
				reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Set(*start)
			}

			//relate end-start
			if reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(*start, *end))
			} else {
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(*end)
			}
		}
	}

	//handle if its returning a slice -- validation has been done at an earlier step
	if p2 > 1{
		//get type in slice
		sliceType := rt.Elem()

		//make a new slice
		retSlice := reflect.MakeSlice(sliceType, len(pks), cap(pks))

		for _, id := range pks{
			retSlice.Set(reflect.Append(retSlice, *nodeLookup[id]))
		}

		respObj = retSlice.Interface()
		return nil
	} else {
		//handles single -- already checked to make sure p2 is at least 1
		respObj = nodeLookup[pks[0]].Interface()
		return nil
	}
}

func getValueAndConfig(id int64, t string, nodeLookup map[int64]*reflect.Value) (val *reflect.Value, conf structDecoratorConfig, err error){
	var ok bool

	val, ok = nodeLookup[id]
	if !ok {
		return nil, structDecoratorConfig{}, fmt.Errorf("value for id (%v) not found", id)
	}

	_, ok = mappedTypes.GetOrInsert(t, &conf)
	if !ok {
		return nil, structDecoratorConfig{}, fmt.Errorf("no config found for type (%s)", t)
	}

	return
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

func convertAndMapEdges(nodes []interface{}, rels []neoEdgeConfig, err error, wg *sync.WaitGroup){
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
		if node, ok := n.(neoEdgeConfig); ok{
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

func convertToValue(graphId int64, conf structDecoratorConfig, props map[string]interface{}, rtype reflect.Type) (*reflect.Value, error){
	var err error
	defer catchPanic(err)

	isPtr := false
	if rtype.Kind() == reflect.Ptr{
		isPtr = true
		rtype = rtype.Elem()
	}

	val := reflect.New(rtype)

	if graphId >= 0{
		reflect.Indirect(val).FieldByName("Id").Set(reflect.ValueOf(graphId))
	}

	for field, fieldConfig := range conf.Fields{
		if fieldConfig.Name == "id"{
			continue //id is handled above
		}

		raw, ok := props[fieldConfig.Name]
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

func convertNodeToValue(boltNode graph.Node) (*reflect.Value, error){

	if boltNode.Labels == nil || len(boltNode.Labels) == 0{
		return nil, errors.New("boltNode has no labels")
	}

	var typeConfig structDecoratorConfig

	_, ok := mappedTypes.GetOrInsert(boltNode.Labels[0], &typeConfig)// mappedTypes[boltNode.Labels[0]]
	if !ok{
		return nil, fmt.Errorf("can not find mapping for node with label %s", boltNode.Labels[0])
	}

	return convertToValue(boltNode.NodeIdentity, typeConfig, boltNode.Properties, typeConfig.Type)
}

func catchPanic(err error){
	if r := recover(); r != nil{
		err = fmt.Errorf("%v", r)
	}
}



