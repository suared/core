package test

import "testing"

func TestExpectedStrings(t *testing.T) {
	if !ContainsExpectedStrings("test1", "test2", "test2", "test1") {
		t.Errorf("first set should match")
	}
	if ContainsExpectedStrings("test1", "test2", "test2", "test11") {
		t.Errorf("second set should not match")
	}
	if ContainsExpectedStrings("test1", "test2", "test3", "test2", "test11") {
		t.Errorf("third set should not match")
	}
	if !ContainsExpectedStrings("test1", "test2", "test3", "test2", "test3", "test1") {
		t.Errorf("fourth set should match")
	}
	if ContainsExpectedStrings("test1", "test2", "test3", "test2", "test3", "test11") {
		t.Errorf("fifths set should not match")
	}
}
