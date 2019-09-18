package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
	"reflect"
)

const maxSaveDepth = 10
const defaultSaveDepth = 1

type nodeCreateConf struct {
	Params map[string]interface{}
	Type reflect.Type
	IsNew bool
}

type relCreateConf struct {
	StartNodeUUID string
	EndNodeUUID string
	Params map[string]interface{}
	Direction dsl.Direction
}

func saveDepth(sess *driver.BoltConn, obj interface{}, depth int) error {
	if sess == nil {
		return errors.New("session can not be nil")
	}

	if obj == nil{
		return errors.New("obj can not be nil")
	}

	if depth < 0 {
		return errors.New("cannot save a depth less than 0")
	}

	if depth > maxSaveDepth {
		return fmt.Errorf("saving depth of (%v) is currently not supported, maximum depth is (%v)", depth, maxSaveDepth)
	}

	//validate that obj is a pointer
	rawType := reflect.TypeOf(obj)

	if rawType.Kind() != reflect.Ptr{
		return fmt.Errorf("obj must be of type pointer, not %T", obj)
	}

	//validate that the dereference type is a struct
	derefType := rawType.Elem()

	if derefType.Kind() != reflect.Struct{
		return fmt.Errorf("dereference type can not be of type %T", obj)
	}

	//signature is [LABEL][UUID]{config}
	nodes := map[string]map[string]nodeCreateConf{}

	//signature is [LABEL] []{config}
	relations := map[string][]relCreateConf{}

	rootVal := reflect.ValueOf(obj)

	err := parseStruct("", "", false, 0, nil, &rootVal, 0, depth, &nodes, &relations)
	if err != nil{
		return err
	}

	ids, err := createNodes(sess, nodes)
	if err != nil{
		return err
	}

	//no relations to make
	if len(ids) == 1 {
		return nil
	}

	return relateNodes(sess, relations, ids)
}

func createNodes(conn *driver.BoltConn, crNodes map[string]map[string]nodeCreateConf) (map[string]int64, error){
	idMap := map[string]int64{}

	for label, nodes := range crNodes{
		var rows []interface{}
		for _, config := range nodes{
			rows = append(rows, config.Params)
		}

		params, err := dsl.ParamsFromMap(
			map[string]interface{}{
				"uuid": dsl.ParamString("row.uuid"),
			})
		if err != nil{
			return nil, err
		}

		path, err := dsl.Path().V(dsl.V{
			Name: "n",
			Type: "`" + label + "`",
			Params: params,
		}).ToCypher()
		if err != nil{
			return nil, err
		}

		//todo replace once unwind is fixed and path
		res, err := dsl.QB().
			Cypher("UNWIND {rows} as row").
			Merge(&dsl.MergeConfig{
				Path: path,
			}).
			Cypher("SET n += row").
			Return(false, dsl.ReturnPart{
				Name: "row.uuid",
				Alias: "uuid",
			}, dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name: "ID",
					Params: []interface{}{dsl.ParamString("n")},
				},
				Alias: "id",
			}).
			WithNeo(conn).
			Query(map[string]interface{}{
				"rows": rows,
			})
		if err != nil{
			return nil, err
		}

		if res == nil{
			return nil, errors.New("res should not be nil")
		}

		resRows, _, err := res.All()
		if err != nil{
			return nil, err
		}

		for _, row := range resRows{
			if len(row) != 2{
				continue
			}

			idMap[row[0].(string)] = row[1].(int64)
		}

		err = res.Close()
		if err != nil {
			return nil, err
		}
	}

	return idMap, nil
}

func relateNodes(conn *driver.BoltConn, relations map[string][]relCreateConf, ids map[string]int64) error{
	if relations == nil || len(relations) == 0{
		return errors.New("relations can not be nil or empty")
	}

	if ids == nil || len(ids) == 0{
		return errors.New("ids can not be nil or empty")
	}

	for label, rels := range relations{
		var params []interface{}

		if len(rels) == 0{
			continue
		}

		for _, rel := range rels{
			startId, ok := ids[rel.StartNodeUUID]
			if !ok{
				return fmt.Errorf("can not find nodeId for uuid %s", rel.StartNodeUUID)
			}

			endId, ok := ids[rel.EndNodeUUID]
			if !ok {
				return fmt.Errorf("can not find nodeId for %T", rel)
			}

			//set map if its empty
			if rel.Params == nil{
				rel.Params = map[string]interface{}{}
			}

			params = append(params, map[string]interface{}{
				"startNodeId": startId,
				"endNodeId": endId,
				"props": rel.Params,
			})
		}

		mergePath, err := dsl.Path().
			V(dsl.V{
				Name: "startNode",
			}).
			E(dsl.E{
				Name: "rel",
				Types: []string{
					label,
				},
				Direction: dsl.DirectionOutgoing,
			}).
			V(dsl.V{
				Name: "endNode",
			}).
			ToCypher()
		if err != nil{
			return err
		}

		_, err = dsl.QB().
			Cypher("UNWIND {rows} as row").
			Match(dsl.Path().V(dsl.V{Name: "startNode"}).Build()).
			Where(dsl.C(&dsl.ConditionConfig{
				FieldManipulationFunction: "ID",
				Name: "startNode",
				ConditionOperator: dsl.EqualToOperator,
				Check: dsl.ParamString("row.startNodeId"),
			})).
			With(&dsl.WithConfig{
				Parts: []dsl.WithPart{
					{
						Name: "row",
					},
					{
						Name: "startNode",
					},
				},
			}).
			Match(dsl.Path().V(dsl.V{Name: "endNode"}).Build()).
			Where(dsl.C(&dsl.ConditionConfig{
				FieldManipulationFunction: "ID",
				Name: "endNode",
				ConditionOperator: dsl.EqualToOperator,
				Check: dsl.ParamString("row.endNodeId"),
			})).
			Merge(&dsl.MergeConfig{
				Path: mergePath,
			}).
			Cypher("SET rel += row.props").
			WithNeo(conn).
			Exec(map[string]interface{}{
				"rows": params,
			})
		if err != nil{
			return err
		}
	}

	return nil
}

func parseValidate(currentDepth, maxDepth int, current *reflect.Value, nodesPtr *map[string]map[string]nodeCreateConf, relationsPtr *map[string][]relCreateConf) error{
	if currentDepth > maxDepth{
		return nil
	}

	if nodesPtr == nil || relationsPtr == nil{
		return errors.New("nodesPtr and/or relationsPtr can not be nil")
	}

	if current == nil{
		return errors.New("current should never be nil")
	}

	return nil
}

func parseStruct(parentId, edgeLabel string, parentIsStart bool, direction dsl.Direction, edgeParams map[string]interface{}, current *reflect.Value, currentDepth int, maxDepth int, nodesPtr *map[string]map[string]nodeCreateConf, relationsPtr *map[string][]relCreateConf) error{
	//check if its done
	if currentDepth > maxDepth{
		return nil
	}

	log.Debugf("on cycle %v", currentDepth)

	//validate params
	err := parseValidate(currentDepth, maxDepth, current, nodesPtr, relationsPtr)
	if err != nil{
		return err
	}

	//get the type
	tString, err := getTypeName(current.Type())
	if err != nil{
		return err
	}

	//get the config
	actual, ok := mappedTypes.Get(tString)
	if !ok{
		return fmt.Errorf("struct config not found type (%s)", tString)
	}

	//cast the config
	currentConf, ok := actual.(structDecoratorConfig)
	if !ok{
		return errors.New("unable to cast into struct decorator config")
	}

	//set this to the actual field name later
	isNewNode, id, err := setUuidIfNeeded(current, "UUID")
	if err != nil{
		return err
	}

	//convert params
	params, err := toCypherParamsMap(*current, currentConf)
	if err != nil{
		return err
	}

	//if its nil, just default it
	if params == nil{
		params = map[string]interface{}{}
	}

	//set the map
	if _, ok := (*nodesPtr)[currentConf.Label]; !ok{
		(*nodesPtr)[currentConf.Label] = map[string]nodeCreateConf{}
	}

	(*nodesPtr)[currentConf.Label][id] = nodeCreateConf{
		Type: current.Type(),
		IsNew: isNewNode,
		Params: params,
	}

	//set edge
	if parentId != ""{
		if _, ok := (*relationsPtr)[edgeLabel]; !ok{
			(*relationsPtr)[edgeLabel] = []relCreateConf{}
		}

		start := ""
		end := ""

		if parentIsStart {
			start = parentId
			end = id
		} else {
			start = id
			end = parentId
		}

		if edgeParams == nil{
			edgeParams = map[string]interface{}{}
		}

		(*relationsPtr)[edgeLabel] = append((*relationsPtr)[edgeLabel], relCreateConf{
			Direction: direction,
			Params: edgeParams,
			StartNodeUUID: start,
			EndNodeUUID: end,
		})
	}

	for _, conf := range currentConf.Fields {
		if conf.Relationship == "" {
			continue
		}

		relField := reflect.Indirect(*current).FieldByName(conf.FieldName)

		//if its nil, just skip it
		if relField.IsNil() {
			continue
		}

		if conf.ManyRelationship{
			slLen := relField.Len()
			if slLen == 0{
				continue
			}

			for i := 0; i < slLen; i++{
				relVal := relField.Index(i)

				newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, skip, err := processStruct(conf, &relVal, id, parentId)
				if err != nil{
					return err
				}

				if skip{
					continue
				}

				err = parseStruct(newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, currentDepth + 1, maxDepth, nodesPtr, relationsPtr)
				if err != nil{
					return err
				}
			}
		} else {
			newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, skip, err := processStruct(conf, &relField, id, parentId)
			if err != nil{
				return err
			}

			if skip{
				continue
			}

			err = parseStruct(newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, currentDepth + 1, maxDepth, nodesPtr, relationsPtr)
			if err != nil{
				return err
			}
		}
	}

	return nil
}

func processStruct(fieldConf decoratorConfig, relVal *reflect.Value, id, oldParentId string) (parentId, edgeLabel string, parentIsStart bool, direction dsl.Direction, edgeParams map[string]interface{}, followVal *reflect.Value, skip bool, err error){
	edgeLabel = fieldConf.Relationship

	relValName, err := getTypeName(relVal.Type())
	if err != nil {
		return "", "", false, 0, nil, nil, false, err
	}

	actual, ok := mappedTypes.Get(relValName)
	if !ok {
		return "", "", false, 0, nil, nil, false, fmt.Errorf("cannot find config for %s", edgeLabel)
	}

	edgeConf, ok := actual.(structDecoratorConfig)
	if !ok{
		return "", "", false, 0, nil, nil, false, errors.New("can not cast to structDecoratorConfig")
	}

	if relVal.Type().Implements(edgeType){
		startValSlice := relVal.MethodByName("GetStartNode").Call(nil)
		endValSlice := relVal.MethodByName("GetEndNode").Call(nil)

		if len(startValSlice) == 0 || len(endValSlice) == 0{
			return "", "", false, 0, nil, nil, false, errors.New("edge is invalid, sides are not set")
		}


		startId := reflect.Indirect(startValSlice[0].Elem()).FieldByName("UUID").String()
		endId := reflect.Indirect(endValSlice[0].Elem()).FieldByName("UUID").String()

		params, err := toCypherParamsMap(*relVal, edgeConf)
		if err != nil{
			return "", "", false, 0, nil, nil, false, err
		}

		//if its nil, just default it
		if params == nil{
			params = map[string]interface{}{}
		}

		if startId == id{

			//check that we're not going in circles
			if oldParentId != ""{
				if endId == oldParentId{
					return "", "", false, 0, nil, &reflect.Value{}, true, nil
				}
			}

			//follow the end
			retVal := endValSlice[0].Elem()
			return startId, edgeLabel, true, fieldConf.Direction, params, &retVal, false, nil
		} else if endId == id{
			///follow the start
			retVal := startValSlice[0].Elem()
			if oldParentId != ""{
				if startId == oldParentId{
					return "", "", false, 0, nil, &reflect.Value{}, true, nil
				}
			}
			return endId, edgeLabel, false, fieldConf.Direction, params, &retVal, false, nil
		} else {
			return "", "", false, 0, nil, nil, false, errors.New("edge is invalid, doesn't point to parent vertex")
		}
	} else {
		if oldParentId != ""{
			if relVal.Kind() == reflect.Ptr{
				*relVal = relVal.Elem()
			}
			if relVal.FieldByName("UUID").String() == oldParentId{
				return "", "", false, 0, nil, &reflect.Value{}, true, nil
			}
		}
		return id, edgeLabel, fieldConf.Direction == dsl.DirectionOutgoing, fieldConf.Direction, map[string]interface{}{}, relVal, false, nil
	}
}