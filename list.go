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

import (
	"strconv"
)

// List 描述了列表节点结构。
type List struct {
	*BaseNode
	*listData
}

// listData 用于记录列表或列表项节点的附加信息。
type listData struct {
	typ          int   // 0：无序列表，1：有序列表，3：任务列表
	tight        bool  // 是否是紧凑模式
	bulletChar   items // 无序列表标识，* - 或者 +
	start        int   // 有序列表起始序号
	delimiter    byte  // 有序列表分隔符，. 或者 )
	padding      int   // 列表内部缩进空格数（包含标识符长度，即规范中的 W+N）
	markerOffset int   // 标识符（* - + 或者 1 2 3）相对缩进空格数
	checked      bool  // 任务列表项是否勾选
	marker       items // 列表标识符
	num          int   // 有序列表项修正过的序号
}

func (list *List) CanContain(nodeType int) bool {
	return NodeListItem == nodeType
}

func (list *List) Finalize(context *Context) {
	item := list.firstChild

	// 检查子列表项之间是否包含空行，包含的话说明该列表是非紧凑的，即松散的
	for nil != item {
		if list.endsWithBlankLine(item) && nil != item.Next() {
			list.tight = false
			break
		}

		var subitem = item.FirstChild()
		for nil != subitem {
			if list.endsWithBlankLine(subitem) &&
				(nil != item.Next() || nil != subitem.Next()) {
				list.tight = false
				break
			}
			subitem = subitem.Next()
		}
		item = item.Next()
	}
}

// parseListMarker 用于解析泛列表（列表、列表项或者任务列表）标记。
func (t *Tree) parseListMarker(container Node) *listData {
	if t.context.indent >= 4 {
		return nil
	}

	ln := t.context.currentLine // 弄短点
	tokens := ln[t.context.nextNonspace:]
	data := &listData{
		typ:          0,                // 默认无序列表
		tight:        true,             // 默认紧凑模式
		markerOffset: t.context.indent, // 设置前置相对缩进
		num:          -1,               // 假设有序列表起始为 -1，后面会进行计算赋值
	}

	markerLength := 1
	marker := items{tokens[0]}
	var delim byte
	if itemPlus == marker[0] || itemHyphen == marker[0] || itemAsterisk == marker[0] {
		data.bulletChar = marker
	} else if marker, delim = t.parseOrderedListMarker(tokens); nil != marker {
		if container.Type() != NodeParagraph || "1" == fromItems(marker) {
			data.typ = 1 // 有序列表
			data.start, _ = strconv.Atoi(fromItems(marker))
			markerLength = len(marker) + 1
			data.delimiter = delim
		} else {
			return nil
		}
	} else {
		return nil
	}

	data.marker = marker

	var token byte

	// 列表项标记后必须是空白字符
	if token = ln[t.context.nextNonspace+markerLength]; !isWhitespace(token) {
		return nil
	}

	// 如果要打断段落，则列表项内容部分不能为空
	if container.Type() == NodeParagraph && itemNewline == ln[t.context.nextNonspace+markerLength] {
		return nil
	}

	// 到这里说明满足列表规则，开始解析并计算内部缩进空格数
	t.context.advanceNextNonspace()             // 把起始下标移动到标记起始位置
	t.context.advanceOffset(markerLength, true) // 把结束下标移动到标记结束位置
	spacesStartCol := t.context.column
	spacesStartOffset := t.context.offset
	for {
		t.context.advanceOffset(1, true)
		token = ln.peek(t.context.offset)
		if t.context.column-spacesStartCol >= 5 || itemEnd == token || (itemSpace != token && itemTab != token) {
			break
		}
	}

	token = ln.peek(t.context.offset)
	var isBlankItem = itemEnd == token || itemNewline == token
	var spaces_after_marker = t.context.column - spacesStartCol
	if spaces_after_marker >= 5 || spaces_after_marker < 1 || isBlankItem {
		data.padding = markerLength + 1
		t.context.column = spacesStartCol
		t.context.offset = spacesStartOffset
		if token = ln.peek(t.context.offset); itemSpace == token || itemTab == token {
			t.context.advanceOffset(1, true)
		}
	} else {
		data.padding = markerLength + spaces_after_marker
	}

	if !isBlankItem {
		// 判断是否是任务列表项
		content := ln[t.context.offset:]
		if 3 <= len(content) { // 至少需要 [ ] 或者 [x] 3 个字符
			if itemOpenBracket == content[0] && ('x' == content[1] || 'X' == content[1] || itemSpace == content[1]) && itemCloseBracket == content[2] {
				data.typ = 3
				data.checked = 'x' == content[1] || 'X' == content[1]
			}
		}
	}

	return data
}

func (t *Tree) parseOrderedListMarker(tokens items) (marker items, delimiter byte) {
	length := len(tokens)
	var i int
	var token byte
	for ; i < length; i++ {
		token = tokens[i]
		if !isDigit(token) || 8 < i {
			delimiter = token
			break
		}
		marker = append(marker, token)
	}

	if 1 > len(marker) || (itemDot != delimiter && itemCloseParen != delimiter) {
		return nil, itemEnd
	}

	return
}

// endsWithBlankLine 判断块节点 block 是否是空行结束。如果 block 是列表或者列表项则迭代下降进入判断。
func (list *List) endsWithBlankLine(block Node) bool {
	for nil != block {
		if block.LastLineBlank() {
			return true
		}
		t := block.Type()
		if !block.LastLineChecked() && (t == NodeList || t == NodeListItem) {
			block.SetLastLineChecked(true)
			block = block.LastChild()
		} else {
			block.SetLastLineChecked(true)
			break
		}
	}

	return false
}
