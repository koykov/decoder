# Decoder

Dynamic decoders based on [inspector](https://github.com/koykov/inspector) framework
and vector packages, like [jsonvector](https://github.com/koykov/jsonvector). 

## Basics

One of the major the problems we ran into was a necessity to convert tons of different response formats from external
services into our internal response format. Also, there was a requirement to connect new services at runtime and provide
a possibility to decode their responses without rebooting the application.

This package provides a possibility to describe decoding rules in Go-like meta-language and add new decoders or edit
existing at runtime.

Decoders is a context-based and all variables should be preliminarily registered in special [context](ctx.go) before
decoding.

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
ctx.Set("dstObj", dst, &DstInspector{})
ctx.Set("user", user, &UserInspector{})
```
In this example user is a Go struct, but you may use as source raw response, example in JSON
```go
jsonResponse = []byte(`{"a":"foo"}`)
ctx := decoder.AcquireCtx()
ctx.SetJson("response", jsonResponse)
// in decoder body:
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
func modMyCustomMod(ctx *Ctx, buf *interface{}, val interface{}, args []interface{}) error {
    // ...
}

decoder.RegisterModFn("myCustomMod", "customMod", modMyCustomMod)
```

Modifier arguments:
* `ctx` is a storage of variables/buffers you may use.
* `buf` is a type-free buffer that receives result of modifier's work. Please note the type `*interface{}` is an
alloc-free trick.
* `val` if a value of variable from left side of modifier separator (`|`).
* `args` array of modifier arguments, specified in rule. 

See [mod.go](mod.go) for details and [mod_builtin.go](mod_builtin.go) for more example of modifiers.
