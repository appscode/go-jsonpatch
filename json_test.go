package jsonpatch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshalNullableValue(t *testing.T) {
	p1 := PatchOperation{
		Op:    "replace",
		Path:  "/a1",
		Value: nil,
	}
	d1, err := json.Marshal(p1)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, `{"op":"replace", "path":"/a1","value":null}`, string(d1))

	p2 := PatchOperation{
		Op:    "replace",
		Path:  "/a2",
		Value: "v2",
	}
	d2, err := json.Marshal(p2)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, `{"op":"replace", "path":"/a2", "value":"v2"}`, string(d2))
}

func TestMarshalNonNullableValue(t *testing.T) {
	p1 := PatchOperation{
		Op:   "remove",
		Path: "/a1",
	}
	d1, err := json.Marshal(p1)
	if err != nil {
		t.Fatal(err)
	}
	assert.JSONEq(t, `{"op":"remove", "path":"/a1"}`, string(d1))
}
