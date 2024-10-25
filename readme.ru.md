# Decoder

Динамический декодер на базе фреймворка [inspector](https://github.com/koykov/inspector/blob/master/readme.ru.md) и
[векторных парсеров](https://github.com/koykov/vector/blob/master/readme.ru.md).

## Ретроспектива

Одной из значительных проблем в highload проекте была необходимость привести огромное количество разнородных ответов от
внешних сервисов к единому внутреннему формату ответа. Причём проблема усугублялась тем, что новые внешние сервисы с
собственным форматом ответа могли появляться в любой момент. Т.к. проект работал в хайлоаде, то использовать стандартные
способы динамики, такие как рефлексия, было нельзя - конвертация должна была происходить очень быстро, не плодить
аллокации и поддерживать динамику, чтобы избежать траты на деплои приложения.

Эта библиотека была разработана в ответ на эти вызовы. Она предоставляет возможность описывать правила декодирования на
метаязыке похожем синтаксисом на Go с полной поддержкой динамичности - изменить существующий декодер или добавить новый
можно на лету, без перезагрузки приложения.

## Принцип работы

Декодеры во многом похожи на библиотку [dyntpl](https://github.com/koykov/dyntpl/blob/master/readme.ru.md) только
наоборот - dyntpl призван генерировать данные в текст, а decoder преобразовывать текст в данные.

Аналогично dyntpl, декодирование поделено на два этапа - парсинг и декодирование. В процессе первого этапа на основе тела
декодера строится специальное дерево (аналог AST) и его необходимо зарегистрировать в регистре декодеров под уникальным
именем. Этот этап не предназначен для использования в хайлоаде, т.к. он очень тяжёлый и дорогой.
Второй этап - декодирование, напротив оптимизирован для использования в хайлоаде.

Декодер является контекстно-зависимым и все переменные, с которыми он оперирует, должны быть заранее зарегистрированы в
объекте [Ctx](ctx.go) до начала декодирования. Каждая переменная задаётся тремя параметрами:
* уникальное имя
* данные - как парвило это какая-то структура, но может быть чем угодно
* тип-инспектор

Что такое тип инспектор описывается [тут](https://github.com/koykov/inspector/blob/master/readme.ru.md#%D0%B2%D0%B2%D0%B5%D0%B4%D0%B5%D0%BD%D0%B8%D0%B5),
но следует объяснить как эта теория реализуется на практике. В предыдущем разделе уже касались условий, но с точки зрения
программиста проблему можно описать так "как получить произвольные данные из одной структуры и записать их в другую
максимально быстро и с нулевым или околонулевым количеством дополнительной памяти". Проблема получения данных была решена
в [dyntpl с помощью фреймворка inspector](https://github.com/koykov/inspector/blob/master/readme.ru.md#%D0%B2%D0%B2%D0%B5%D0%B4%D0%B5%D0%BD%D0%B8%D0%B5)
и поэтому логичным решением стало использовать те же инспекторы для записи данных в структуры. Таким образом, общий
принцип работы свёлся к задаче "с помощью инспетора прочитать данные из переменной-источника и с помощью другого
инспектора записать их в переменную-приёмник".

## Пример использования

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

Содержимое функции init должно выполняться один раз (или периодически, на лету, с обновлением шаблона из какого-то
источника, например БД).

Содержимое функции main это пример использования декодеров в хайлоаде.

## Синтаксис

### Присваивание

Базовой операцией при декодировании является присваивание данных из переменной-источника к пременной-приёмнику. Это
обеспечивается типичнной операцией присваивания `=` и имеет вид `lvalue.field1 = rvalue.field2`. Из примера использования выше:
```
data.Id = resp.identifier
data.Name = resp.person.full_name
```
, где `data` это `lvalue` (или переменная-приёмник), а `resp` это `rvalue` (или переменная-источник).

Здесь важно понять что присходит неявно для пользователя. За каждой из переменных закреплён свой инспектор:
* `data` - [testobj_ins.TestObjectInspector](https://github.com/koykov/inspector/blob/master/testobj_ins/testobject_ins.go#L19)
* `resp` - [vector_inspector.VectorInspector](https://github.com/koykov/vector_inspector/blob/master/inspector.go#L12) (задаётся неявно вызовом [Ctx::SetVector](https://github.com/koykov/decoder/blob/master/ctx.go#L120) или [Ctx::SetVectorNode](https://github.com/koykov/decoder/blob/master/ctx.go#L128))

Когда декодеру необходимо произвести присваивание `data.Name = resp.person.full_name`, то процесс делится на два этапа:
* экземпляр `VectorInspector`-а с помощью метода [`GetTo`](https://github.com/koykov/vector_inspector/blob/master/inspector.go#L26) читает из `resp` данные по пути `person.full_name`
* экземпляр `TestObjectInspector`-а с помощью метода [`SetWithBuffer`](https://github.com/koykov/inspector/blob/master/testobj_ins/testobject_ins.go#L750) записывает данные в `data` по пути `Name`

В итоге данные перенесены (или скопированы с буфферизацией) из `rvalue` в `lvalue`.