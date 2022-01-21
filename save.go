// Copyright (c) 2021 MindStand Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package gogm

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	dsl "github.com/mindstand/go-cypherdsl"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

// nodeCreate holds configuration for creating new nodes
type nodeCreate struct {
	// params to save
	Params map[string]interface{}
	// type to save by
	Type reflect.Type
	// Id
	Id int64
	// Pointer value (if id is not yet set)
	Pointer uintptr
	// whether the node is new or not
	IsNew bool
}

// relCreate holds configuration for nodes to link together
type relCreate struct {
	// start uuid of relationship
	StartNodePtr uintptr
	// end uuid of relationship
	EndNodePtr uintptr
	// any data to store in edge
	Params map[string]interface{}
	// holds direction of the edge
	Direction dsl.Direction
}

func saveDepth(gogm *Gogm, obj interface{}, depth int) neo4j.TransactionWork {
	return func(tx neo4j.Transaction) (interface{}, error) {
		if obj == nil {
			return nil, errors.New("obj can not be nil")
		}

		if depth < 0 {
			return nil, errors.New("cannot save a depth less than 0")
		}

		//validate that obj is a pointer
		rawType := reflect.TypeOf(obj)

		if rawType.Kind() != reflect.Ptr {
			return nil, fmt.Errorf("obj must be of type pointer, not %T", obj)
		}

		//validate that the dereference type is a struct
		derefType := rawType.Elem()

		if derefType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("dereference type can not be of type %T", obj)
		}

		var (
			// [LABEL][int64 (graphid) or uintptr]{config}
			nodes = map[string]map[uintptr]*nodeCreate{}
			// [LABEL] []{config}
			relations = map[string][]*relCreate{}
			// node id -- [field] config
			oldRels = map[uintptr]map[string]*RelationConfig{}
			// node id -- [field] config
			curRels = map[int64]map[string]*RelationConfig{}
			// id to reflect value
			nodeIdRef = map[uintptr]int64{}
			// uintptr to reflect value (for new nodes that dont have a graph id yet)
			nodeRef = map[uintptr]*reflect.Value{}
		)

		rootVal := reflect.ValueOf(obj)
		err := parseStruct(gogm, 0, "", false, dsl.DirectionBoth, nil, &rootVal, 0, depth,
			nodes, relations, nodeIdRef, nodeRef, oldRels)
		if err != nil {
			return nil, fmt.Errorf("failed to parse struct, %w", err)
		}
		// save/update nodes
		err = createNodes(tx, nodes, nodeRef, nodeIdRef)
		if err != nil {
			return nil, fmt.Errorf("failed to create nodes, %w", err)
		}

		// generate rel maps
		err = generateCurRels(gogm, 0, &rootVal, 0, depth, curRels)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate current relationships, %w", err)
		}

		dels, err := calculateDels(oldRels, curRels, nodeIdRef)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate relationships to delete, %w", err)
		}

		//fix the cur rels and write them to their perspective nodes
		for ptr, val := range nodeRef {
			graphId, ok := nodeIdRef[ptr]
			if !ok {
				return nil, fmt.Errorf("graph id for node ptr [%v] not found", ptr)
			}
			loadConf, ok := curRels[graphId]
			if !ok {
				return nil, fmt.Errorf("load config not found for node [%v]", graphId)
			}

			//handle if its a pointer
			if val.Kind() == reflect.Ptr {
				*val = val.Elem()
			}

			reflect.Indirect(*val).FieldByName("LoadMap").Set(reflect.ValueOf(loadConf))
		}

		if len(dels) != 0 {
			err := removeRelations(tx, dels)
			if err != nil {
				return nil, err
			}
		}

		if len(relations) != 0 {
			err := relateNodes(tx, relations, nodeIdRef)
			if err != nil {
				return nil, err
			}
		}

		return obj, nil
	}
}

// relateNodes connects nodes together using edge config
func relateNodes(transaction neo4j.Transaction, relations map[string][]*relCreate, lookup map[uintptr]int64) error {
	if len(relations) == 0 {
		return errors.New("relations can not be nil or empty")
	}

	for label, rels := range relations {
		var params []interface{}

		if len(rels) == 0 {
			continue
		}

		for _, rel := range rels {
			// grab start id
			startId, ok := lookup[rel.StartNodePtr]
			if !ok {
				return fmt.Errorf("graph id not found for ptr %v", rel.StartNodePtr)
			}

			endId, ok := lookup[rel.EndNodePtr]
			if !ok {
				return fmt.Errorf("graph id not found for ptr %v", rel.EndNodePtr)
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

		cyp, err := dsl.QB().
			Cypher("UNWIND $rows as row").
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
			ToCypher()
		if err != nil {
			return fmt.Errorf("failed to build query, %w", err)
		}

		res, err := transaction.Run(cyp, map[string]interface{}{
			"rows": params,
		})
		if err != nil {
			return fmt.Errorf("failed to relate nodes, %w", err)
		} else if err = res.Err(); err != nil {
			return fmt.Errorf("failed to relate nodes %w", res.Err())
		}
	}

	return nil
}

// removes relationships between specified nodes
func removeRelations(transaction neo4j.Transaction, dels map[int64][]int64) error {
	if len(dels) == 0 {
		return nil
	}

	var params []interface{}

	for id, ids := range dels {
		params = append(params, map[string]interface{}{
			"startNodeId": id,
			"endNodeIds":  ids,
		})
	}

	cyq, err := dsl.QB().
		Cypher("UNWIND $rows as row").
		Match(dsl.Path().
			V(dsl.V{
				Name: "start",
			}).E(dsl.E{
			Name: "e",
		}).V(dsl.V{
			Name: "end",
		}).Build()).
		Cypher("WHERE id(start) = row.startNodeId and id(end) in row.endNodeIds").
		Delete(false, "e").
		ToCypher()
	if err != nil {
		return err
	}

	res, err := transaction.Run(cyq, map[string]interface{}{
		"rows": params,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	} else if err = res.Err(); err != nil {
		return fmt.Errorf("%s: %w", err.Error(), ErrInternal)
	}
	//todo sanity check to make sure the affects worked

	return nil
}

// calculates which relationships to delete
func calculateDels(oldRels map[uintptr]map[string]*RelationConfig, curRels map[int64]map[string]*RelationConfig, lookup map[uintptr]int64) (map[int64][]int64, error) {
	if len(oldRels) == 0 {
		return map[int64][]int64{}, nil
	}

	dels := map[int64][]int64{}

	for ptr, oldRelConf := range oldRels {
		oldId, ok := lookup[ptr]
		if !ok {
			return nil, fmt.Errorf("graph id not found for ptr [%v]", ptr)
		}
		curRelConf, ok := curRels[oldId]
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
					if deleteAllRels || deleteAllRelsOnField {
						if _, ok := dels[oldId]; !ok {
							dels[oldId] = []int64{id}
						} else {
							dels[oldId] = append(dels[oldId], id)
						}
					} else {
						if !int64SliceContains(curConf.Ids, id) {
							if _, ok := dels[oldId]; !ok {
								dels[oldId] = []int64{id}
							} else {
								dels[oldId] = append(dels[oldId], id)
							}
						}
					}
				}
			}
		}
	}

	return dels, nil
}

func generateCurRels(gogm *Gogm, parentPtr uintptr, current *reflect.Value, currentDepth, maxDepth int, curRels map[int64]map[string]*RelationConfig) error {
	if currentDepth > maxDepth {
		return nil
	}

	curPtr := current.Pointer()

	// check for going in circles
	if parentPtr == curPtr {
		return nil
	}

	idVal := reflect.Indirect(*current).FieldByName(DefaultPrimaryKeyStrategy.FieldName)
	if idVal.IsNil() {
		return errors.New("id not set")
	}

	var id int64

	if !idVal.Elem().IsZero() {
		id = idVal.Elem().Int()
	} else {
		id = 0
	}

	if _, ok := curRels[id]; ok {
		//this node has already been seen
		return nil
	} else {
		//create the record for it
		curRels[id] = map[string]*RelationConfig{}
	}

	//get the type
	tString, err := getTypeName(current.Type())
	if err != nil {
		return err
	}

	//get the config
	actual, ok := gogm.mappedTypes.Get(tString)
	if !ok {
		return fmt.Errorf("struct config not found type (%s)", tString)
	}

	//cast the config
	currentConf, ok := actual.(structDecoratorConfig)
	if !ok {
		return errors.New("unable to cast into struct decorator config")
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

				newParentId, _, _, _, _, followVal, err := processStruct(gogm, conf, &relVal, curPtr)
				if err != nil {
					return err
				}

				followIdVal := reflect.Indirect(*followVal).FieldByName(DefaultPrimaryKeyStrategy.FieldName)
				var followId int64
				if !followIdVal.IsNil() {
					followIdVal = followIdVal.Elem()
					if followIdVal.IsZero() {
						followId = 0
					} else {
						followId = followIdVal.Int()
					}
				} else {
					followId = 0
				}

				//check the config is there for the specified field
				if _, ok = curRels[id][conf.FieldName]; !ok {
					curRels[id][conf.FieldName] = &RelationConfig{
						Ids:          []int64{},
						RelationType: Multi,
					}
				}

				curRels[id][conf.FieldName].Ids = append(curRels[id][conf.FieldName].Ids, followId)

				if followVal.Pointer() != parentPtr {
					err = generateCurRels(gogm, newParentId, followVal, currentDepth+1, maxDepth, curRels)
					if err != nil {
						return err
					}
				}
			}
		} else {
			newParentId, _, _, _, _, followVal, err := processStruct(gogm, conf, &relField, curPtr)
			if err != nil {
				return err
			}

			followIdVal := reflect.Indirect(*followVal).FieldByName(DefaultPrimaryKeyStrategy.FieldName)
			var followId int64
			if !followIdVal.IsNil() {
				followIdVal = followIdVal.Elem()
				if followIdVal.IsZero() {
					followId = 0
				} else {
					followId = followIdVal.Int()
				}
			} else {
				followId = 0
			}

			//check the config is there for the specified field
			if _, ok = curRels[id][conf.FieldName]; !ok {
				curRels[id][conf.FieldName] = &RelationConfig{
					Ids:          []int64{},
					RelationType: Single,
				}
			}

			curRels[id][conf.FieldName].Ids = append(curRels[id][conf.FieldName].Ids, followId)

			if followVal.Pointer() != parentPtr {
				err = generateCurRels(gogm, newParentId, followVal, currentDepth+1, maxDepth, curRels)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// createNodes updates existing nodes and creates new nodes while also making a lookup table for ptr -> neoid
func createNodes(transaction neo4j.Transaction, crNodes map[string]map[uintptr]*nodeCreate, nodeRef map[uintptr]*reflect.Value, nodeIdRef map[uintptr]int64) error {
	for label, nodes := range crNodes {
		var updateRows, newRows []interface{}
		for ptr, config := range nodes {
			row := map[string]interface{}{
				"obj": config.Params,
			}

			if id, ok := nodeIdRef[ptr]; ok {
				row["id"] = id
				updateRows = append(updateRows, row)
			} else {
				row["ptr"] = fmt.Sprintf("%v", ptr)
				newRows = append(newRows, row)
			}
		}

		// create new stuff
		if len(newRows) != 0 {
			cyp, err := dsl.QB().
				Cypher("UNWIND $rows as row").
				Cypher(fmt.Sprintf("CREATE(n:`%s`)", label)).
				Cypher("SET n += row.obj").
				Return(false, dsl.ReturnPart{
					Name:  "row.ptr",
					Alias: "ptr",
				}, dsl.ReturnPart{
					Function: &dsl.FunctionConfig{
						Name:   "ID",
						Params: []interface{}{dsl.ParamString("n")},
					},
					Alias: "id",
				}).
				ToCypher()
			if err != nil {
				return fmt.Errorf("failed to build query, %w", err)
			}

			res, err := transaction.Run(cyp, map[string]interface{}{
				"rows": newRows,
			})
			if err != nil {
				return fmt.Errorf("failed to execute new node query, %w", err)
			} else if res.Err() != nil {
				return fmt.Errorf("failed to execute new node query from result error, %w", res.Err())
			}

			for res.Next() {
				row := res.Record().Values
				if len(row) != 2 {
					continue
				}

				strPtr, ok := row[0].(string)
				if !ok {
					return fmt.Errorf("cannot cast row[0] to string, %w", ErrInternal)
				}

				ptrInt, err := strconv.ParseUint(strPtr, 10, 64)
				if err != nil {
					return fmt.Errorf("failed to parse ptr string to int64, %w", err)
				}

				ptr := uintptr(ptrInt)

				graphId, ok := row[1].(int64)
				if !ok {
					return fmt.Errorf("cannot cast row[1] to int64, %w", ErrInternal)
				}

				// update the lookup
				nodeIdRef[ptr] = graphId

				//set the new id
				val, ok := nodeRef[ptr]
				if !ok {
					return fmt.Errorf("cannot find val for ptr [%v]", ptr)
				}

				reflect.Indirect(*val).FieldByName(DefaultPrimaryKeyStrategy.FieldName).Set(reflect.ValueOf(&graphId))
			}
		}

		// process stuff that we're updating
		// dont need any data back from this other than did it work
		if len(updateRows) != 0 {
			path, err := dsl.Path().V(dsl.V{
				Name: "n",
				Type: "`" + label + "`",
			}).ToCypher()
			if err != nil {
				return err
			}

			cyp, err := dsl.QB().
				Cypher("UNWIND $rows as row").
				Cypher(fmt.Sprintf("MATCH %s", path)).
				Cypher("WHERE ID(n) = row.id").
				Cypher("SET n += row.obj").
				ToCypher()
			if err != nil {
				return fmt.Errorf("failed to build query, %w", err)
			}

			res, err := transaction.Run(cyp, map[string]interface{}{
				"rows": updateRows,
			})
			if err != nil {
				return fmt.Errorf("failed to run update query, %w", err)
			} else if res.Err() != nil {
				return fmt.Errorf("failed to run update query, %w", res.Err())
			}
		}
	}

	return nil
}

// parseStruct
// we are intentionally using pointers as identifiers in this stage because graph ids are not guaranteed
func parseStruct(gogm *Gogm, parentPtr uintptr, edgeLabel string, parentIsStart bool, direction dsl.Direction, edgeParams map[string]interface{}, current *reflect.Value,
	currentDepth, maxDepth int, nodes map[string]map[uintptr]*nodeCreate, relations map[string][]*relCreate, nodeIdLookup map[uintptr]int64, nodeRef map[uintptr]*reflect.Value, oldRels map[uintptr]map[string]*RelationConfig) error {
	//check if its done
	if currentDepth > maxDepth {
		return nil
	}

	curPtr := current.Pointer()

	//get the type
	nodeType, err := getTypeName(current.Type())
	if err != nil {
		return err
	}

	// get the config
	actual, ok := gogm.mappedTypes.Get(nodeType)
	if !ok {
		return fmt.Errorf("struct config not found type (%s)", nodeType)
	}

	//cast the config
	currentConf, ok := actual.(structDecoratorConfig)
	if !ok {
		return errors.New("unable to cast into struct decorator config")
	}

	// grab info and set ids of current node
	isNew, graphID, relConf, err := handleNodeState(gogm.pkStrategy, current)
	if err != nil {
		return fmt.Errorf("failed to handle node, %w", err)
	}

	// handle edge
	if parentPtr != 0 {
		if _, ok := relations[edgeLabel]; !ok {
			relations[edgeLabel] = []*relCreate{}
		}

		var start, end uintptr
		curDir := direction

		if parentIsStart {
			start = parentPtr
			end = curPtr
		} else {
			start = curPtr
			end = parentPtr
			if curDir == dsl.DirectionIncoming {
				curDir = dsl.DirectionOutgoing
			} else if curDir == dsl.DirectionOutgoing {
				curDir = dsl.DirectionIncoming
			}
		}

		if edgeParams == nil {
			edgeParams = map[string]interface{}{}
		}

		found := false
		//check if this edge is already here
		if len(relations[edgeLabel]) != 0 {
			for _, conf := range relations[edgeLabel] {
				if conf.StartNodePtr == start && conf.EndNodePtr == end {
					found = true
				}
			}
		}

		// if not found already register the relationships
		if !found {
			relations[edgeLabel] = append(relations[edgeLabel], &relCreate{
				Direction:    curDir,
				Params:       edgeParams,
				StartNodePtr: start,
				EndNodePtr:   end,
			})
		}
	}

	if !isNew {
		if _, ok := nodeIdLookup[curPtr]; !ok {
			nodeIdLookup[curPtr] = graphID
		}

		if _, ok := oldRels[curPtr]; !ok {
			oldRels[curPtr] = relConf
		}
	}

	// set the lookup table
	if _, ok := nodeRef[curPtr]; !ok {
		nodeRef[curPtr] = current
	}

	//convert params
	params, err := toCypherParamsMap(gogm, *current, currentConf)
	if err != nil {
		return err
	}

	//if its nil, just default it
	if params == nil {
		params = map[string]interface{}{}
	}

	//set the nodes lookup map
	if _, ok := nodes[currentConf.Label]; !ok {
		nodes[currentConf.Label] = map[uintptr]*nodeCreate{}
	}

	nodes[currentConf.Label][curPtr] = &nodeCreate{
		Params:  params,
		Type:    current.Type(),
		Id:      graphID,
		Pointer: curPtr,
		IsNew:   isNew,
	}

	// loop through fields looking for edges
	for _, conf := range currentConf.Fields {
		if conf.Relationship == "" {
			// not a relationship field
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
				newParentId, newEdgeLabel, newParentIsStart, newDirection, newEdgeParams, followVal, err := processStruct(gogm, conf, &relVal, curPtr)
				if err != nil {
					return err
				}

				err = parseStruct(gogm, newParentId, newEdgeLabel, newParentIsStart, newDirection, newEdgeParams, followVal, currentDepth+1, maxDepth, nodes, relations, nodeIdLookup, nodeRef, oldRels)
				if err != nil {
					return err
				}
			}
		} else {
			newParentId, newEdgeLabel, newParentIsStart, newDirection, newEdgeParams, followVal, err := processStruct(gogm, conf, &relField, curPtr)
			if err != nil {
				return err
			}

			err = parseStruct(gogm, newParentId, newEdgeLabel, newParentIsStart, newDirection, newEdgeParams, followVal, currentDepth+1, maxDepth, nodes, relations, nodeIdLookup, nodeRef, oldRels)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// processStruct generates configuration for individual struct for saving
func processStruct(gogm *Gogm, fieldConf decoratorConfig, relValue *reflect.Value, curPtr uintptr) (parentId uintptr, edgeLabel string, parentIsStart bool, direction dsl.Direction, edgeParams map[string]interface{}, followVal *reflect.Value, err error) {
	edgeLabel = fieldConf.Relationship

	relValName, err := getTypeName(relValue.Type())
	if err != nil {
		return 0, "", false, 0, nil, nil, err
	}

	actual, ok := gogm.mappedTypes.Get(relValName)
	if !ok {
		return 0, "", false, 0, nil, nil, fmt.Errorf("cannot find config for %s", edgeLabel)
	}

	edgeConf, ok := actual.(structDecoratorConfig)
	if !ok {
		return 0, "", false, 0, nil, nil, errors.New("can not cast to structDecoratorConfig")
	}

	if relValue.Type().Implements(edgeType) {
		startValSlice := relValue.MethodByName("GetStartNode").Call(nil)
		endValSlice := relValue.MethodByName("GetEndNode").Call(nil)

		if len(startValSlice) == 0 || len(endValSlice) == 0 {
			return 0, "", false, 0, nil, nil, errors.New("edge is invalid, sides are not set")
		}

		startVal := startValSlice[0].Elem()
		endVal := endValSlice[0].Elem()

		params, err := toCypherParamsMap(gogm, *relValue, edgeConf)
		if err != nil {
			return 0, "", false, 0, nil, nil, err
		}

		//if its nil, just default it
		if params == nil {
			params = map[string]interface{}{}
		}

		if startVal.Pointer() == curPtr {

			//follow the end
			retVal := endValSlice[0].Elem()

			return curPtr, edgeLabel, true, fieldConf.Direction, params, &retVal, nil
		} else if endVal.Pointer() == curPtr {
			///follow the start
			retVal := startValSlice[0].Elem()

			return curPtr, edgeLabel, false, fieldConf.Direction, params, &retVal, nil
		} else {
			return 0, "", false, 0, nil, nil, errors.New("edge is invalid, doesn't point to parent vertex")
		}
	} else {
		return curPtr, edgeLabel, fieldConf.Direction == dsl.DirectionOutgoing, fieldConf.Direction, map[string]interface{}{}, relValue, nil
	}
}
