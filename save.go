package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
	driver "github.com/mindstand/golang-neo4j-bolt-driver"
	"reflect"
	"sync"
)

const maxSaveDepth = 10
const defaultSaveDepth = 1

type nodeCreateConf struct {
	Params map[string]interface{}
	Type   reflect.Type
	IsNew  bool
}

type relCreateConf struct {
	StartNodeUUID string
	EndNodeUUID   string
	Params        map[string]interface{}
	Direction     dsl.Direction
}

type relDelConf struct {
	StartNodeId string
	EndNodeId int64
}

func saveDepth(sess *driver.BoltConn, obj interface{}, depth int) error {
	if sess == nil {
		return errors.New("session can not be nil")
	}

	if obj == nil {
		return errors.New("obj can not be nil")
	}

	if depth < 0 {
		return errors.New("cannot save a depth less than 0")
	}

	if depth > maxSaveDepth {
		return fmt.Errorf("saving depth of (%v) is currently not supported, maximum depth is (%v), %w", depth, maxSaveDepth, ErrConfiguration)
	}

	//validate that obj is a pointer
	rawType := reflect.TypeOf(obj)

	if rawType.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be of type pointer, not %T", obj)
	}

	//validate that the dereference type is a struct
	derefType := rawType.Elem()

	if derefType.Kind() != reflect.Struct {
		return fmt.Errorf("dereference type can not be of type %T", obj)
	}

	//signature is [LABEL][UUID]{config}
	nodes := map[string]map[string]nodeCreateConf{}

	//signature is [LABEL] []{config}
	relations := map[string][]relCreateConf{}

	// node id -- [field] config
	oldRels := map[string]map[string]*RelationConfig{}
	curRels := map[string]map[string]*RelationConfig{}
	newIdRef := int64(-1)
	newIdMap := map[int64]string{}

	// uuid -> reflect value
	nodeRef := map[string]*reflect.Value{}

	newNodes := []*string{}

	rootVal := reflect.ValueOf(obj)

	err := parseStruct("", "", false, 0, nil, &rootVal, 0, depth, &nodes, &relations, &curRels, &oldRels, &newNodes, &nodeRef, newIdRef, &newIdMap)
	if err != nil {
		return err
	}

	ids, err := createNodes(sess, nodes)
	if err != nil {
		return err
	}

	//calculate dels
	dels := calculateDels(oldRels, curRels)

	var wg sync.WaitGroup
	var err1, err2, err3 error
	//fix the cur rels and write them to their perspective nodes
	wg.Add(1)
	go func(wg *sync.WaitGroup, _curRels *map[string]map[string]*RelationConfig, _nodeRef *map[string]*reflect.Value, _ids *map[string]int64, _newIdMap *map[int64]string, _err *error) {
		if len(newNodes) != 0 {
			for _, relConfs := range *_curRels {
				for _, relConf := range relConfs {
					//if len(relConf.UUIDs) != 0 {
					//	for _, uuid := range relConf.UUIDs {
					//		if id, ok := (*_ids)[uuid]; ok {
					//			relConf.Ids = append(relConf.Ids, id)
					//		} else {
					//			_err = fmt.Errorf("id not found for node [%s]", uuid)
					//			return
					//		}
					//	}
					//}

					//todo make this more efficient
					for i, id := range relConf.Ids {
						if id < 0 {
							if uuid, ok := (*_newIdMap)[id]; ok {
								if actualId, ok := (*_ids)[uuid]; ok {
									relConf.Ids[i] = actualId
								} else {
									*_err = fmt.Errorf("actual id not found for uuid [%s]", uuid)
									wg.Done()
									return
								}
							}  else {
								*_err = fmt.Errorf("uuid not found for holder id [%v]", id)
								wg.Done()
								return
							}
						}
					}

					//reset uuids
					relConf.UUIDs = []string{}
				}
			}
		}

		for uuid, val := range *_nodeRef {
			loadConf, ok := (*_curRels)[uuid]
			if !ok {
				*_err = fmt.Errorf("load config not found for node [%s]", uuid)
				wg.Done()
				return
			}

			//handle if its a pointer
			if val.Kind() == reflect.Ptr {
				*val = val.Elem()
			}

			id, ok := (*_ids)[uuid]
			if !ok {
				*_err = fmt.Errorf("GraphId not found for node [%s]", uuid)
				wg.Done()
				return
			}

			reflect.Indirect(*val).FieldByName("LoadMap").Set(reflect.ValueOf(loadConf))
			reflect.Indirect(*val).FieldByName("Id").Set(reflect.ValueOf(id))
		}
		wg.Done()
	}(&wg, &curRels, &nodeRef, &ids, &newIdMap, &err3)

	//execute concurrently

	if len(dels) != 0 {
		wg.Add(1)

		go func(wg *sync.WaitGroup, _dels map[string][]int64, _conn *driver.BoltConn, _err *error) {
			err := removeRelations(_conn, _dels)
			if err != nil {
				*_err = err
			}
			wg.Done()
		}(&wg, dels, sess, &err1)
	}

	if len(relations) != 0 {
		wg.Add(1)
		go func(wg *sync.WaitGroup, _conn *driver.BoltConn, _relations map[string][]relCreateConf, _ids map[string]int64, _err *error) {
			err := relateNodes(_conn, _relations, _ids)
			if err != nil {
				*_err = err
			}
			wg.Done()
		}(&wg, sess, relations, ids, &err2)
	}

	wg.Wait()

	if err1 != nil || err2 != nil || err3 != nil{
		return fmt.Errorf("delErr=(%v) | relErr=(%v) | reallocErr=(%v)", err1, err2, err3)
	} else {
		return nil
	}
}

func calculateDels(oldRels, curRels map[string]map[string]*RelationConfig) map[string][]int64 {
	if len(oldRels) == 0 {
		return map[string][]int64{}
	}

	dels := map[string][]int64{}

	for uuid, oldRelConf := range oldRels {
		curRelConf, ok := curRels[uuid]
		deleteAllRels := false
		if !ok {
			//this means that the node is gone, remove all rels to this node
			deleteAllRels = true
		} else {
			for field, oldConf := range oldRelConf {
				curConf, ok := curRelConf[field]
				deleteAllRelsOnField := false
				if !ok {
					//this means that either the field has been removed or there are no more rels on this field,
					//either way delete anything left over
					deleteAllRelsOnField = true
				}
				for _, id := range oldConf.Ids {
					//check if this id is new rels in the same location
					if deleteAllRels || deleteAllRelsOnField{
						if _, ok := dels[uuid]; !ok {
							dels[uuid] = []int64{id}
						} else {
							dels[uuid] = append(dels[uuid], id)
						}
					} else {
						if !int64SliceContains(curConf.Ids, id) {
							if _, ok := dels[uuid]; !ok {
								dels[uuid] = []int64{id}
							} else {
								dels[uuid] = append(dels[uuid], id)
							}
						}
					}
				}
			}
		}
	}

	return dels
}

func removeRelations(conn *driver.BoltConn, dels map[string][]int64) error {
	if dels == nil || len(dels) == 0 {
		return nil
	}

	if conn == nil {
		return fmt.Errorf("connection can not be nil, %w", ErrInternal)
	}

	var params []interface{}

	for uuid, ids := range dels {
		params = append(params, map[string]interface{}{
			"startNodeId": uuid,
			"endNodeIds": ids,
		})
	}

	startParams, err := dsl.ParamsFromMap(map[string]interface{}{
		"uuid": dsl.ParamString("{row.startNodeId}"),
	})
	if err != nil {
		return fmt.Errorf("%s, %w", err.Error(), ErrInternal)
	}

	res, err := dsl.QB().
		Cypher("UNWIND {rows} as row").
		Match(dsl.Path().
			V(dsl.V{
				Name: "start",
				Params: startParams,
			}).E(dsl.E{
				Name: "e",
			}).V(dsl.V{
				Name: "end",
			}).Build()).
		Cypher("id(end) IN row.endNodeIds").
		Delete(false, "e").
		WithNeo(conn).
		Exec(map[string]interface{}{
			"rows": params,
		},
	)
	if err != nil {
		return fmt.Errorf("%s, %w", err.Error(), ErrInternal)
	}

	if rows, err := res.RowsAffected(); err != nil {
		return fmt.Errorf("%s, %w", err.Error(), ErrInternal)
	} else if int(rows) != len(dels) {
		return fmt.Errorf("sanity check failed, rows affected [%v] not equal to num deletions [%v], %w", rows, len(dels), ErrInternal)
	} else {
		return nil
	}
}

func createNodes(conn *driver.BoltConn, crNodes map[string]map[string]nodeCreateConf) (map[string]int64, error) {
	idMap := map[string]int64{}

	for label, nodes := range crNodes {
		var rows []interface{}
		for _, config := range nodes {
			rows = append(rows, config.Params)
		}

		params, err := dsl.ParamsFromMap(
			map[string]interface{}{
				"uuid": dsl.ParamString("row.uuid"),
			})
		if err != nil {
			return nil, err
		}

		path, err := dsl.Path().V(dsl.V{
			Name:   "n",
			Type:   "`" + label + "`",
			Params: params,
		}).ToCypher()
		if err != nil {
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
				Name:  "row.uuid",
				Alias: "uuid",
			}, dsl.ReturnPart{
				Function: &dsl.FunctionConfig{
					Name:   "ID",
					Params: []interface{}{dsl.ParamString("n")},
				},
				Alias: "id",
			}).
			WithNeo(conn).
			Query(map[string]interface{}{
				"rows": rows,
			})
		if err != nil {
			return nil, err
		}

		if res == nil {
			return nil, errors.New("res should not be nil")
		}

		resRows, _, err := res.All()
		if err != nil {
			return nil, err
		}

		for _, row := range resRows {
			if len(row) != 2 {
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

func relateNodes(conn *driver.BoltConn, relations map[string][]relCreateConf, ids map[string]int64) error {
	if relations == nil || len(relations) == 0 {
		return errors.New("relations can not be nil or empty")
	}

	if ids == nil || len(ids) == 0 {
		return errors.New("ids can not be nil or empty")
	}

	for label, rels := range relations {
		var params []interface{}

		if len(rels) == 0 {
			continue
		}

		for _, rel := range rels {
			startId, ok := ids[rel.StartNodeUUID]
			if !ok {
				return fmt.Errorf("can not find nodeId for uuid %s", rel.StartNodeUUID)
			}

			endId, ok := ids[rel.EndNodeUUID]
			if !ok {
				return fmt.Errorf("can not find nodeId for %T", rel)
			}

			//set map if its empty
			if rel.Params == nil {
				rel.Params = map[string]interface{}{}
			}

			params = append(params, map[string]interface{}{
				"startNodeId": startId,
				"endNodeId":   endId,
				"props":       rel.Params,
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
		if err != nil {
			return err
		}

		_, err = dsl.QB().
			Cypher("UNWIND {rows} as row").
			Match(dsl.Path().V(dsl.V{Name: "startNode"}).Build()).
			Where(dsl.C(&dsl.ConditionConfig{
				FieldManipulationFunction: "ID",
				Name:                      "startNode",
				ConditionOperator:         dsl.EqualToOperator,
				Check:                     dsl.ParamString("row.startNodeId"),
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
				Name:                      "endNode",
				ConditionOperator:         dsl.EqualToOperator,
				Check:                     dsl.ParamString("row.endNodeId"),
			})).
			Merge(&dsl.MergeConfig{
				Path: mergePath,
			}).
			Cypher("SET rel += row.props").
			WithNeo(conn).
			Exec(map[string]interface{}{
				"rows": params,
			})
		if err != nil {
			return err
		}
	}

	return nil
}

func parseValidate(currentDepth, maxDepth int, current *reflect.Value, nodesPtr *map[string]map[string]nodeCreateConf, relationsPtr *map[string][]relCreateConf) error {
	if currentDepth > maxDepth {
		return nil
	}

	if nodesPtr == nil || relationsPtr == nil {
		return errors.New("nodesPtr and/or relationsPtr can not be nil")
	}

	if current == nil {
		return errors.New("current should never be nil")
	}

	return nil
}

func parseStruct(parentId, edgeLabel string, parentIsStart bool, direction dsl.Direction, edgeParams map[string]interface{}, current *reflect.Value,
	currentDepth int, maxDepth int, nodesPtr *map[string]map[string]nodeCreateConf, relationsPtr *map[string][]relCreateConf, curRels , oldRels *map[string]map[string]*RelationConfig,
	newNodes *[]*string, nodeRef *map[string]*reflect.Value, newIdRef int64, newIdMap *map[int64]string) error {
	//check if its done
	if currentDepth > maxDepth {
		return nil
	}

	log.Debugf("on cycle %v", currentDepth)

	//validate params
	err := parseValidate(currentDepth, maxDepth, current, nodesPtr, relationsPtr)
	if err != nil {
		return err
	}

	//get the type
	tString, err := getTypeName(current.Type())
	if err != nil {
		return err
	}

	//get the config
	actual, ok := mappedTypes.Get(tString)
	if !ok {
		return fmt.Errorf("struct config not found type (%s)", tString)
	}

	//cast the config
	currentConf, ok := actual.(structDecoratorConfig)
	if !ok {
		return errors.New("unable to cast into struct decorator config")
	}

	//set this to the actual field name later
	isNewNode, id, err := setUuidIfNeeded(current, "UUID")
	if err != nil {
		return err
	}

	// set old rels to current load map of the node
	(*curRels)[id] = map[string]*RelationConfig{}

	if !isNewNode {
		if _, ok := (*oldRels)[id]; !ok{
			iConf := reflect.Indirect(*current).FieldByName("LoadMap").Interface()

			var relConf map[string]*RelationConfig

			if iConf != nil {
				relConf, ok = iConf.(map[string]*RelationConfig)
				if !ok {
					relConf = map[string]*RelationConfig{}
				}
			}

			(*oldRels)[id] = relConf
		}
	} else {
		//if parentField != "" {
		//	if _, ok := (*curRels)[parentId]; !ok {
		//		(*curRels)[parentId] = map[string]*RelationConfig{}
		//	}
		//
		//	if (*curRels)[parentId][parentField] == nil {
		//		(*curRels)[parentId][parentField] = &RelationConfig{
		//			Ids: []int64{},
		//			UUIDs: []string{},
		//			RelationType: 0,
		//		}
		//	}
		//	(*curRels)[parentId][parentField].UUIDs = append((*curRels)[parentId][parentField].UUIDs, id)
		//}
		//this was already set i guess
		for {
			if _, ok :=(*newIdMap)[newIdRef]; ok {
				newIdRef--
			} else {
				break
			}
		}

		(*newIdMap)[newIdRef] = id
		*newNodes = append(*newNodes, &id)
	}

	//set the reflect pointer so we can update the map later
	if _, ok := (*nodeRef)[id]; !ok {
		(*nodeRef)[id] = current
	}

	//convert params
	params, err := toCypherParamsMap(*current, currentConf)
	if err != nil {
		return err
	}

	//if its nil, just default it
	if params == nil {
		params = map[string]interface{}{}
	}

	//set the map
	if _, ok := (*nodesPtr)[currentConf.Label]; !ok {
		(*nodesPtr)[currentConf.Label] = map[string]nodeCreateConf{}
	}

	(*nodesPtr)[currentConf.Label][id] = nodeCreateConf{
		Type:   current.Type(),
		IsNew:  isNewNode,
		Params: params,
	}

	//set edge
	if parentId != "" {
		if _, ok := (*relationsPtr)[edgeLabel]; !ok {
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

		if edgeParams == nil {
			edgeParams = map[string]interface{}{}
		}

		(*relationsPtr)[edgeLabel] = append((*relationsPtr)[edgeLabel], relCreateConf{
			Direction:     direction,
			Params:        edgeParams,
			StartNodeUUID: start,
			EndNodeUUID:   end,
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

		if conf.ManyRelationship {
			slLen := relField.Len()
			if slLen == 0 {
				continue
			}

			for i := 0; i < slLen; i++ {
				relVal := relField.Index(i)

				newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, followId, skip, err := processStruct(conf, &relVal, id, parentId)
				if err != nil {
					return err
				}

				//add the current relation
				if _, ok := (*curRels)[id]; !ok {
					(*curRels)[id] = map[string]*RelationConfig{}
				}

				if _, ok = (*curRels)[id][conf.FieldName]; !ok {
					(*curRels)[id][conf.FieldName] = &RelationConfig{
						Ids:          []int64{},
						UUIDs: []string{},
						RelationType: Multi,
					}
				}

				if followId == 0 {
					var t int64
					if parentId == "" {
						t = newIdRef - 1
					} else {
						t = newIdRef + 1
					}
					(*curRels)[id][conf.FieldName].Ids = append((*curRels)[id][conf.FieldName].Ids, t)
					newIdRef--
				} else {
					(*curRels)[id][conf.FieldName].Ids = append((*curRels)[id][conf.FieldName].Ids, followId)
				}

				//makes us go backwards
				if skip {
					continue
				}
				
				err = parseStruct(newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, currentDepth+1, maxDepth, nodesPtr, relationsPtr, curRels, oldRels, newNodes, nodeRef, newIdRef, newIdMap)
				if err != nil {
					return err
				}
			}
		} else {
			newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, followId, skip, err := processStruct(conf, &relField, id, parentId)
			if err != nil {
				return err
			}

			//add the current relation
			if _, ok := (*curRels)[id]; !ok {
				(*curRels)[id] = map[string]*RelationConfig{}
			}

			if _, ok = (*curRels)[id][conf.FieldName]; !ok {
				(*curRels)[id][conf.FieldName] = &RelationConfig{
					Ids:          []int64{},
					UUIDs: []string{},
					RelationType: Single,
				}
			}

			if followId == 0 {
				var t int64
				if parentId == "" {
					t = newIdRef - 1
				} else {
					t = newIdRef + 1
				}
				(*curRels)[id][conf.FieldName].Ids = append((*curRels)[id][conf.FieldName].Ids, t)
				//newIdRef--
			} else {
				(*curRels)[id][conf.FieldName].Ids = append((*curRels)[id][conf.FieldName].Ids, followId)
			}

			if skip {
				continue
			}

			err = parseStruct(newParentId, newEdgeLabel, newParentIdStart, newDirection, newEdgeParams, followVal, currentDepth+1, maxDepth, nodesPtr, relationsPtr, curRels, oldRels, newNodes, nodeRef, newIdRef, newIdMap)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func processStruct(fieldConf decoratorConfig, relVal *reflect.Value, id, oldParentId string) (parentId, edgeLabel string, parentIsStart bool, direction dsl.Direction, edgeParams map[string]interface{}, followVal *reflect.Value, followId int64, skip bool, err error) {
	edgeLabel = fieldConf.Relationship

	relValName, err := getTypeName(relVal.Type())
	if err != nil {
		return "", "", false, 0, nil, nil, -1, false, err
	}

	actual, ok := mappedTypes.Get(relValName)
	if !ok {
		return "", "", false, 0, nil, nil, -1, false, fmt.Errorf("cannot find config for %s", edgeLabel)
	}

	edgeConf, ok := actual.(structDecoratorConfig)
	if !ok {
		return "", "", false, 0, nil, nil, -1, false, errors.New("can not cast to structDecoratorConfig")
	}

	if relVal.Type().Implements(edgeType) {
		startValSlice := relVal.MethodByName("GetStartNode").Call(nil)
		endValSlice := relVal.MethodByName("GetEndNode").Call(nil)

		if len(startValSlice) == 0 || len(endValSlice) == 0 {
			return "", "", false, 0, nil, nil, -1, false, errors.New("edge is invalid, sides are not set")
		}

		startId := reflect.Indirect(startValSlice[0].Elem()).FieldByName("UUID").String()
		endId := reflect.Indirect(endValSlice[0].Elem()).FieldByName("UUID").String()

		params, err := toCypherParamsMap(*relVal, edgeConf)
		if err != nil {
			return "", "", false, 0, nil, nil, -1, false, err
		}

		//if its nil, just default it
		if params == nil {
			params = map[string]interface{}{}
		}

		if startId == id {

			//follow the end
			retVal := endValSlice[0].Elem()

			Iid := reflect.Indirect(retVal).FieldByName("Id").Interface()

			followId, ok := Iid.(int64)
			if !ok {
				followId = 0
			}

			//check that we're not going in circles
			if oldParentId != "" {
				if endId == oldParentId {
					return "", "", false, 0, nil, &reflect.Value{}, followId, true, nil
				}
			}

			return startId, edgeLabel, true, fieldConf.Direction, params, &retVal, followId, false, nil
		} else if endId == id {
			///follow the start
			retVal := startValSlice[0].Elem()

			Iid := reflect.Indirect(retVal).FieldByName("Id").Interface()

			followId, ok := Iid.(int64)
			if !ok {
				followId = 0
			}

			if oldParentId != "" {
				if startId == oldParentId {
					return "", "", false, 0, nil, &reflect.Value{}, followId, true, nil
				}
			}

			return endId, edgeLabel, false, fieldConf.Direction, params, &retVal, followId, false, nil
		} else {
			return "", "", false, 0, nil, nil, -1, false, errors.New("edge is invalid, doesn't point to parent vertex")
		}
	} else {
		var followId int64

		if oldParentId != "" {
			if relVal.Kind() == reflect.Ptr {
				*relVal = relVal.Elem()
			}

			Iid := reflect.Indirect(*relVal).FieldByName("Id").Interface()

			followId, ok = Iid.(int64)
			if !ok {
				followId = 0
			}

			if relVal.FieldByName("UUID").String() == oldParentId {
				return "", "", false, 0, nil, &reflect.Value{}, followId, true, nil
			}
		} else {
			Iid := reflect.Indirect(*relVal).FieldByName("Id").Interface()

			followId, ok = Iid.(int64)
			if !ok {
				followId = 0
			}
		}

		return id, edgeLabel, fieldConf.Direction == dsl.DirectionOutgoing, fieldConf.Direction, map[string]interface{}{}, relVal, followId, false, nil
	}
}
