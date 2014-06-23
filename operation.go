package jsonpatch

import (
	//"encoding/json"
	//"github.com/bitly/go-simplejson"
	"fmt"
	ptr "github.com/xeipuuv/gojsonpointer"
	"reflect"
	"strconv"
	"strings"
)

// Operation is a...
type OperationType string

// All available operations.
const (
	Remove  OperationType = "remove"
	Add     OperationType = "add"
	Replace OperationType = "replace"
	Move    OperationType = "move"
	Test    OperationType = "test"
	Copy    OperationType = "copy"
)

type PatchOperation struct {
	From  string        `json:"from,omitempty"`
	Op    OperationType `json:"op"`
	Path  string        `json:"path"`
	Value interface{}   `json:"value,omitempty"`
}

func lastObj(path string) (ptr.JsonPointer, string, error) {
	lastSep := strings.LastIndex(path, "/")
	parentPath := path[0:lastSep]
	lastToken := path[lastSep+1:] // Skip "/"
	parentPtr, err := ptr.NewJsonPointer(parentPath)
	return parentPtr, lastToken, err
}

func getValue(path string, doc interface{}) (*ptr.JsonPointer, reflect.Kind, interface{}, error) {
	ptr, err := ptr.NewJsonPointer(path)
	if err != nil {
		return nil, reflect.Invalid, nil, err
	}
	val, kind, err := ptr.Get(doc)
	if err != nil {
		return nil, reflect.Invalid, nil, err
	}
	return &ptr, kind, val, nil
}

func (self *PatchOperation) Apply(doc interface{}) error {
	doct, isMap := doc.(*map[string]interface{})
	var mod map[string]interface{}
	if isMap {
		mod = map[string]interface{}{"__": *doct}
	}
	doca, isArr := doc.(*[]interface{})
	if isArr {
		mod = map[string]interface{}{"__": *doca}
	}
	path := "/__" + self.Path
	from := "/__" + self.From
	var err error
	switch self.Op {
	case Add:
		err = add(path, self.Value, mod)
	case Copy:
		err = copyOp(path, from, mod)
	case Move:
		err = move(path, from, self.Value, mod)
	case Remove:
		err = remove(path, mod)
	case Replace:
		err = replace(path, self.Value, mod)
	case Test:
		err = test(path, self.Value, mod)
	default:
		err = fmt.Errorf("Unknown operation type: %s", self.Op)
	}
	if isMap {
		*doct = mod["__"].(map[string]interface{})
	} else if isArr {
		*doca = mod["__"].([]interface{})
	}
	return err
}

func add(path string, value interface{}, doc interface{}) error {
	parentPtr, lastToken, err := lastObj(path)
	if err != nil {
		return err
	}
	parentValue, kind, err := parentPtr.Get(doc)
	if err != nil {
		return err
	}
	if "/"+lastToken == path {
		// This is a path to the root object
		kind = reflect.ValueOf(doc).Kind()
	}
	switch kind {
	case reflect.Map:
		m := parentValue.(map[string]interface{})
		m[lastToken] = value
	case reflect.Slice:
		existing := parentValue.([]interface{})
		var index = len(existing)
		if lastToken != "-" {
			var err error
			if index, err = strconv.Atoi(lastToken); err != nil {
				return err
			}
		}
		existing = append(existing, 0)
		copy(existing[index+1:], existing[index:])
		existing[index] = value
		// Need to replace
		parentPtr.Set(doc, existing)
	default:
		return fmt.Errorf("Cannot add to document type: %v\n", kind)
	}
	return nil
}

func copyOp(path string, from string, doc interface{}) error {
	_, _, value, err := getValue(from, doc)
	if err != nil {
		return err
	}
	return add(path, value, doc)
}

func move(path string, from string, value interface{}, doc interface{}) error {
	//if strings.HasPrefix(path, from) {
	//	return fmt.Errorf("Cannot move values into its own children")
	//}
	_, _, value, err := getValue(from, doc)
	if err != nil {
		return err
	}
	if err = remove(from, doc); err != nil {
		return err
	}
	return add(path, value, doc)
}

func remove(path string, doc interface{}) error {
	parentPtr, lastToken, err := lastObj(path)
	if err != nil {
		return err
	}
	parentVal, parentKind, err := parentPtr.Get(doc)
	if err != nil {
		return err
	}
	if "/"+lastToken == path {
		parentKind = reflect.ValueOf(doc).Kind()
	}
	switch parentKind {
	case reflect.Map:
		m := parentVal.(map[string]interface{})
		delete(m, lastToken)
	case reflect.Slice:
		index, err := strconv.Atoi(lastToken)
		if err != nil {
			return err
		}
		existing := parentVal.([]interface{})
		existing[index] = nil
		copy(existing[index:], existing[index+1:])
		existing = existing[0 : len(existing)-1]

		// Need to replace
		_, err = parentPtr.Set(doc, existing)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unable to remove from kind %s", parentKind)
	}
	return nil
}

func replace(path string, value interface{}, doc interface{}) error {
	parentPtr, lastToken, err := lastObj(path)
	if err != nil {
		return err
	}
	parentVal, parentKind, err := parentPtr.Get(doc)
	if err != nil {
		return err
	}
	if "/"+lastToken == path {
		parentKind = reflect.ValueOf(doc).Kind()
	}
	switch parentKind {
	case reflect.Map:
		m := parentVal.(map[string]interface{})
		m[lastToken] = value
	case reflect.Slice:
		s := parentVal.([]interface{})
		index, err := strconv.Atoi(lastToken)
		if err != nil {
			return err
		}
		s[index] = value
	default:
		return fmt.Errorf("Unable to replace item of kind %s", parentKind)
	}
	return nil
}

func test(path string, value interface{}, doc interface{}) error {
	_, _, pathValue, err := getValue(path, doc)
	if err != nil {
		return err
	}
	if value != nil && !reflect.DeepEqual(pathValue, value) {
		return fmt.Errorf("Tested path %s: %v != %v", path, pathValue, value)
	}
	return nil
}