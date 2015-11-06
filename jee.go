package jee

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"unicode"
)

const (
	ZERO = iota
	CONST
	OP
	FUNC
	KEY
	K_START
	K_END
	Q_START
	Q_END
	SPACE
	NEXT
	D_STR
	S_STR
	ESC
	RESERVED
	EQ
)

var Ident = map[rune]int{
	'$':  FUNC,
	'.':  KEY,
	'+':  OP,
	'-':  OP,
	'/':  OP,
	'*':  OP,
	'!':  OP,
	'=':  OP,
	'>':  OP,
	'<':  OP,
	'&':  OP,
	'|':  OP,
	'(':  Q_START,
	')':  Q_END,
	'[':  K_START,
	']':  K_END,
	'"':  D_STR,
	'\'': S_STR,
	'\\': ESC,
	',':  NEXT,
}

var IdentStr = map[int]string{
	FUNC:     "FUNC",
	OP:       "OP",
	KEY:      "KEY",
	CONST:    "CONST",
	Q_START:  "Q_START",
	Q_END:    "Q_END",
	K_START:  "K_START",
	K_END:    "K_END",
	NEXT:     "NEXT",
	D_STR:    "D_STR",
	S_STR:    "S_STR",
	RESERVED: "RES",
	EQ:       "EQ",
}

type BMsg interface{}

type Token struct {
	Type  int
	Value string
}

type TokenTree struct {
	Type   int
	Value  interface{}
	Tokens []*TokenTree
	Parent *TokenTree
}

var tokenPopMap = map[int]func(rune, string) bool{
	D_STR: func(r rune, c string) bool {
		switch Ident[r] {
		case D_STR:
			return true
		}
		return false
	},
	S_STR: func(r rune, c string) bool {
		switch Ident[r] {
		case S_STR:
			return true
		}
		return false
	},
	KEY: func(r rune, c string) bool {
		switch getIdent(r) {
		case Q_START, Q_END, K_START, K_END, OP, FUNC, NEXT, KEY, D_STR, S_STR:
			return true
		}
		return false
	},
	OP: func(r rune, c string) bool {
		switch c {
		case "*", "+", "-", "/":
			return true
		}

		if len(c) >= 2 {
			return true
		}
		switch getIdent(r) {
		case Q_START, Q_END, K_START, K_END, FUNC, CONST, KEY, D_STR, S_STR, RESERVED:
			return true
		}
		return false
	},
	CONST: func(r rune, c string) bool {
		switch getIdent(r) {
		case Q_START, Q_END, K_START, K_END, OP, FUNC, NEXT, D_STR, S_STR, RESERVED:
			return true
		}
		return false
	},
	FUNC: func(r rune, c string) bool {
		switch getIdent(r) {
		case Q_START, Q_END, K_START, K_END, OP, FUNC, NEXT, KEY, D_STR, S_STR:
			return true
		}
		return false
	},
	RESERVED: func(r rune, c string) bool {
		switch getIdent(r) {
		case Q_START, Q_END, K_START, K_END, OP, FUNC, NEXT, KEY, D_STR, S_STR:
			return true
		}
		return false
	},
}

func getIdent(r rune) int {
	i, ok := Ident[r]
	switch {
	case ok:
		return i
	case unicode.IsNumber(r):
		return CONST
	case unicode.IsLetter(r) || unicode.IsPunct(r) || unicode.IsSymbol(r):
		return RESERVED
	case unicode.IsSpace(r):
		return SPACE
	}
	return ZERO
}

func emitToken(t []*Token, state int, value string) ([]*Token, string) {
	t = append(t, &Token{
		Type:  state,
		Value: value,
	})
	return t, ""
}

// probably should just use bufio.Scanner...
func Lexer(input string) ([]*Token, error) {
	var tokens []*Token
	var currWord string
	var state int
	var poppedStr bool
	var escaped bool

	for _, r := range input {

		// if we have a space and we aren't in a string
		if getIdent(r) == SPACE && state != D_STR && state != S_STR {
			continue
		}

		// if we have an escape char and we are in a string
		if getIdent(r) == ESC && (state == D_STR || state == S_STR) {
			escaped = true
			continue
		}

		if getIdent(r) == ZERO && state != D_STR && state != S_STR {
			return nil, errors.New(fmt.Sprintf("unexpected token: %s", string(r)))
		}

		switch state {
		case OP, FUNC, CONST, KEY, RESERVED:
			if tokenPopMap[state](r, currWord) {
				tokens, currWord = emitToken(tokens, state, currWord)
			}
		case D_STR, S_STR:
			if escaped {
				escaped = false
				break
			}
			if tokenPopMap[state](r, currWord) {
				currWord += string(r)
				tokens, currWord = emitToken(tokens, state, currWord)
				poppedStr = true
			} else {
				poppedStr = false
			}
		case Q_START, Q_END, K_START, K_END, NEXT:
			tokens, currWord = emitToken(tokens, state, currWord)
		}

		if !poppedStr {
			if len(currWord) == 0 {
				currWord = string(r)
				state = getIdent(r)
			} else {
				currWord += string(r)
			}
		} else {
			poppedStr = false
			state = ZERO
		}
	}

	if len(currWord) > 0 {
		tokens, _ = emitToken(tokens, state, currWord)
	}

	return tokens, nil
}

func buildTree(tokens []*Token) (*TokenTree, error) {
	var state int
	tree := &TokenTree{}
	var inKey bool // TODO: this needs to go
	var nested int
	var knested int

	// first pass:
	// take care of quantities, funcs, keys.
	for _, t := range tokens {

		item := &TokenTree{
			Value:  t.Value,
			Type:   t.Type,
			Parent: tree,
		}

		// this item should probably be in Lexer
		// convert value to float64 if number
		if item.Type == CONST {
			f, err := strconv.ParseFloat(t.Value, 64)
			if err != nil {

			} else {
				item.Value = f
			}
		}

		// this item should probably be in Lexer
		// get rid of quotes around our strings
		if item.Type == D_STR || item.Type == S_STR {
			item.Value = item.Value.(string)[1 : len(item.Value.(string))-1]
		}

		// this item should probably be in Lexer
		// create bool type
		if item.Type == RESERVED {
			switch item.Value {
			case "true", "false":
				f, err := strconv.ParseBool(t.Value)
				if err != nil {

				} else {
					item.Value = f
				}
			case "null":
				item.Value = nil
			default:
				return nil, errors.New(fmt.Sprintf("unexpected token: %s", item.Value))
			}
		}

		//
		if item.Type == K_START {
			item.Value = nil
		}

		// remove '.' from key name
		if item.Type == KEY {
			item.Value = item.Value.(string)[1:]
		}

		switch t.Type {
		case FUNC, CONST, RESERVED, D_STR, S_STR, NEXT:
			if inKey {
				for tree.Parent != nil && tree.Type == KEY && (tree.Type == KEY || tree.Type == K_START) {
					tree = tree.Parent
				}
				inKey = false
			}
			tree.Tokens = append(tree.Tokens, item)
		case OP:
			if inKey {
				for tree.Parent != nil && tree.Type == KEY && (tree.Type == KEY || tree.Type == K_START) {
					tree = tree.Parent
				}
				inKey = false
			}
			tree.Tokens = append(tree.Tokens, item)
		case KEY:
			tree.Tokens = append(tree.Tokens, item)
			if !inKey {
				tree = item
			}
			inKey = true
		case Q_START:
			nested++
			if state == FUNC {
				// we are a function
				tree = tree.Tokens[len(tree.Tokens)-1]
			} else {
				// we are a quantity
				tree.Tokens = append(tree.Tokens, item)
				tree = item
			}
		case K_START:
			if tree.Type != KEY && tree.Type != K_START {
				return nil, errors.New("unexpected [")
			}

			knested++
			tree.Tokens = append(tree.Tokens, item)
			tree = item

		case K_END:
			knested--
			// ???????
			if tree.Parent != nil {
				tree = tree.Parent
			} else {
				return nil, errors.New("unbalanced () or []")
			}
			inKey = true

		case Q_END:
			nested--
			if inKey {
				for tree.Parent != nil && (tree.Type == KEY || tree.Type == K_START) {
					tree = tree.Parent
				}
				inKey = false
			}

			if tree.Parent != nil {
				tree = tree.Parent
			} else {
				return nil, errors.New("unbalanced () or []")
			}
		}

		switch state {
		default:
		}

		state = t.Type
	}

	if nested != 0 || knested != 0 {
		return nil, errors.New("unbalanced () or []")
	}

	for tree.Parent != nil {
		tree = tree.Parent
	}

	return tree, nil
}

func not(tree *TokenTree) *TokenTree {
	var negate *TokenTree
	var newTokens []*TokenTree

	for _, t := range tree.Tokens {

		if t.Type == OP && t.Value == "!" {
			negate = t
			newTokens = append(newTokens, t)
			continue
		}

		if negate != nil {
			negate.Tokens = append(negate.Tokens, t)
			negate = nil
		} else {
			t.Parent = negate
			newTokens = append(newTokens, t)
		}

		if len(t.Tokens) > 0 {
			t = not(t)
		}
	}

	tree.Tokens = newTokens
	return tree
}

func split(tree *TokenTree, TokenType int, Values []string) *TokenTree {
	var nextTokens []*TokenTree

	popTokens := make([]*TokenTree, len(tree.Tokens))

	if len(tree.Tokens) > 0 {
		for i, t := range tree.Tokens {
			tree.Tokens[i] = split(t, TokenType, Values)
		}
	}

	copy(popTokens, tree.Tokens)

	for len(popTokens) > 2 {
		prev := popTokens[0]
		curr := popTokens[1]
		next := popTokens[2]

		if curr.Type == TokenType && inStringSlice(Values, curr.Value.(string)) {
			prev.Parent = curr
			next.Parent = curr
			curr.Tokens = append(curr.Tokens, prev, next)
			popTokens = popTokens[2:]
			popTokens[0] = curr
		} else {
			nextTokens = append(nextTokens, popTokens[0])
			popTokens = popTokens[1:]
		}
	}

	nextTokens = append(nextTokens, popTokens...)

	tree.Tokens = nextTokens

	return tree
}

func inStringSlice(a []string, b string) bool {
	for _, s := range a {
		if s == b {
			return true
		}
	}
	return false
}

func negative(tree *TokenTree) *TokenTree {
	var negate *TokenTree
	var newTokens []*TokenTree
	var state int

	for _, t := range tree.Tokens {

		// not entirely sure this case is correct
		if t.Type == OP && t.Value == "-" && state != CONST && state != KEY && state != Q_START {
			negate = t
			newTokens = append(newTokens, t)
			continue
		}

		if negate != nil {
			negate.Tokens = append(negate.Tokens, t)
			negate = nil
		} else {
			t.Parent = negate
			newTokens = append(newTokens, t)
		}

		state = t.Type

		if len(t.Tokens) > 0 {
			t = negative(t)
		}
	}

	tree.Tokens = newTokens
	return tree
}

func Parser(tokens []*Token) (*TokenTree, error) {

	tree, err := buildTree(tokens)
	if err != nil {
		return nil, err
	}

	tree = not(tree)
	tree = negative(tree)

	tree = split(tree, OP, []string{"&&", "||"})
	tree = split(tree, OP, []string{"*", "/"})
	tree = split(tree, OP, []string{"+", "-"})
	tree = split(tree, OP, []string{"==", ">=", ">", "<", "<=", "!="})

	return tree, nil
}

func getKeyValues(t *TokenTree, input BMsg) (interface{}, error) {
	s, ok := t.Value.(string)

	if ok && len(s) > 0 {
		inputMap, ok := input.(map[string]interface{})
		if !ok {
			return nil, errors.New("could not assert to map")
		}

		input = inputMap[s]
	}

	var output []interface{}
	output = append(output, input)
	var accessed bool // this needs to be figured out!

	for _, sub := range t.Tokens {
		switch sub.Type {
		case K_START:
			switch c := sub.Value.(type) {
			case string:
				for j, _ := range output {
					outputMap, ok := output[j].(map[string]interface{})
					if !ok {
						return nil, errors.New("could not assert to map")
					}

					output[j] = outputMap[c]
				}
			case float64:
				for j, _ := range output {
					outputSlice, ok := output[j].([]interface{})
					if !ok {
						return nil, errors.New("could not assert to slice")
					}
					sliceIndex := int(c)
					if c < 0 || sliceIndex >= len(outputSlice) {
						output[j] = nil
					} else {
						output[j] = outputSlice[sliceIndex]
					}
				}
			default:
				accessed = true
				var newOutput []interface{}
				for j, _ := range output {
					arr, ok := output[j].([]interface{})
					if !ok {
						return nil, errors.New("could not assert to slice")
					}
					for _, e := range arr {
						newOutput = append(newOutput, e)
					}
				}
				output = newOutput
			}
		case KEY:
			for j, _ := range output {
				outputMap, ok := output[j].(map[string]interface{})
				if !ok {
					errors.New("could not assert to map")
				}

				subValue, ok := sub.Value.(string)
				if !ok {
					errors.New("invalid key for map")
				}
				output[j] = outputMap[subValue]
			}
		}
	}

	if len(output) == 1 && !accessed {
		return output[0], nil
	}

	return output, nil
}

func Eval(t *TokenTree, msg BMsg) (interface{}, error) {
	return EvalCustom(nil, t, msg)
}

func EvalCustom(opMap *OpMap, t *TokenTree, msg BMsg) (interface{}, error) {
	if opMap == nil {
		opMap = defaultOpMap
	}

	var tokenVal string

	switch t.Type {
	case OP, KEY, FUNC:
		_, ok := t.Value.(string)
		if !ok {
			return nil, errors.New(fmt.Sprintf("bad operation, key, or function: %s", t.Value))
		}
		tokenVal = t.Value.(string)
	}

	switch t.Type {
	case OP:
		if len(t.Tokens) == 1 {
			switch tokenVal {
			case "-":
				r, err := EvalCustom(opMap, t.Tokens[0], msg)
				if err != nil {
					return nil, err
				}

				f, ok := r.(float64)
				if !ok {
					return nil, errors.New("cannot use - operator on non-number type")
				}

				return -1 * f, nil

			case "!":
				r, err := EvalCustom(opMap, t.Tokens[0], msg)
				if err != nil {
					return nil, err
				}

				b, ok := r.(bool)
				if !ok {
					return nil, errors.New("cannot use ! operator on non-bool type")
				}

				return !b, nil
			}
			break
		}
		if len(t.Tokens) == 2 {
			a, err := EvalCustom(opMap, t.Tokens[0], msg)
			if err != nil {
				return nil, err
			}

			b, err := EvalCustom(opMap, t.Tokens[1], msg)
			if err != nil {
				return nil, err
			}
			// need to do comparisons for falsy-null || X
			// as well as != and ==

			switch ta := a.(type) {
			case float64:
				bf, ok := b.(float64)
				if !ok && tokenVal == "!=" {
					return true, nil
				} else if !ok && tokenVal == "==" {
					return false, nil
				} else if !ok {
					return nil, errors.New(fmt.Sprintf("cannot compare types: %s, %s", reflect.TypeOf(a), reflect.TypeOf(b)))
				}

				_, ok = opMap.Float[tokenVal]
				if !ok {
					return nil, errors.New(fmt.Sprintf("invalid operator for type: %s, %s", tokenVal, reflect.TypeOf(a)))
				}

				return opMap.Float[tokenVal](ta, bf), nil
			case string:
				bs, ok := b.(string)
				if !ok && tokenVal == "!=" {
					return true, nil
				} else if !ok && tokenVal == "==" {
					return false, nil
				} else if !ok {
					return nil, errors.New(fmt.Sprintf("cannot compare types: %s, %s", reflect.TypeOf(a), reflect.TypeOf(b)))
				}

				_, ok = opMap.String[tokenVal]
				if !ok {
					return nil, errors.New(fmt.Sprintf("invalid operator for type: %s, %s", tokenVal, reflect.TypeOf(a)))
				}

				return opMap.String[tokenVal](ta, bs), nil
			case bool:
				bb, ok := b.(bool)
				if !ok && tokenVal == "!=" {
					return true, nil
				} else if !ok && tokenVal == "==" {
					return false, nil
				} else if !ok {
					return nil, errors.New(fmt.Sprintf("cannot compare types: %s, %s", reflect.TypeOf(a), reflect.TypeOf(b)))
				}

				_, ok = opMap.Bool[tokenVal]
				if !ok {
					return nil, errors.New(fmt.Sprintf("invalid operator for type: %s, %s", tokenVal, reflect.TypeOf(a)))
				}

				return opMap.Bool[tokenVal](ta, bb), nil
			default:
				_, ok := opMap.Nil[tokenVal]
				if !ok {
					return nil, errors.New(fmt.Sprintf("invalid operator for type: %s, %s", tokenVal, reflect.TypeOf(a)))
				}

				return opMap.Nil[tokenVal](a, b), nil
			}
		}
	case S_STR, D_STR, CONST, RESERVED:
		return t.Value, nil
	case KEY:
		input := msg

		for _, sub := range t.Tokens {
			if len(sub.Tokens) > 0 {
				key, err := EvalCustom(opMap, sub.Tokens[0], input)
				if err != nil {
					return nil, err
				}

				switch key.(type) {
				case string:
					sub.Type = KEY
				}
				sub.Value = key
				sub.Tokens = nil
			}
		}

		return getKeyValues(t, input)
	case FUNC:
		if len(t.Tokens) == 0 {
			_, ok := opMap.Nullary[tokenVal]
			if !ok {
				return nil, errors.New(fmt.Sprintf("func does not exist or wrong num of arguments: %s", tokenVal))
			}
			return opMap.Nullary[tokenVal]()
		}
		if len(t.Tokens) == 1 {
			a, err := EvalCustom(opMap, t.Tokens[0], msg)
			if err != nil {
				return nil, err
			}

			_, ok := opMap.Unary[tokenVal]
			if !ok {
				return nil, errors.New(fmt.Sprintf("func does not exist or wrong num of arguments: %s", tokenVal))
			}

			return opMap.Unary[tokenVal](a)
		} else if len(t.Tokens) == 3 {

			a, err := EvalCustom(opMap, t.Tokens[0], msg)
			if err != nil {
				return nil, err
			}

			b, err := EvalCustom(opMap, t.Tokens[2], msg)
			if err != nil {
				return nil, err
			}

			_, ok := opMap.Binary[tokenVal]
			if !ok {
				return nil, errors.New(fmt.Sprintf("func does not exist or wrong num of arguments: %s", tokenVal))
			}

			return opMap.Binary[tokenVal](a, b)
		}
		return nil, errors.New(fmt.Sprintf("func does not exist or wrong num of arguments: %s", tokenVal))
	default:
		if len(t.Tokens) > 0 {
			return EvalCustom(opMap, t.Tokens[0], msg)
		}
	}

	return nil, nil
}

func FmtTokens(tl []*Token) {
	for _, t := range tl {
		fmt.Printf("(" + IdentStr[t.Type] + " " + t.Value + ") ")
	}
}

func FmtTokenTree(tree *TokenTree, d int) {
	fmt.Printf("\n")
	for i := 0; i < d; i++ {
		fmt.Printf("  ")
	}

	fmt.Printf("[")
	if tree.Type != ZERO {
		fmt.Printf("%s ", IdentStr[tree.Type])
	}
	fmt.Printf("%s", tree.Value)
	d++
	for _, t := range tree.Tokens {
		FmtTokenTree(t, d)
	}
	fmt.Printf("]")
}
