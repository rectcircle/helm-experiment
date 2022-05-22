package main

import (
	"fmt"
	"os"
	"text/template"
)

type Foo struct {
	Bar string
	Baz int
}

func (f *Foo) String() string {
	return fmt.Sprintf("Foo{ Bar: %s, Baz: %d }", f.Bar, f.Baz)
}

func (f *Foo) MethodHasParam(a, b string) string {
	return fmt.Sprintf("a: %s, b: %s", a, b)
}

func (f *Foo) Add(i int) *Foo {
	return &Foo{Bar: f.Bar, Baz: f.Baz + i}
}

var data = map[string]interface{}{
	"Foo": &Foo{
		Bar: "bar",
		Baz: 42,
	},
	"Field": "string",
	"List": []string{
		"a", "b", "c",
	},
	"EmptyList": []string{},
	"Cond":      true,
	"Func": func(p string) string {
		return "func called: " + p
	},
	"Map": map[string]string{
		"a": "1",
		"b": "2",
	},
	"Nil": map[string]string(nil),
}

var globalFuncs = template.FuncMap{
	"version": func() string { return "v1.2.3" },
	"add":     func(a, b int) int { return 1 + 1 },
}

const tpl = `1. 不使用 {{"{{}}"}} 包裹的字符串会被原样输出：
	没有两个 { 开头和两个 } 结尾包裹的字符串被原样渲染。

2. 使用 {{"{{}}"}} 包裹，可以使用模板引擎提供的动态渲染的能力：
	{{ "模板语法" }}

3. 前导和后续空白字符删除
	a. {{"{{- 模板语法 }}"}} ，会删除前导空白字符：

		abc

		{{- "前导空白字符会被删除"}}
		def

	b. {{"{{ 模板语法 -}}"}} ，会删除后续空白字符：

		abc
		{{ "后续空白字符会被删除" -}}

		def

	c. {{"{{- 模板语法 -}}"}} ，会删除前导和后续空白字符：

		abc

		{{- "前导和后续空白字符会被删除" -}}

		def

4. 注释语法 {{ "{{/* a comment */}}" }} 注释内容不会被渲染：
	abc {{/* a comment */}} def

5. 字面量：
	a. bool {{true}}
	b. 字符串 {{"abc"}}
	c. 字符 {{'a'}}
	d. int (位数取决于操作系统) {{ 1 }}
	e. float {{ 1.1 }}
	f. 虚数 {{ 1i }}
	g. 复数 {{ 1+1i }}


6. 渲染数据 {{ "{{ .Field }}" }}：
	a. . 是一个特殊的变量，默认指向的是 Execute 函数 data 参数：{{ . }}
	b. 渲染 Field 字段：{{ .Field }}
	c. 渲染 Foo 结构体的 Bar 字段：{{ .Foo.Bar }}
	d. 渲染并调用无参数方法：{{ .Foo.String }}
	d. 调用并渲染有参数方法：{{ .Foo.MethodHasParam "a" "b" }}
	e. 渲染 data 的 List 切片的 0 号元素：{{ index .List 0 }}
	f. 调用渲染 data 的 函数类型变量 Func：{{ call .Func "abc" }}
	g. 通过 key 渲染 map 的元素：{{ index .Map "a" }} 或 {{ .Map.a }}

7. 临时改变 . 的指向: 
	with: {{ with .Foo }} Bar = {{ .Bar }}, Baz = {{ .Baz }} {{ end }}
	with-else: {{ with .Nil }} {{.}} {{else}} . 是 nil {{ end }}

8. 流程控制
	a. 条件渲染：
		{{if .Cond}} .Cond is true {{end}}
		{{if not .Cond}} .Cond is false {{else}} .Cond is true {{end}} 
		{{if lt .Foo.Baz 10 }} .Foo.Baz < 10 {{ else if lt .Foo.Baz 100 }} 10 <= .Foo.Baz < 100 {{else}} .Foo.Baz >= 100 {{end}} 
	b. 遍历（只支持  array, slice, map, or channel）：
		遍历并改变 . 指向：
			range-end: {{range .List}} {{.}}, {{end}}
			range-else-end: {{range .EmptyList}} {{.}}, {{else}} is empty list {{end}}
			range-if-break-continue: {{range .List}} {{.}}, {{if eq . "a"}} {{continue}} {{else}} {{break}} {{end}}  {{end}}
		遍历不改变 . 指向，并获取索引：
			range-with-index: {{range $i, $e := .List}} {{$i}}: {{$e}}, {{end}}
		遍历 map：
			range-map: {{range $k, $v := .Map}} {{$k}}: {{$v}}, {{end}}

9. Arguments 概念，如下几种语法都叫做 Arguments：
	a. 上文提到的字面量: {{1}}
	b. nil 关键字：{{printf "%v" nil}}
	c. 数据指向数据的引用 . 以及对数据中的字段的引用(map 或 结构体) .Field: {{.Field}}
	d. $ 等对变量的引用，以及对变量的字段的引用（map 或结构体） $Xxx.Xxx，下文有阐述
	e. 数据或变量中，无参方法调用：{{.Foo.String}}
	f. 无参全局函数：{{version}}
	g. 上述之一的带括号的实例，用于分组。结果可以通过字段或映射键调用来访问。
		{{ printf ".List[0]=%s" (index .List 0) }}
		{{ (.Foo.Add 1).String }}

10. 全局函数调用
	语法为 {{ "{{函数名 Arguments1 Arguments2 ...}}" }}: {{ printf "%s" "hello" }}
	内置全局函数(参见 https://pkg.go.dev/text/template#hdr-Functions ）调用：{{ printf "%s" "hello" }}
	自定义全局函数，通过 func (*Template).Funcs(funcMap template.FuncMap) *template.Template 函数添加：{{add 1 2}}

11. Pipelines 能力，和 shell 中的 Pipelines 类似，如下元素可以通过 | 管道符连接。
	a. 语法为: Argument | 函数或方法（可选） | 函数或方法（可选） | ...
		Argument 就是上文定义的
		函数或方法可能是：
			.XXx.XxxMethod [Argument...] 保证最后一个是有参数有返回值的方法
			$Xxx.XxxMethod [Argument...] 保证最后一个是有参数有返回值的方法
			有参数有返回值的全局函数 Func [Argument...]
		函数或方法的写法，不应书写最后一个参数，因为最后一个参数从 Pipeline 中来
	d. 一个例子： {{"Pipeline" | printf "hello %s"}}

12. 变量
	a. 定义 $variable := pipeline: {{ $variable := "value" }} {{ $variable }}
	b. 赋值 $variable = pipeline: {{ $variable = "override" }} {{ $variable }}

13. 嵌套模板
	a. 定义: {{define "T1"}}ONE{{end}}
	b. 调用: {{template "T1" .}}
	c. 不要求定义一定发生在调用之前：{{template "T2" .}} {{define "T2"}}TWO{{end}}
	d. 同一个 tpl 不允许重复定义，但是多个模板可以重复定义，后 Parse 的将覆盖之前的嵌套模板
		{{template "T3" .}} {{define "T3"}}THREE{{end}}
	e. block 定义一个模板并立即调用，等价于 define 后立即 template 调用，用于实现给 template 提供一个默认值。
		{{block "T4" .}}T4 没有定义{{end}}
		{{block "T5" .}}T5 没有定义{{end}}

14. Parse 多个模板时，Execute 函数，只会渲染最后一个包含实际内容的模板（不包含实际内容的模板指的是该模板只包含 define 没有其他内容）。

15. Action 和 函数
	在 Go 模板中 Action 和 全局函数，使用起来看似相同。但是 Action 不支持 Pipeline。
	除了 if 、else、range 这类流程控制的 Action 外，template 就是一个 Action，因此 template 不支持管道符。
`

const tpl2 = `{{define "T3"}}THREE 来自 tpl2{{end}}{{define "T5"}}T5 有定义{{end}}`

func main() {
	t, err := template.New("tpl").Funcs(globalFuncs).Parse(tpl)
	if err != nil {
		panic(err)
	}
	_, err = t.Parse(tpl2)
	if err != nil {
		panic(err)
	}

	err = t.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}
