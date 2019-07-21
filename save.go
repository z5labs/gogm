package gogm

import (
	"errors"
	"fmt"
	dsl "github.com/mindstand/go-cypherdsl"
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
	Type reflect.Type
}

func save(sess *dsl.Session, obj interface{}) error{
	return saveDepth(sess, obj, defaultSaveDepth)
}

func saveDepth(sess *dsl.Session, obj interface{}, depth int) error {
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

	err := parseStruct(nil, nil, &rootVal, 0, depth, &nodes, &relations)
	if err != nil{
		return err
	}

	return nil
}

func parseStruct(parent *reflect.Value, parentConf *structDecoratorConfig, current *reflect.Value, curDepth, maxDepth int, nodesPtr *map[string]map[string]nodeCreateConf, relationsPtr *map[string][]relCreateConf) error{
	if curDepth > maxDepth{
		return nil
	}

	if current == nil{
		return errors.New("current should never be nil")
	}

	var err error
	defer func() {
		if r := recover(); r != nil{
			err = fmt.Errorf("%v", r)
		}
	}()

	tString, err := getTypeName(current.Type())
	if err != nil{
		return err
	}

	actual, ok := mappedTypes.Get(tString)
	if !ok{
		return fmt.Errorf("struct config not found type (%s)", tString)
	}

	currentConf, ok := actual.(structDecoratorConfig)
	if !ok{
		return errors.New("unable to cast into struct decorator config")
	}

	//set this to the actual field name later
	isNewNode, id, err := setUuidIfNeeded(current, "UUID")
	if err != nil{
		return err
	}

	params, err := toCypherParamsMap(*current, currentConf)
	if err != nil{
		return err
	}

	nc := nodeCreateConf{
		Type: current.Type(),
		IsNew: isNewNode,
		Params: params,
	}

	//set the map
	if _, ok := (*nodesPtr)[currentConf.Label]; !ok{
		(*nodesPtr)[currentConf.Label] = map[string]nodeCreateConf{}
	}

	(*nodesPtr)[currentConf.Label][id] = nc

	for _, conf := range currentConf.Fields{
		if conf.Relationship == ""{
			continue
		}

		relField := current.FieldByName(conf.FieldName)

		//if its nil, just skip it
		if relField.IsNil(){
			continue
		}

		if conf.ManyRelationship{
			slLen := relField.Len()
			if slLen != 0{
				//iterate through map
				for i := 0; i < slLen; i++{
					newCurrent := relField.Index(i)

					//check that we're not using an edge
					if !newCurrent.Type().Implements(edgeType){
						if parent != nil && !parent.IsNil(){
							if parent.FieldByName("UUID").Interface() == newCurrent.FieldByName("UUID").Interface(){
								continue //skip if its the parent node
							}
						}
					}

					err := parseStruct(current, &currentConf, &newCurrent, curDepth + 1, maxDepth, nodesPtr, relationsPtr)
					if err != nil{
						return err
					}
				}
			}
		} else {
			err := parseStruct(current, &currentConf, &relField, curDepth + 1, maxDepth, nodesPtr, relationsPtr)
			if err != nil{
				return err
			}
		}
	}
	return nil
}

