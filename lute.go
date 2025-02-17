// Lute - A structured markdown engine.
// Copyright (c) 2019-present, b3log.org
//
// Lute is licensed under the Mulan PSL v1.
// You can use this software according to the terms and conditions of the Mulan PSL v1.
// You may obtain a copy of Mulan PSL v1 at:
//     http://license.coscl.org.cn/MulanPSL
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v1 for more details.

package lute

// Lute 描述了 Lute 引擎的顶层使用入口。
type Lute struct {
	options
}

// New 创建一个新的 Lute 引擎，默认启用：
//  * GFM 支持
//  * 代码块语法高亮
//  * 软换行转硬换行
//  * 中西文间插入空格
//  * 修正术语拼写
func New(opts ...option) (ret *Lute) {
	ret = &Lute{}
	GFM(true)(ret)
	SoftBreak2HardBreak(true)(ret)
	CodeSyntaxHighlight(true)(ret)
	AutoSpace(true)(ret)
	FixTermTypo(true)(ret)
	for _, opt := range opts {
		opt(ret)
	}
	return ret
}

// Markdown 将 markdown 文本字符数组处理为相应的 html 字符数组。name 参数仅用于标识文本，比如可传入 id 或者标题，也可以传入 ""。
func (lute *Lute) Markdown(name string, markdown []byte) (html []byte, err error) {
	var tree *Tree
	tree, err = parse(name, markdown, lute.options)
	if nil != err {
		// fmt.Println(string(markdown))
		return
	}

	renderer := newHTMLRenderer(lute.options)
	html, err = tree.render(renderer)
	return
}

// MarkdownStr 接受 string 类型的 markdown 后直接调用 Markdown 进行处理。
func (lute *Lute) MarkdownStr(name, markdown string) (html string, err error) {
	var htmlBytes []byte
	htmlBytes, err = lute.Markdown(name, toItems(markdown))
	if nil != err {
		return
	}

	html = fromItems(htmlBytes)
	return
}

// Format 将 markdown 文本字符数组进行格式化。
func (lute *Lute) Format(name string, markdown []byte) (formatted []byte, err error) {
	var tree *Tree
	tree, err = parse(name, markdown, lute.options)
	if nil != err {
		// fmt.Println(string(markdown))
		return
	}

	renderer := newFormatRenderer(lute.options)
	formatted, err = tree.render(renderer)
	return
}

// FormatStr 接受 string 类型的 markdown 后直接调用 Format 进行处理。
func (lute *Lute) FormatStr(name, markdown string) (formatted string, err error) {
	var formattedBytes []byte
	formattedBytes, err = lute.Format(name, toItems(markdown))
	if nil != err {
		return
	}

	formatted = fromItems(formattedBytes)
	return
}

// GFM 设置是否打开所有 GFM 支持。
func GFM(b bool) option {
	return func(lute *Lute) {
		lute.GFMTable = b
		lute.GFMTaskListItem = b
		lute.GFMStrikethrough = b
		lute.GFMAutoLink = b
	}
}

// GFMTable 设置是否打开“GFM 表”支持。
func GFMTable(b bool) option {
	return func(lute *Lute) {
		lute.GFMTable = b
	}
}

// GFMTaskListItem 设置是否打开“GFM 任务列表项”支持。
func GFMTaskListItem(b bool) option {
	return func(lute *Lute) {
		lute.GFMTaskListItem = b
	}
}

// GFMStrikethrough 设置是否打开“GFM 删除线”支持。
func GFMStrikethrough(b bool) option {
	return func(lute *Lute) {
		lute.GFMStrikethrough = b
	}
}

// GFMAutoLink 设置是否打开“GFM 自动链接”支持。
func GFMAutoLink(b bool) option {
	return func(lute *Lute) {
		lute.GFMAutoLink = b
	}
}

// SoftBreak2HardBreak 设置是否将软换行（\n）渲染为硬换行（<br />）。
func SoftBreak2HardBreak(b bool) option {
	return func(lute *Lute) {
		lute.SoftBreak2HardBreak = b
	}
}

// CodeSyntaxHighlight 设置是否对代码块进行语法高亮。
func CodeSyntaxHighlight(b bool) option {
	return func(lute *Lute) {
		lute.CodeSyntaxHighlight = b
	}
}

// AutoSpace 设置是否对普通文本中的中西文间自动插入空格。
// https://github.com/sparanoid/chinese-copywriting-guidelines
func AutoSpace(b bool) option {
	return func(lute *Lute) {
		lute.AutoSpace = b
	}
}

// FixTermTypo 设置是否对普通文本中出现的术语进行修正。
// https://github.com/sparanoid/chinese-copywriting-guidelines
func FixTermTypo(b bool) option {
	return func(lute *Lute) {
		lute.FixTermTypo = b
	}
}

// options 描述了一些列解析和渲染选项。
type options struct {
	GFMTable            bool
	GFMTaskListItem     bool
	GFMStrikethrough    bool
	GFMAutoLink         bool
	SoftBreak2HardBreak bool
	CodeSyntaxHighlight bool
	AutoSpace           bool
	FixTermTypo         bool
}

// option 描述了解析渲染选项设置函数签名。
type option func(lute *Lute)
