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

import "bytes"

// Table 描述了表节点结构。
type Table struct {
	*BaseNode
	Aligns []int // 从左到右每个表格节点的对齐方式，0：默认对齐，1：左对齐，2：居中对齐，3：右对齐
}

// TableHead 描述了表头节点结构。
type TableHead struct {
	*BaseNode
}

// TableRow 描述了表行节点结构。
type TableRow struct {
	*BaseNode
	Aligns []int
}

// TableCell 描述了表格节点结构。
type TableCell struct {
	*BaseNode
	Aligns int
}

func (context *Context) parseTable(lines [][]byte) (ret *Table) {
	length := len(lines)
	if 2 > length {
		return
	}

	aligns := context.parseTableDelimRow(bytes.TrimSpace(lines[1]))
	if nil == aligns {
		return
	}

	headRow := context.parseTableRow(bytes.TrimSpace(lines[0]), aligns, true)
	if nil == headRow {
		return
	}

	ret = &Table{&BaseNode{typ: NodeTable}, aligns}
	ret.Aligns = aligns
	ret.AppendChild(ret, context.newTableHead(headRow))
	for i := 2; i < length; i++ {
		tableRow := context.parseTableRow(bytes.TrimSpace(lines[i]), aligns, false)
		if nil == tableRow {
			return
		}
		ret.AppendChild(ret, tableRow)
	}
	return
}

func (context *Context) newTableHead(headRow *TableRow) *TableHead {
	ret := &TableHead{&BaseNode{typ: NodeTableHead}}
	for c := headRow.FirstChild(); c != nil; {
		next := c.Next()
		ret.AppendChild(ret, c)
		c = next
	}
	return ret
}

func (context *Context) parseTableRow(line items, aligns []int, isHead bool) (ret *TableRow) {
	ret = &TableRow{&BaseNode{typ: NodeTableRow}, aligns}
	cols := line.splitWithoutBackslashEscape(itemPipe)
	if isBlank(cols[0]) {
		cols = cols[1:]
	}
	if len(cols) > 0 && isBlank(cols[len(cols)-1]) {
		cols = cols[:len(cols)-1]
	}

	colsLen := len(cols)
	alignsLen := len(aligns)
	if isHead && colsLen > alignsLen { // 分隔符行定义了表的列数，如果表头列数还大于这个列数，则说明不满足表格式
		return nil
	}

	var i int
	var col items
	for ; i < colsLen && i < alignsLen; i++ {
		col = bytes.TrimSpace(cols[i])
		cell := &TableCell{&BaseNode{typ: NodeTableCell}, aligns[i]}
		col = col.removeFirst(itemBackslash) // 删掉一个反斜杠来恢复语义
		cell.tokens = col
		ret.AppendChild(ret, cell)
	}

	// 可能需要补全剩余的列
	for ; i < alignsLen; i++ {
		cell := &TableCell{&BaseNode{typ: NodeTableCell}, aligns[i]}
		ret.AppendChild(ret, cell)
	}
	return
}

func (context *Context) parseTableDelimRow(line items) (aligns []int) {
	length := len(line)
	var token byte
	var i, pipes int
	for ; i < length; i++ {
		token = line[i]
		if itemPipe != token && itemHyphen != token && itemColon != token && itemSpace != token {
			return nil
		}
		if itemPipe == token {
			pipes++
		}
	}
	if 1 > pipes {
		return nil
	}

	cols := line.splitWithoutBackslashEscape(itemPipe)
	if isBlank(cols[0]) {
		cols = cols[1:]
	}
	if len(cols) > 0 && isBlank(cols[len(cols)-1]) {
		cols = cols[:len(cols)-1]
	}

	var alignments []int
	for _, col := range cols {
		col = bytes.TrimSpace(col)
		if 1 > length || nil == col {
			return nil
		}

		align := context.tableDelimAlign(col)
		if -1 == align {
			return nil
		}
		alignments = append(alignments, align)
	}
	return alignments
}

func (context *Context) tableDelimAlign(col items) int {
	var left, right bool
	length := len(col)
	first := col[0]
	left = itemColon == first
	last := col[length-1]
	right = itemColon == last

	i := 1
	var token byte
	for ; i < length-1; i++ {
		token = col[i]
		if itemHyphen != token {
			return -1
		}
	}

	if left && right {
		return 2
	}
	if left {
		return 1
	}
	if right {
		return 3
	}

	return 0
}
