package db

import (
	"reflect"
	"sort"
	"testing"
)

func TestStore_GetByIndex(t *testing.T) {
	store := newStore()
	_, _ = store.Set("string", "String", 0)
	_, _ = store.Set("list", []interface{}{"One", "Two", "Three"}, 0)
	_, _ = store.Set("list2", []interface{}{"One", "Two", "Three"}, 0)
	_, _ = store.Set("map", map[string]interface{}{"One": "42", "Two": ""}, 0)

	var tests = []struct {
		key           string
		index         interface{}
		expected      interface{}
		expectedError error
	}{
		{"", "invalid", nil, ErrKeyNotFound},
		{"string", "invalid", nil, ErrIndexAccess},

		{"list", "1", "Two", nil},
		{"list", 1, "Two", nil},
		{"list", "0", "One", nil},
		{"list", 0, "One", nil},
		{"list", "-1", "Three", nil},
		{"list", -1, "Three", nil},
		{"list", "-2", "Two", nil},
		{"list", -2, "Two", nil},

		{"list", "-5", nil, ErrIndexAccess},
		{"list", -5, nil, ErrIndexAccess},
		{"list", "", nil, ErrNonIntegerSubkey},
		{"list2", -42.55, nil, ErrIllegalIndexType},

		{"map", "One", "42", nil},

		{"map", "Three", nil, ErrIndexAccess},
		{"map", 1, nil, ErrIllegalIndexType},
	}
	for _, tt := range tests {
		result, err := store.GetAtIndex(tt.key, tt.index)
		if tt.expectedError != nil {
			if err == nil {
				t.Errorf("TestStore_GetByIndex for %v did not raise error %v", tt, tt.expectedError)
				continue
			}
			if err != tt.expectedError {
				t.Errorf("TestStore_GetByIndex for %v expected error %v, got %v", tt, tt.expectedError, err)
				continue
			}
		} else {
			if err != nil {
				t.Errorf("TestStore_GetByIndex for %v returned unexpected error %v", tt, err)
				continue
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("TestStore_GetByIndex %v expected value %v, got %v", tt, tt.expected, result)
			}
		}
	}
}

func TestStore_keys(t *testing.T) {
	strval := Value{STRING, "string", 0}
	listval := Value{LIST, []interface{}{"1", 2, nil}, 0}
	mapval := Value{MAP, map[string]interface{}{"1": 42, "2": "St"}, 0}

	s := newStore()
	s.store["str"] = &strval
	s.store["list"] = &listval
	s.store["map"] = &mapval

	result, err := s.Keys()
	if err != nil {
		t.Errorf("TestStore_keys got error %v", err)
	}
	expected := []string{"str", "list", "map"}
	sort.Strings(result)
	sort.Strings(expected)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("TestStore_keys expected %v, got %v", expected, result)
	}
}

func TestStore_remove(t *testing.T) {
	strval := Value{STRING, "string", 0}
	listval := Value{LIST, []interface{}{"1", 2, nil}, 0}
	mapval := Value{MAP, map[string]interface{}{"1": 42, "2": "St"}, 0}

	s := newStore()
	s.store["str"] = &strval
	s.store["str_to_keep"] = &strval
	s.store["list"] = &listval
	s.store["map"] = &mapval

	type test struct {
		key           string
		expectedError error
	}

	tests := []test{
		{"invalid", nil},
		{"str", nil},
		{"list", nil},
		{"map", nil},
	}

	for _, tt := range tests {
		err := s.Remove(tt.key)
		if tt.expectedError != nil {
			if err == nil {
				t.Errorf("TestStore_remove %v expected error %v, got none", tt.key, tt.expectedError)
				continue
			}
			if err != tt.expectedError {
				t.Errorf("TestStore_remove %v expected error %v, got %v", tt.key, tt.expectedError, err)
				continue
			}

		} else {
			if err != nil {
				t.Errorf("TestStore_remove %v expected no error, got %v", tt.key, err)
				continue
			}
		}
	}

	if l := len(s.store); l != 1 {
		t.Errorf("After Remove map is expected to store %v values, got %v", 1, l)
	}
	if !reflect.DeepEqual(&strval, s.store["str_to_keep"]) {
		t.Errorf("Unmodified entry expected to be %v, actually %v", &strval, s.store["str_to_keep"])
	}
}

func TestStore_set(t *testing.T) {
	s := newStore()

	type test struct {
		key           string
		value         interface{}
		expected      *Value
		expectedError error
	}

	tests := []test{
		{"string", "Something", &Value{Type: STRING, Data: "Something"}, nil},
		{"list", []interface{}{1, 2, "a"}, &Value{Type: LIST, Data: []interface{}{1, 2, "a"}}, nil},
		{"map", map[string]interface{}{"a": "str", "b": 1}, &Value{Type: MAP, Data: map[string]interface{}{"a": "str", "b": 1}}, nil},
		{"err", struct{}{}, nil, ErrInvalidValueType},
		{"nil", nil, nil, ErrInvalidValueType},
		{"wrongMap", map[int]int{1: 42}, nil, ErrInvalidValueType},
	}

	for _, tt := range tests {
		result, err := s.Set(tt.key, tt.value, 0)
		if tt.expectedError != nil {
			if err == nil {
				t.Errorf("TestStore_set %v expected error %v, got none", tt.key, tt.expectedError)
				continue
			}
			if err != tt.expectedError {
				t.Errorf("TestStore_set %v expected error %v, got %v", tt.key, tt.expectedError, err)
				continue
			}

		} else {
			if err != nil {
				t.Errorf("TestStore_set %v expected no error, got %v", tt.key, err)
				continue
			}
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("TestStore_set %v expected %v, got %v", tt.key, tt.expected, result)
			}
		}
	}

}

func TestStore_get(t *testing.T) {
	strval := Value{STRING, "string", 0}
	listval := Value{LIST, []interface{}{"1", 2, nil}, 0}
	mapval := Value{MAP, map[string]interface{}{"1": 42, "2": "St"}, 0}

	s := newStore()
	s.store["str"] = &strval
	s.store["list"] = &listval
	s.store["map"] = &mapval

	type test struct {
		key           string
		expected      *Value
		expectedError error
	}

	tests := []test{
		{"invalid", nil, ErrKeyNotFound},
		{"str", &strval, nil},
		{"list", &listval, nil},
		{"map", &mapval, nil},
	}

	for _, tt := range tests {
		result, err := s.Get(tt.key)
		if tt.expectedError != nil {
			if err == nil {
				t.Errorf("TestStore_get %v expected error %v, got none", tt.key, tt.expectedError)
				continue
			}
			if err != tt.expectedError {
				t.Errorf("TestStore_get %v expected error %v, got %v", tt.key, tt.expectedError, err)
				continue
			}

		} else {
			if err != nil {
				t.Errorf("TestStore_get %v expected no error, got %v", tt.key, err)
				continue
			}
			if !reflect.DeepEqual(tt.expected, result) {
				t.Errorf("TestStore_get %v expected %v, got %v", tt.key, tt.expected, result)
			}
		}
	}
}
