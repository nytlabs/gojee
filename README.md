# jee 
**jee** (json expression evaluator) transforms JSON  through logical and mathematical expressions. jee can be used from the command line or as a Go package. It is inspired by the fantastic (and much more fully featured) [jq]("http://stedolan.github.io/jq/"). 

jee was created out of the need for a simple JSON query language in [streamtools]("https://github.com/nytlabs/streamtools/"). jee is designed for stream processing and provides a reusable token tree. 

####get the library

    go get github.com/nytlabs/gojee

####make and install the binary

    cd $gopath/src/github.com/nytlabs/gojee/jee
    go install


### usage (binary)
##### querying JSON

get the entire input object:

    > echo '{"a": 3}' | jee '.'
    {
        "a": 3
    }

get a value for a specific key:

    > echo '{"a": 3, "b": 4}' | jee '.a'
    3

get a value from an array:

    > echo '{"a": [4,5,6]}' | jee '.a[0]'
    4

get all values from an array:

    > echo '{"a": [4,5,6]}' | jee '.a[]'
    [
        4,
        5,
        6
    ]

query all objects inside array 'a' for key 'id':

    > echo '{"a": [{"id":"foo"},{"id":"bar"},{"id":"baz"}]}' | jee '.a[].id'
    [
        "foo",
        "bar",
        "baz"
    ]


##### arithmetic 
\+ - * /

    > echo '{"a": 10}' | jee '(.a * 100)/-10 * 5'
    -500
    
##### comparison 
\> >= < <= !=

    > echo '{"a": 10}' | jee '(.a * 100)/-10 * 5 == -500'
    true
    > echo '{"a": 10}' | jee '(.a * 100)/-10 * 5 > 0'
    false

##### logical
|| &&
    
    > echo '{"a": false}' | jee '!(.a && true) || false  == true'
    true
    
##### functions

###### types

`$num(x {bool, float64, string, nil})`

Converts `x` to a float64. If `x` is a bool, 1 is returned for true and 0 for false. If `x` is nil, 0 is returned. 

`$str(x {bool, float64, string, nil, object))`

Converts `x` to a string. If `x` is a bool, "true" is returned for true and "false" for false. "null" is returned for nil. If `x` is an object it is marshaled into a JSON string. 

###### math

`$sqrt(x float64)`

Returns square root of `x`.

`$pow(x float64, y float64)`

Returns `x`^`y`.

`$floor(x float64)`

Returns nearest downward integer for `x`.

`$abs(x float64)`

Returns absolute value of `x`.

###### arrays

`$len(a []interface{})`

Returns the length of array `a`. 

`$has( a {[]bool, []float64, []string, []nil}, val {bool, float64, string, nil} )`

Checks to see if array `a` contains `val`. Returns bool. `val` cannot be an object.

`$sum(a []float64)`

Returns the sum of array `a`.

`$min(a []float64)`

Returns the minumum of array `a`.

`$max(a []float64)`

Returns the maximum of array `a`.

###### objects

`$keys(o object)`

Returns an array of keys in object `o`.


`$exists(o object, key string)`

Checks to see if `key` exists in map `o`. Returns bool. `$exists()` does a map lookup and is faster than `$has($keys(o), "foo")`

###### date and time

`$now()`

Returns current system time in float64.

`$parseTime(layout string, t string)`

Accepts a time layout in golang [time format](http://golang.org/pkg/time/#pkg-constants). t is parsed and returned as epoch milliseconds in float64.

`$fmtTime(layout string, t float64)`

Accepts a time layout in golang [time format](http://golang.org/pkg/time/#pkg-constants). t is expected in epoch milliseconds. Returns a formatted string. 

###### strings

`$contains(s string, substr string)`

see [strings.Contains](http://golang.org/pkg/strings/#Contains)

`$regex(pattern string, s string)`

see [regexp.MatchString](http://golang.org/pkg/regexp/#MatchString). Much slower than `$contains()`


see `jee_test.go` for examples.

### package usage
#####`Lexer(string) []*Token, error`
converts a jee query string into a slice of tokens

#####`Parser([]*Tokens) *TokenTree, error`
builds a parse tree out token slice from `Lexer()`

#####`Eval(*TokenTree, {}interface) {}interface, error`
evaluates a variable of type interface{} with a *TokenTree generated from `Parser()`. Only types given by [`json.Unmarshal`]("http://golang.org/pkg/encoding/json/#Unmarshal") are supported.

### quirks
* Types are strictly enforced. `false || "foo"` will produce a type error.
* `null` and `0` are not falsey
* Using a JSON key as an array index or an escaped key in bracket notation will not currently be evaluated. ie: `.a[.b]`
* All numbers in a jee query must start with a digit. numbers <1 should start with a 0. use `0.1` instead of `.1`
* Bracket notation is available for keys that need escaping `.["foo"]["bar"]`]
* Queries for JSON keys or indices that do not exist return `null` (to test if a key exists, use `$exists`)
* jee does not support variables, conditional expressions, or assignment 
* jee is v0.1.0 and may be very quirky in general.