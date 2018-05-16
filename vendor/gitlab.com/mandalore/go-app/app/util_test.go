package app

import "testing"

func TestUtilAssertPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should panic")
		} else if r.(string) != "not equal" {
			t.Error("should panic with 'not equal'")
		}
	}()
	Util.Assert(false, "not equal")
}

func TestUtilAssertDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Error("should not panic")
		}
	}()
	Util.Assert(true, "not equal")
}

func TestUtilGenerateID(t *testing.T) {
	id1 := Util.GenerateUniqueID("name")
	id2 := Util.GenerateUniqueID("mane")

	if id1 == id2 {
		t.Error("should produce 2 unique ID's")
	}
}

func TestMustSerialize(t *testing.T) {
	user := struct{ Name string }{"john"}
	jsondata := Util.MustSerialize(user)
	if string(jsondata) != `{"Name":"john"}` {
		t.Error("failed to serialize")
	}
}

func TestMustSerializeFailure(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("should have panic")
		}
	}()

	Util.MustSerialize(make(chan bool))
}

func TestShouldSerialize(t *testing.T) {
	user := struct{ Name string }{"john"}
	jsondata := Util.ShouldSerialize(user)
	if string(jsondata) != `{"Name":"john"}` {
		t.Error("failed to serialize")
	}
}

func TestShouldSerializeFailure(t *testing.T) {
	jsondata := Util.ShouldSerialize(make(chan bool))
	if string(jsondata) != "SERIALIZATION_ERROR" {
		t.Error("should have returned a serialization error string")
	}
}
