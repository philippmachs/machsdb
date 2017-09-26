package db

import (
	"reflect"
	"strconv"
	"time"
)

type store struct {
	store map[string]*Value
}

func newStore() *store {
	return &store{
		map[string]*Value{},
	}
}

func (s *store) Set(key string, value interface{}, expires time.Duration) (result *Value, err error) {
	if expires < 0 {
		return nil, ErrInvalidTTL
	}

	targetType := STRING
	switch reflect.ValueOf(value).Kind() {
	case reflect.String:
		targetType = STRING
	case reflect.Slice:
		targetType = LIST
	case reflect.Map:
		keyType := reflect.TypeOf(value).Key().Kind()
		if keyType != reflect.String {
			return nil, ErrInvalidValueType
		}
		targetType = MAP
	default:
		return nil, ErrInvalidValueType
	}

	var expireNs int64
	if expires > 0 {
		expireNs = (time.Now().Add(expires)).UnixNano()
	}
	newData := &Value{Type: targetType, Data: value, Expires: expireNs}
	s.store[key] = newData
	return newData, nil
}

func (s *store) Get(key string) (result *Value, err error) {
	result, ok := s.store[key]
	if !ok {
		return nil, ErrKeyNotFound
	}
	return result, nil
}

func (s *store) Remove(key string) (err error) {
	delete(s.store, key)
	return nil
}

func (s *store) Keys() (result []string, err error) {
	now := time.Now().UnixNano()
	result = make([]string, len(s.store))
	i := 0
	for k := range s.store {
		if s.store[k].Expires == 0 || s.store[k].Expires > now {
			result[i] = k
			i++
		}
	}
	result = result[:i]
	return
}

func (s *store) GetAtIndex(key string, index interface{}) (result interface{}, err error) {
	value, ok := s.store[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if value.Expires != 0 && value.Expires < time.Now().UnixNano() {
		return nil, ErrKeyNotFound
	}

	return value.getItemAtIndex(index)
}

func (v *Value) getItemAtIndex(index interface{}) (result interface{}, err error) {
	switch v.Type {
	case STRING:
		err = ErrIndexAccess
	case LIST:
		i := 0
		i, err = indexToInt(index)
		if err != nil {
			break
		}

		li, ok := v.Data.([]interface{})
		if !ok {
			err = ErrConversionError
			break
		}
		if i < 0 {
			i = len(li) + i
		}
		if i >= len(li) || i < 0 {
			err = ErrIndexAccess
			break
		}

		result = li[i]

	case MAP:
		i, ok := index.(string)
		if !ok {
			err = ErrIllegalIndexType
			break
		}
		in, ok := v.Data.(map[string]interface{})
		if !ok {
			err = ErrConversionError
			break
		}
		result, ok = in[i]
		if !ok {
			err = ErrIndexAccess
			break
		}
	default:
		err = ErrInvalidValueType
	}
	return

}

func indexToInt(index interface{}) (idx int, err error) {
	idx, ok := index.(int)
	if !ok {
		strIndex, ok := index.(string)
		if !ok {
			err = ErrIllegalIndexType
			return
		}
		idx, err = strconv.Atoi(strIndex)
		if err != nil {
			err = ErrNonIntegerSubkey
			return
		}
	}
	return
}
