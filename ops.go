package jee

import (
	"encoding/json"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type OpMap struct {
	Float   map[string]func(float64, float64) interface{}
	String  map[string]func(string, string) interface{}
	Bool    map[string]func(bool, bool) interface{}
	Nil     map[string]func(interface{}, interface{}) interface{}
	Nullary map[string]func() (interface{}, error)
	Unary   map[string]func(interface{}) (interface{}, error)
	Binary  map[string]func(interface{}, interface{}) (interface{}, error)
}

func DefaultOpMap() *OpMap {
	return defaultOpMap.Clone()
}

func (o *OpMap) Clone() *OpMap {
	clone := &OpMap{
		Float:   map[string]func(float64, float64) interface{}{},
		String:  map[string]func(string, string) interface{}{},
		Bool:    map[string]func(bool, bool) interface{}{},
		Nil:     map[string]func(interface{}, interface{}) interface{}{},
		Nullary: map[string]func() (interface{}, error){},
		Unary:   map[string]func(interface{}) (interface{}, error){},
		Binary:  map[string]func(interface{}, interface{}) (interface{}, error){},
	}

	for k, v := range o.Float {
		clone.Float[k] = v
	}

	for k, v := range o.String {
		clone.String[k] = v
	}

	for k, v := range o.Bool {
		clone.Bool[k] = v
	}

	for k, v := range o.Nil {
		clone.Nil[k] = v
	}

	for k, v := range o.Nullary {
		clone.Nullary[k] = v
	}

	for k, v := range o.Unary {
		clone.Unary[k] = v
	}

	for k, v := range o.Binary {
		clone.Binary[k] = v
	}

	return clone
}

func (o *OpMap) AddUnary(key string, f func(interface{}) (interface{}, error)) {
	o.Unary[key] = f
}

func (o *OpMap) AddBinary(key string, f func(interface{}, interface{}) (interface{}, error)) {
	o.Binary[key] = f
}

var (
	defaultOpMap = &OpMap{
		Float:   opFuncsFloat,
		String:  opFuncsString,
		Bool:    opFuncsBool,
		Nil:     opFuncsNil,
		Nullary: nullaryFuncs,
		Unary:   unaryFuncs,
		Binary:  binaryFuncs,
	}
)

var opFuncsFloat = map[string]func(float64, float64) interface{}{
	"+": func(a float64, b float64) interface{} {
		return a + b
	},
	"-": func(a float64, b float64) interface{} {
		return a - b
	},
	"*": func(a float64, b float64) interface{} {
		return a * b
	},
	"/": func(a float64, b float64) interface{} {
		return a / b
	},
	"==": func(a float64, b float64) interface{} {
		return a == b
	},
	">=": func(a float64, b float64) interface{} {
		return a >= b
	},
	">": func(a float64, b float64) interface{} {
		return a > b
	},
	"<": func(a float64, b float64) interface{} {
		return a < b
	},
	"<=": func(a float64, b float64) interface{} {
		return a <= b
	},
	"!=": func(a float64, b float64) interface{} {
		return a != b
	},
}

var opFuncsString = map[string]func(string, string) interface{}{
	"+": func(a string, b string) interface{} {
		return a + b
	},
	"==": func(a string, b string) interface{} {
		return a == b
	},
	"!=": func(a string, b string) interface{} {
		return a != b
	},
}

var opFuncsBool = map[string]func(bool, bool) interface{}{
	"&&": func(a bool, b bool) interface{} {
		return a && b
	},
	"||": func(a bool, b bool) interface{} {
		return a || b
	},
	"==": func(a bool, b bool) interface{} {
		return a == b
	},
	"!=": func(a bool, b bool) interface{} {
		return a != b
	},
}

var opFuncsNil = map[string]func(interface{}, interface{}) interface{}{
	"==": func(a interface{}, b interface{}) interface{} {
		if a == nil && b == nil {
			return true
		}

		// comparing objects is a horrible condition and should be avoided
		return reflect.DeepEqual(a, b)
	},
	"!=": func(a interface{}, b interface{}) interface{} {
		return a != b
	},
}

var nullaryFuncs = map[string]func() (interface{}, error){
	"$now": func() (interface{}, error) {
		return float64(time.Now().UnixNano() / 1000 / 1000), nil
	},
}

var unaryFuncs = map[string]func(interface{}) (interface{}, error){
	"$sum": func(val interface{}) (interface{}, error) {
		valsArray, ok := val.([]interface{})
		if !ok {
			return nil, nil
		}
		sum := 0.0
		for _, i := range valsArray {
			sum += i.(float64)
		}
		return sum, nil
	},
	"$min": func(val interface{}) (interface{}, error) {
		valsArray, ok := val.([]interface{})
		if !ok {
			return nil, nil
		}

		min := valsArray[0].(float64)
		for i := 1; i < len(valsArray); i++ {
			min = math.Min(min, valsArray[i].(float64))
		}
		return min, nil
	},
	"$max": func(val interface{}) (interface{}, error) {
		valsArray, ok := val.([]interface{})
		if !ok {
			return nil, nil
		}

		max := valsArray[0].(float64)
		for i := 1; i < len(valsArray); i++ {
			max = math.Max(max, valsArray[i].(float64))
		}
		return max, nil
	},
	"$len": func(val interface{}) (interface{}, error) {
		valsArray, ok := val.([]interface{})
		if !ok {
			return nil, nil
		}

		return float64(len(valsArray)), nil
	},
	"$sqrt": func(val interface{}) (interface{}, error) {
		f, ok := val.(float64)
		if !ok || f < 0 {
			return nil, nil
		}

		return math.Sqrt(f), nil
	},
	"$abs": func(val interface{}) (interface{}, error) {
		f, ok := val.(float64)
		if !ok {
			return nil, nil
		}
		return math.Abs(f), nil
	},
	"$floor": func(val interface{}) (interface{}, error) {
		f, ok := val.(float64)
		if !ok {
			return nil, nil
		}

		return math.Floor(f), nil
	},
	"$keys": func(val interface{}) (interface{}, error) {
		var keyList []interface{}
		m, ok := val.(map[string]interface{})
		if !ok {
			return nil, nil
		}

		for k, _ := range m {
			keyList = append(keyList, k)
		}

		return keyList, nil
	},
	"$str": func(val interface{}) (interface{}, error) {
		switch v := val.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', -1, 64), nil
		case bool:
			if v {
				return "true", nil
			}
			return "false", nil
		case string:
			return v, nil
		case nil:
			return "null", nil
		case map[string]interface{}, []interface{}:
			b, err := json.Marshal(v)
			return string(b), err
		}
		return "", nil
	},
	"$num": func(val interface{}) (interface{}, error) {
		switch v := val.(type) {
		case float64:
			return v, nil
		case string:
			return strconv.ParseFloat(v, 64)
		case bool:
			if v {
				return 1, nil
			}
		}
		return 0.0, nil
	},
	"$~bool": func(val interface{}) (interface{}, error) {
		switch v := val.(type) {
		case []interface{}:
			if len(v) > 0 {
				return true, nil
			}
		case map[string]interface{}:
			return true, nil
		case float64:
			if math.IsNaN(v) {
				return false, nil
			}

			if v > 0 {
				return true, nil
			}
		case string:
			if len(v) > 0 {
				return true, nil
			}
		case bool:
			return v, nil
		}
		return false, nil
	},
	"$bool": func(val interface{}) (interface{}, error) {
		switch v := val.(type) {
		case string:
			return strconv.ParseBool(v)
		case bool:
			return v, nil
		}

		return nil, nil
	},
}

var binaryFuncs = map[string]func(interface{}, interface{}) (interface{}, error){
	"$parseTime": func(a interface{}, b interface{}) (interface{}, error) {
		layout, ok := a.(string)
		if !ok {
			return nil, nil
		}
		value, ok := b.(string)
		if !ok {
			return nil, nil
		}
		t, err := time.Parse(layout, value)
		if err != nil {
			return nil, err
		}
		return float64(t.UnixNano() / 1000 / 1000), nil
	},
	"$fmtTime": func(a interface{}, b interface{}) (interface{}, error) {
		layout, ok := a.(string)
		if !ok {
			return nil, nil
		}

		t, ok := b.(float64)
		if !ok {
			return nil, nil
		}

		return time.Unix(0, int64(time.Duration(t)*time.Millisecond)).Format(layout), nil
	},
	"$pow": func(a interface{}, b interface{}) (interface{}, error) {
		fa, ok := a.(float64)
		if !ok {
			return nil, nil
		}
		fb, ok := b.(float64)
		if !ok {
			return nil, nil
		}

		return math.Pow(fa, fb), nil
	},
	"$exists": func(a interface{}, b interface{}) (interface{}, error) {
		sb, ok := b.(string)
		if !ok {
			return nil, nil
		}

		ma, ok := a.(map[string]interface{})
		if !ok {
			return nil, nil
		}

		_, ok = ma[sb]
		if ok {
			return true, nil
		}
		return false, nil
	},
	"$contains": func(a interface{}, b interface{}) (interface{}, error) {
		sa, ok := a.(string)
		if !ok {
			return nil, nil
		}

		sb, ok := b.(string)
		if !ok {
			return nil, nil
		}
		return strings.Contains(sa, sb), nil
	},
	"$regex": func(a interface{}, b interface{}) (interface{}, error) {
		sa, ok := a.(string)
		if !ok {
			return nil, nil
		}

		sb, ok := b.(string)
		if !ok {
			return nil, nil
		}

		return regexp.MatchString(sb, sa)
	},
	"$has": func(a interface{}, b interface{}) (interface{}, error) {
		s, ok := a.([]interface{})
		if !ok {
			return nil, nil
		}

		for _, e := range s {
			switch c := e.(type) {
			case string:
				if c == b.(string) {
					return true, nil
				}
			case float64:
				if c == b.(float64) {
					return true, nil
				}
			case bool:
				if c == b.(bool) {
					return true, nil
				}
			default:
				if a == nil && b == nil {
					return true, nil
				}
			}
		}
		return false, nil
	},
}
