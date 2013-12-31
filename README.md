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

$len(), $contains(), $keys(), $has(), $sum(), $min(), $max(), $pow(), $sqrt(), $abs(), $floor(), $exists(), $regex()

see `jee_test.go` for more examples.

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