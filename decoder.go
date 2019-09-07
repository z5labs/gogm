package gogm

import (
	"errors"
	"fmt"
	neo "github.com/mindstand/golang-neo4j-bolt-driver"
	"github.com/mindstand/golang-neo4j-bolt-driver/structures/graph"
	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
	"sync"
	"time"
)

func decodeNeoRows(rows neo.Rows, respObj interface{}) error{
	defer rows.Close()

	arr, err := dsl.RowsTo2DInterfaceArray(rows)
	if err != nil{
		return err
	}

	return decode(arr, respObj)
}

func decode(rawArr [][]interface{}, respObj interface{}) (err error){
	defer func() {
		if r := recover(); r != nil{
			err = fmt.Errorf("%v", r)
		}
	}()

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

	if rawArr == nil || len(rawArr) != 1{
		return fmt.Errorf("invalid rawArr size, %v", len(rawArr))
	}

	arr1 := rawArr[0]

	if len(arr1) != 3{
		return  fmt.Errorf("malformed response, invalid number of rows (%v != 3)", len(arr1))
	}

	var arr [][]interface{}

	arr = append(arr, arr1[0].([]interface{}))
	arr = append(arr, arr1[1].([]interface{}))
	arr = append(arr, arr1[2].([]interface{}))

	//check for empty stuff
	for i := 0; i < 3; i++ {
		if len(arr[i]) == 0 {
			continue
		}
		if aCheck, ok := arr[i][0].([]interface{}); ok {
			if len(aCheck) == 0 {
				//set it to just be empty
				arr[i] = []interface{}{}
			}
		}
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

	//build relationships
	for i, relationConfig := range rels{
		if i == 0 {
			continue
		}
		start, _, err := getValueAndConfig(relationConfig.StartNodeId, relationConfig.StartNodeType, nodeLookup)
		if err != nil {
			return err
		}

		end, _, err := getValueAndConfig(relationConfig.EndNodeId, relationConfig.EndNodeType, nodeLookup)
		if err != nil {
			return err
		}

		configs, err := mappedRelations.GetConfigs(relationConfig.StartNodeType, relationConfig.Type, relationConfig.EndNodeType)
		if err != nil {
			return err
		}

		var internalEdgeConf decoratorConfig

		if len(configs) == 0 {
			return errors.New("no config found")
		} else if len(configs) == 1 {
			internalEdgeConf = configs[0]
		} else {
			for _, config := range configs{
				if config.Direction == dsl.Outgoing {
					internalEdgeConf = config
					break
				}
				return errors.New("cannot find outgoing edge")
			}
		}

		if internalEdgeConf.UsesEdgeNode {
			var typeConfig structDecoratorConfig

			it := internalEdgeConf.Type

			//get the actual type if its a slice
			if it.Kind() == reflect.Slice{
				it = it.Elem()
			}

			label := ""

			if internalEdgeConf.Type.Kind() == reflect.Ptr{
				label = it.Elem().Name()
			} else {
				label = it.Name()
				it = reflect.PtrTo(it)
			}

			temp, ok := mappedTypes.Get(label)// mappedTypes[boltNode.Labels[0]]
			if !ok{
				return fmt.Errorf("can not find mapping for node with label %s", label)
			}
			
			typeConfig = temp.(structDecoratorConfig)
			if !ok{
				return errors.New("unable to cast to structDecoratorConfig")
			}

			//create value
			val, err := convertToValue(-1, typeConfig, relationConfig.Obj, it)
			if err != nil{
				return err
			}

			var startCall reflect.Value
			var endCall reflect.Value

			if start.Kind() != reflect.Ptr {
				startCall = start.Addr()
			} else {
				startCall = *start
			}

			if end.Kind() != reflect.Ptr{
				endCall = end.Addr()
			} else {
				endCall = *end
			}

			//can ensure that it implements proper interface if it made it this far
			res := val.MethodByName("SetStartNode").Call([]reflect.Value{startCall})
			if res == nil || len(res) == 0 {
				return errors.New("invalid response")
			} else if !res[0].IsNil(){
				return res[0].Interface().(error)
			}

			res = val.MethodByName("SetEndNode").Call([]reflect.Value{endCall})
			if res == nil || len(res) == 0 {
				return errors.New("invalid response")
			} else if !res[0].IsNil(){
				return res[0].Interface().(error)
			}

			//relate end-start
			if reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName), reflect.Indirect(*val)))
			} else {
				//non slice relationships are already asserted to be pointers
				end.FieldByName(internalEdgeConf.FieldName).Set(*val)
			}

			//relate start-start
			if reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName), reflect.Indirect(*val)))
			} else {
				start.FieldByName(internalEdgeConf.FieldName).Set(*val)
			}
		} else {
			if end.FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(reflect.Indirect(*end).FieldByName(internalEdgeConf.FieldName), *start))
			} else {
				end.FieldByName(internalEdgeConf.FieldName).Set(start.Addr())
			}

			//relate end-start
			if start.FieldByName(internalEdgeConf.FieldName).Kind() == reflect.Slice{
				reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName).Set(reflect.Append(reflect.Indirect(*start).FieldByName(internalEdgeConf.FieldName), *end))
			} else {
				start.FieldByName(internalEdgeConf.FieldName).Set(end.Addr())
			}
		}
	}

	//handle if its returning a slice -- validation has been done at an earlier step
	if rt.Elem().Kind() == reflect.Slice{

		reflection := reflect.MakeSlice(rt.Elem(), 0, cap(pks))

		reflectionValue := reflect.New(reflection.Type())
		reflectionValue.Elem().Set(reflection)

		slicePtr := reflect.ValueOf(reflectionValue.Interface())

		sliceValuePtr := slicePtr.Elem()

		for _, id := range pks{
			val, ok := nodeLookup[id]
			if !ok{
				return fmt.Errorf("cannot find value with id (%v)", id)
			}

			sliceValuePtr.Set(reflect.Append(sliceValuePtr, *val))
		}

		reflect.Indirect(rv).Set(sliceValuePtr)

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
		wg.Done()
		return
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
		//the way this is returned, skip index 0
		if i == 0{
			continue
		}

		//this is because of how resp is structured
		narr, ok := n.([]interface{})
		if !ok{
			err <- fmt.Errorf("unable to cast to []interface, type is %T", n)
			wg.Done()
			return
		}

		if len(narr) == 0{
			err <- errors.New("length should not be nil")
			wg.Done()
			return
		}

		for _, nr := range narr{
			var node neoEdgeConfig
			err1 := mapstructure.Decode(nr.(map[string]interface{}), &node)
			if err1 != nil{
				err <- err1
				wg.Done()
				return
			} else {
				rels[i] = node
			}
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

var sliceOfEmptyInterface []interface{}
var emptyInterfaceType = reflect.TypeOf(sliceOfEmptyInterface).Elem()

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

		if fieldConfig.Properties {
			mapType := reflect.MapOf(reflect.TypeOf(""), emptyInterfaceType)
			mapVal := reflect.MakeMap(mapType)

			for k, v := range props {
				if !strings.Contains(k, fieldConfig.Name) {
					//not one of our map fields
					continue
				}

				parts := strings.Split(k, ".")
				if len(parts) != 2 {
					return nil, fmt.Errorf("invalid key [%s]", k)
				}

				mapKey := parts[1]

				mapVal.SetMapIndex(reflect.ValueOf(mapKey), reflect.ValueOf(v))
			}

			reflect.Indirect(val).FieldByName(field).Set(mapVal)
			continue
		}

		raw, ok := props[fieldConfig.Name]
		if !ok{
			return nil, fmt.Errorf("unrecognized field [%s]", fieldConfig.Name)
		}

		if raw == nil{
			continue //its already initialized to 0 value, no need to do anything
		} else {
			if fieldConfig.IsTime {
				timeStr, ok := raw.(string)
				if !ok {
					return nil, errors.New("can not convert interface{} time to string")
				}

				convTime, err := time.Parse(time.RFC3339, timeStr)
				if err != nil{
					return nil, err
				}

				var writeVal reflect.Value

				if fieldConfig.Type.Kind() == reflect.Ptr{
					writeVal = reflect.ValueOf(convTime).Addr()
				} else {
					writeVal = reflect.ValueOf(convTime)
				}

				reflect.Indirect(val).FieldByName(field).Set(writeVal)
			} else {
				reflect.Indirect(val).FieldByName(field).Set(reflect.ValueOf(raw))
			}
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


