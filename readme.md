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

---

## Syntax

Recommend reading [inspector](https://github.com/koykov/inspector) first.

Each decoder consists of name and body. Name uses only as index and may contain any symbols. There is only one
requirement - name should be unique.

Decoder's body is a multi-line string (that calls ruleset) and each line (rule) has the following format
```
<destination path> = [<source path>|<callback>]
```

Both `<source path>` and `<destination path>` describes a source/destination fields accessed using a dot, example:
```
dstObj.Name = user.FullName
```
As mentioned above, both `dstObj` and `user` variables should be preliminarily registered in the context like this
```go
ctx := decoder.AcquireCtx()
ctx.Set("dstObj", dst, DstInspector{})
ctx.Set("user", user, UserInspector{})
```
In this example user is a Go struct, but you may use as source raw response, example in JSON
```go
import _ "github.com/koykov/decoder_vector"
...
jsonResponse = []byte(`{"a":"foo"}`)
ctx := decoder.AcquireCtx()
ctx.SetStatic("raw", jsonResponse)
// in decoder body:
// ctx.response = vector::parseJSON(raw)
// dstObj.Name = response.a
```

In this way, decoder provides a possibility to describe where source data should be taken and where it should be came.

Enough easy, isn't it?

### Modifiers

Sometimes, isn't enough just specify source data address, but need to modify it before assigning to destination.

Especially for that cases was added support of source modifiers. The syntax:
```
<destination path> = <source path>|<modifier name>(<arg0>, <arg1>, ...)
```
Example
```
dstObj.Balance = user.Finance.Rest|default(0)
```

Default is a built-in modifier, but you may register your own modifiers using modifiers registry:
```go
func modMyCustomMod(ctx *Ctx, buf *any, val any, args []any) error {
    // ...
}

decoder.RegisterModFn("myCustomMod", "customMod", modMyCustomMod)
```

Modifier arguments:
* `ctx` is a storage of variables/buffers you may use.
* `buf` is a type-free buffer that receives result of modifier's work. Please note the type `*any` is an
alloc-free trick.
* `val` if a value of variable from left side of modifier separator (`|`).
* `args` array of modifier arguments, specified in rule.

See [mod.go](mod.go) for details and [mod_builtin.go](mod_builtin.go) for more example of modifiers.

## Extensions

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
