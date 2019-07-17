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

func decode(arr [][]interface{}, respObj interface{}) (err error){
	defer func() {
		if r := recover(); r != nil{
			err = fmt.Errorf("%v", r)
		}
	}()

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
		return fmt.Errorf("invalid resp type %T", respObj)
	}

	if len(arr) != 3{
		return  fmt.Errorf("malformed response, invalid number of rows (%v != 3)", len(arr[0]))
	}

	p0 := len(arr[0])
	p1 := len(arr[1])
	p2 := len(arr[2])

	//setup vals
	nodeLookup := make(map[int64]*reflect.Value, p1+ p2)
	pks := make([]int64, p2, p2)
	rels := make([]neoEdgeConfig, p0, p0)

	//validate the type provided is compatible with return
	if p2 == 0{
		return errors.New("no primary node to return")
	}

	nodes := append(arr[1], arr[2]...)

	var wg sync.WaitGroup

	wg.Add(3)

	errChan := make(chan error, 3)

	go convertAndMapNodes(nodes, &nodeLookup, errChan, &wg)
	go getPks(arr[2], pks, errChan, &wg)
	go convertAndMapEdges(arr[0], rels, errChan, &wg)

	//wait for mapping to commence
	wg.Wait()

	select {
	case err := <- errChan:
		log.WithError(err).Error()
		return err
	default:
		log.Debugf("passed setup")
	}

	close(errChan)

	//sanity check
	if len(nodeLookup) != p1 + p2{
		return fmt.Errorf("sanity check failed, nodeLookup not correct length (%v) != (%v)", len(nodeLookup), p1 + p2)
	}

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

		key := makeRelMapKey(relationConfig.StartNodeType, relationConfig.Type)

		var internalEdgeConf decoratorConfig
		temp, ok := mappedRelations.Get(key)
		if !ok {
			return fmt.Errorf("cannot find decorator config for key %s", key)
		}

		internalEdgeConf, ok = temp.(decoratorConfig)
		if !ok{
			return errors.New("unable to cast into decoratorConfig")
		}

		if internalEdgeConf.UsesEdgeNode {
			var typeConfig structDecoratorConfig

			it := internalEdgeConf.Type

			//get the actual type if its a slice
			if it.Kind() == reflect.Slice{
				it = it.Elem()
			}

			temp, ok := mappedTypes.Get(internalEdgeConf.Type.String())// mappedTypes[boltNode.Labels[0]]
			if !ok{
				return fmt.Errorf("can not find mapping for node with label %s", internalEdgeConf.Type.String())
			}
			
			typeConfig = temp.(structDecoratorConfig)
			if !ok{
				return errors.New("unable to cast to structDecoratorConfig")
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
				//non slice relationships are already asserted to be pointers
				reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Set(val.Addr())
			}

			//relate start-start
			if reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(*start, *val))
			} else {
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(val.Addr())
			}
		} else {
			if end.FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				end.FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(*end, *start))
			} else {
				end.FieldByName(internalEdgeConf.FieldName).Set(start.Addr())
			}

			//relate end-start
			if start.FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				start.FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(*start, *end))
			} else {
				start.FieldByName(internalEdgeConf.FieldName).Set(end.Addr())
			}
		}
	}

	//handle if its returning a slice -- validation has been done at an earlier step
	if rt.Elem().Kind() == reflect.Slice{

		reflection := reflect.MakeSlice(rt.Elem(), 0, 0)

		reflectionValue := reflect.New(reflection.Type())
		reflectionValue.Elem().Set(reflection)

		slicePtr := reflect.ValueOf(reflectionValue.Interface())

		sliceValuePtr := slicePtr.Elem()

		for _, id := range pks{
			val, ok := nodeLookup[id]
			if !ok{
				return fmt.Errorf("cannot find value with id (%v)", id)
			}

			log.Info(val.Interface())
			sliceValuePtr.Set(reflect.Append(sliceValuePtr, *val))
		}

		rv.Set(sliceValuePtr)

		return err
	} else {
		//handles single -- already checked to make sure p2 is at least 1
		reflect.Indirect(rv).Set(*nodeLookup[pks[0]])

		return err
	}
}

func getValueAndConfig(id int64, t string, nodeLookup map[int64]*reflect.Value) (val *reflect.Value, conf structDecoratorConfig, err error){
	var ok bool

	val, ok = nodeLookup[id]
	if !ok {
		return nil, structDecoratorConfig{}, fmt.Errorf("value for id (%v) not found", id)
	}

	temp, ok := mappedTypes.Get(t)
	if !ok {
		return nil, structDecoratorConfig{}, fmt.Errorf("no config found for type (%s)", t)
	}

	conf, ok = temp.(structDecoratorConfig)
	if !ok{
		return nil, structDecoratorConfig{}, errors.New("unable to cast to structDecoratorConfig")
	}

	return
}

func getPks(nodes []interface{}, pks []int64, err chan error, wg *sync.WaitGroup) {
	if nodes == nil || len(nodes) == 0{
		err <- fmt.Errorf("nodes can not be nil or empty")
	}

	for i, node := range nodes{
		nodeConv, ok := node.(graph.Node)
		if !ok{
			err <- fmt.Errorf("unable to cast node to type graph.Node")
			wg.Done()
			return
		}

		pks[i] = nodeConv.NodeIdentity
	}

	wg.Done()
}

func convertAndMapEdges(nodes []interface{}, rels []neoEdgeConfig, err chan error, wg *sync.WaitGroup){
	if nodes == nil{
		err <- errors.New("edges can not be nil or empty")
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
			err <- fmt.Errorf("unknown type %s", reflect.TypeOf(n).String())
		}
	}

	wg.Done()
}

func convertAndMapNodes(nodes []interface{}, lookup *map[int64]*reflect.Value, err chan error, wg *sync.WaitGroup) {
	if nodes == nil || len(nodes) == 0{
		err <- errors.New("nodes can not be nil or empty")
		wg.Done()
		return
	}

	if lookup == nil{
		err <- errors.New("lookup can not be nil")
		wg.Done()
		return
	}

	for _, node := range nodes{
		boltNode, ok := node.(graph.Node)
		if !ok{
			err <- fmt.Errorf("unable to convert bolt node to graph.Node, it is type %T", node)
			wg.Done()
			return
		}

		var val *reflect.Value
		var e error
		val, e = convertNodeToValue(boltNode)
		if e != nil{
			err <- e
			wg.Done()
			return
		}

		(*lookup)[boltNode.NodeIdentity] = val
	}

	wg.Done()
}

func convertToValue(graphId int64, conf structDecoratorConfig, props map[string]interface{}, rtype reflect.Type) (valss *reflect.Value, err error){
	defer func() {
		if r := recover(); r != nil{
			err = fmt.Errorf("%v", r)
		}
	}()

	if rtype == nil{
		return nil, errors.New("rtype can not be nil")
	}

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

		//skip if its a relation field
		if fieldConfig.Relationship != ""{
			continue
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

	temp, ok := mappedTypes.Get(boltNode.Labels[0])// mappedTypes[boltNode.Labels[0]]
	if !ok{
		return nil, fmt.Errorf("can not find mapping for node with label %s", boltNode.Labels[0])
	}

	typeConfig, ok = temp.(structDecoratorConfig)
	if !ok{
		return nil, errors.New("unable to cast to struct decorator config")
	}

	return convertToValue(boltNode.NodeIdentity, typeConfig, boltNode.Properties, typeConfig.Type)
}


