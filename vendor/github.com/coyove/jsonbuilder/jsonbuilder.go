package jsonbuilder

import (
	"encoding/json"
	"log"
	"reflect"
)

type JsonHelper struct {
	parents []*JsonHelper
	self    interface{}

	Objects      map[string]interface{}
	ObjectsArray []interface{}
}

func Object() *JsonHelper {
	j := &JsonHelper{}
	j.parents = make([]*JsonHelper, 0)
	j.Objects = make(map[string]interface{})
	j.ObjectsArray = nil
	return j
}

func From(s interface{}) *JsonHelper {
	j := &JsonHelper{}
	j.parents = make([]*JsonHelper, 0)

	buf, _ := json.Marshal(s)
	k := reflect.ValueOf(s).Kind()

	if k == reflect.Array {
		json.Unmarshal(buf, &j.ObjectsArray)
		j.Objects = nil
	} else if k == reflect.Struct {
		json.Unmarshal(buf, &j.Objects)
		j.ObjectsArray = nil
	}

	return j
}

func Array(values ...interface{}) *JsonHelper {
	j := &JsonHelper{}
	j.parents = make([]*JsonHelper, 0)
	j.Objects = nil
	j.ObjectsArray = make([]interface{}, 0)

	if len(values) > 0 {
		j = j.Push(values...)
	}

	return j
}

func (jo *JsonHelper) Get(name interface{}) interface{} {
	if jo.Objects != nil {
		return jo.Objects[name.(string)]
	} else {
		return jo.ObjectsArray[name.(int)]
	}
}

func (jo *JsonHelper) Set(name interface{}, values ...interface{}) *JsonHelper {
	if jo.Objects != nil {
		jo.ObjectsArray = nil

		var x []interface{}
		if len(values) > 1 {
			jo.Objects[name.(string)] = make([]interface{}, len(values))
			x = jo.Objects[name.(string)].([]interface{})
		}

		for i, v := range values {
			var value interface{}

			switch v.(type) {
			case *JsonHelper:
				if v.(*JsonHelper).Objects == nil {
					value = v.(*JsonHelper).ObjectsArray
				} else {
					value = v.(*JsonHelper).Objects
				}
			default:
				value = v
			}

			if len(values) == 1 {
				jo.Objects[name.(string)] = value
			} else {
				x[i] = value
			}
		}

	} else {
		jo.Objects = nil
		index := name.(int)

		var x []interface{}
		if len(values) > 1 {
			jo.ObjectsArray[index] = make([]interface{}, len(values))
			x = jo.ObjectsArray[index].([]interface{})
		}

		for i, v := range values {
			var value interface{}

			switch v.(type) {
			case *JsonHelper:
				if v.(*JsonHelper).Objects == nil {
					value = v.(*JsonHelper).ObjectsArray
				} else {
					value = v.(*JsonHelper).Objects
				}
			default:
				value = v
			}

			if len(values) == 1 {
				jo.ObjectsArray[index] = value
			} else {
				x[i] = value
			}
		}
	}

	return jo
}

func (jo *JsonHelper) Push(values ...interface{}) *JsonHelper {
	if jo.Objects == nil {
		for _, v := range values {
			switch v.(type) {
			case *JsonHelper:
				_v := v.(*JsonHelper)
				if _v.Objects == nil {
					jo.ObjectsArray = append(jo.ObjectsArray, v.(*JsonHelper).ObjectsArray)
				} else {
					jo.ObjectsArray = append(jo.ObjectsArray, v.(*JsonHelper).Objects)
				}
			default:
				jo.ObjectsArray = append(jo.ObjectsArray, v)
			}
		}
	}

	return jo
}

func (jo *JsonHelper) Delete(name interface{}) *JsonHelper {

	if jo.Objects != nil {
		delete(jo.Objects, name.(string))
	} else {
		index := name.(int)

		if index < len(jo.ObjectsArray) {
			foo := append(jo.ObjectsArray[:index], jo.ObjectsArray[index+1:]...)

			switch jo.self.(type) {
			case int:
				jo.parents[len(jo.parents)-1].ObjectsArray[jo.self.(int)] = foo
			case string:
				jo.parents[len(jo.parents)-1].Objects[jo.self.(string)] = foo
			}
		}

	}
	return jo
}

func (jo *JsonHelper) End() *JsonHelper {
	return jo.Leave()
}

func (jo *JsonHelper) Leave() *JsonHelper {
	if len(jo.parents) == 0 {
		log.Fatal("[JsonHelper] No more upstairs")
		return jo
	}

	lp := jo.parents[len(jo.parents)-1]
	np := jo.parents[:len(jo.parents)-1]

	return &JsonHelper{np, lp.self, lp.Objects, lp.ObjectsArray}
}

func (jo *JsonHelper) Begin(n interface{}) *JsonHelper {
	return jo.Enter(n)
}

func (jo *JsonHelper) Enter(n interface{}) *JsonHelper {
	if jo.Objects != nil {
		name := n.(string)
		if _, e := jo.Objects[name]; !e {
			jo.Objects[name] = make(map[string]interface{})
		}

		switch jo.Objects[name].(type) {
		case map[string]interface{}:
			return &JsonHelper{
				append(jo.parents, jo),
				n,
				jo.Objects[name].(map[string]interface{}),
				nil,
			}
		case []interface{}:
			return &JsonHelper{
				append(jo.parents, jo),
				n,
				nil,
				jo.Objects[name].([]interface{}),
			}
		}
	} else {
		index := n.(int)
		switch jo.ObjectsArray[index].(type) {
		case map[string]interface{}:
			return &JsonHelper{
				append(jo.parents, jo),
				n,
				jo.ObjectsArray[index].(map[string]interface{}),
				nil,
			}
		case []interface{}:
			return &JsonHelper{
				append(jo.parents, jo),
				n,
				nil,
				jo.ObjectsArray[index].([]interface{}),
			}
		}
	}

	return nil
}

func (jo *JsonHelper) Dive(name ...string) *JsonHelper {
	p := jo
	for _, v := range name {
		p = p.Enter(v)
	}
	return p
}

func (jo *JsonHelper) Marshal() string {
	if jo.ObjectsArray == nil {
		buf, _ := json.Marshal(jo.Objects)
		return string(buf)
	} else {
		buf, _ := json.Marshal(jo.ObjectsArray)
		return string(buf)
	}
}

func (jo *JsonHelper) MarshalPretty() string {
	if jo.ObjectsArray == nil {
		buf, _ := json.MarshalIndent(jo.Objects, "", "    ")
		return string(buf)
	} else {
		buf, _ := json.MarshalIndent(jo.ObjectsArray, "", "    ")
		return string(buf)
	}
}
