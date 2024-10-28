# Decoder

Dynamic decoders based on [inspector](https://github.com/koykov/inspector) framework
and [vector parsers](https://github.com/koykov/vector).

## Retrospective

One of the major the problems we ran into was a necessity to convert tons of different response formats from external
services into internal response format. The problem became hardener due to new external services with own response
formats may appear at any time. Due to highload conditions there is no way to use standard dynamic approaches like
reflection - the convertation must work very fast, make zero allocations nd support dynamic to avoid application deploys.

This package was developed as an answer to this challenge. It provides a possibility to describe decoding rules in
Go-like meta-language with full dynamic support - registering new decoders (or edit en existing) may on the fly.

## How it works

Decoders are similar to [dyntpl](https://github.com/koykov/dyntpl) package in opposite - dyntpl makes a text from
structures but decoders parses text and assign data to structures.

Similar to dyntpl, decoding divides to two phases - parsing and decoding. The parsing phase builds from decoder's body
a tree (like AST) and registers it in decoders registry by unique name afterward. This phase isn't intended to be used in
highload conditions due to high pressure to cpu/mem. The second phase - decoding, against intended to use in highload.

Decoding phase required a preparation to pass data to the decoder. There is a special object [Ctx](ctx.go), that collects
variables to use in decoder. Each variable must have three params:
* unique name
* data - anything you need to use in template
* inspector type

What is the inspector describes [here](https://github.com/koykov/inspector), but need an extra explanation how it works
together with decoders. In general, decoding problem sounds like "grab an arbitrary data from one struct and write it
to another struct as fast as it possible and with zero allocations". The first part of the problem was solved in
[dyntpl using inspectors,](https://github.com/koykov/inspector/tree/master?tab=readme-ov-file#intro) and it was a good
decision to extend inspectors with possibility to write data to destination structs. Thus, the problem became like
"using one inspector, read data from source struct and, using another inspector, write it to destination struct".

## Usage

The typical usage of decoders looks like this:
```go
package main

import (
	"github.com/koykov/decoder"
	"github.com/koykov/inspector/testobj"
	"github.com/koykov/inspector/testobj_ins"
	"github.com/koykov/jsonvector"
)

var (
	data     testobj.TestObject
	response = []byte(`{"identifier":"xf44e","person":{"full_name":"Marquis Warren","status":67},"finance":{"balance":"164.5962"","is_active":true}}`)
	decBody  = []byte(`data.Id = resp.identifier
data.Name = resp.person.full_name
data.Status = resp.person.status|default(-1)
data.Finance.Balance = atof(resp.finance.balance)`)
)

func init() {
	// Parse decoder body and register it.
	dec, _ := decoder.Parse(decBody)
	decoder.RegisterDecoderKey("myDecoder", dec)
}

func main() {
	// Prepare response as vector object.
	vec := jsonvector.Acquire()
	defer jsonvector.Release(vec)
	_ = vec.Parse(response)

	ctx := decoder.AcquireCtx()
	defer decoder.ReleaseCtx(ctx)
	
	// Prepare context.
	ctx.SetVector("resp", vec)
	ctx.Set("data", &data, testobj_ins.TestObjectInspector{})
	// Execute the decoder.
	err := decoder.Decode("myDecoder", ctx)
	println(err)                  // nil
	println(data.Id)              // xf44e
	println(data.Name)            // []byte("Marquis Warren")
	println(data.Status)          // 67
	println(data.Finance.Balance) // 164.5962
}
```

Content of init() function should be executed once (or periodically on the fly from some source, eg DB).

Content of main() function is how to use decoders in a general way in highload.

## Syntax

Decoders inherits Go syntax, but provides an extra features like modifiers and coalesce operator (see below).

### Assigning

The base decoding operation is assigning the data from source variable to destination variable. The syntax is typical
`lvalue.Field1 = rvalue.Field2`. From [example](#usage):
```
data.Id = resp.identifier
data.Name = resp.person.full_name
```
where `data` represents `lvalue` (source variable) and `resp` - `rvalue` (destination variable).

### Coalesce operator

Decoders provide a possibility to read one-of-many fields when read nested fields from struct:
```
dst.Field = src.Nested.{Field1|Field2|Field3|...}
```
The first non-empty field between curly brackets will be read as data to assign. This syntax sugar allows to avoid tons
of comparisons or build chain of `default` modifiers. Example of usage see [here](testdata/decoder/decoder4.dec).

### Modifiers

Decoders supports user-defined modifiers, which applies additional logic to data before assigning. It may be helpful for
edge cases (no data, conditional assignment, etc.). Modifiers usage syntax is typical - after source of data, using `|`
symbol, modifier calls as function call:
```
dst.Field = src.Field|modifier0(arg0, arg1, ...)|modifier1(arg0, arg1, ...)|...
```

Example:
```
data.Status = src.Nested.Blocked|ifThenElse(src.Nested.State, -1)
                                ^ simple modifier
data.Name = src.FullName|default("N\D")|toUpper()
                        ^ first mod    ^ second modifier
```

Modifiers may collect in chain with variadic length. In that case, each modifier will take to input the result of
previous modifier. Each modifier may take an arbitrary count of arguments.

Modifier is a Go function with special signature:
```go
type ModFn func(ctx *Ctx, buf *any, val any, args []any) error
```
where:
* ctx - context of the decoder
* buf - pointer to return the result
* val - value to pass to the modifier (value of `varName` in example `varName|modifier()`) 
* args - list of all arguments

You should register your modifier using one of the functions:
* `RegisterModFn(name, alias string, mod ModFn)`
* `RegisterModFnNS(namespace, name, alias string, mod ModFn)`

They are the same, but NS version allows to specify the namespace of the function. In that case, you should specify namespace
in modifiers call:
```
dst.Field = src.Field|namespaceName::modifier()
```

### Conditions

Decoders supports classic syntax of conditions:
```
if leftVar [==|!=|>|>=|<|<=] rightVar {
    true branch
} else {
    false branch
}
```

Examples: [1](testdata/parser/cond.dec), [2](testdata/parser/cond_else.dec), [3](testdata/parser/condOK.dec).

Decoders can't handle complicated conditions containing more than one comparison, like:
```
if user.Id == 0 || user.Finance.Balance == 0 {...}
```
In the future this problem will be solved, but now you can make nested conditions or use conditions helpers - functions
with signature:
```go
type CondFn func(ctx *Ctx, args []any) bool
```
, where you may pass an arbitrary amount of arguments and these functions will return bool to choose the right execution branch.
These function are user-defined, like modifiers, and you may write your own and then register it using one of the functions:
```go
func RegisterCondFn(name string, cond CondFn)
func RegisterCondFnNS(namespace, name string, cond CondFn) // namespace version
```

Then condition helper will be accessible inside templates and you may use it using the name:
```
if helperName(user.Id, user.Finance.Balance) {...}
```

For multiple conditions, you can use `switch` statement, examples:
* [classic switch](testdata/parser/switch.dec)
* [no-condition switch](testdata/parser/switch_no_cond.dec)
* [no-condition switch with helpers](testdata/parser/switch_no_cond_helper.dec)

### Loops

Decoders supports both types of loops:
* counter loops, like `for i:=0; i<5; i++ {...}`
* range-loop, like `for k, v := range obj.Items {...}`

Edge cases like `for k < 2000 {...}` or `for ; i < 10 ; {...}` isn't supported.
Also, you can't make an infinite loop by using `for {...}`.

#### Loop breaking

Decoders supports default instructions `break` and `continue` to break loop/iteration, example:
```
for _, v := list
  if v.ID == 0 {
    continue
  }
  if v.Status == -1 {
    break
  }
}
```

These instructions works as intended, but they required condition wrapper and that's bulky. Therefore, decoders provide
combined `break if` and `continue if` that works the same:
```
for _, v := list {
  continue if v.ID == 0
  break if v.Status == -1
}
```

The both examples are equal, but the second is more compact.

#### Lazy breaks

Imagine the case - you've decided in the middle of iteration that loop requires a break, but the iteration must finish its
work the end. For that case, decoders supports special instruction `lazybreak`. It breaks the loop but allows current
iteration works till the end.

### Extensions

Decoders may be extended by including modules to the project. Currently supported modules:
* [decoder_vector](https://github.com/koykov/decoder_vector) provide support of vector parsers.
* [decoder_i18n](https://github.com/koykov/decoder_legacy) allows legacy features in the project.

To enable necessary module just import it to the project, eg:
```go
import (
	_ "https://github.com/koykov/decoder_vector"
)
```
and vector's [features](https://github.com/koykov/decoder_vector) will be available inside decoders.

Feel free to develop your own extensions. Strongly recommend to register new modifiers using namespaces, like
[this](https://github.com/koykov/decoder_vector/blob/master/init.go#L15).

### Conclusion

Due to two phases (parsing and decoding) in using decoders it isn't handy to use in simple cases, especially outside
highload. The good condition to use it is a highload project and dynamic support requirement. Use decoders in proper
conditions and wish you happy decoding.
