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
	"bytes"
	"strconv"
)

// newFormatRenderer 创建一个格式化渲染器。
func newFormatRenderer(option options) (ret *Renderer) {
	ret = &Renderer{rendererFuncs: map[int]RendererFunc{}, option: option}

	// 注册 CommonMark 渲染函数

	ret.rendererFuncs[NodeDocument] = ret.renderDocumentMarkdown
	ret.rendererFuncs[NodeParagraph] = ret.renderParagraphMarkdown
	ret.rendererFuncs[NodeText] = ret.renderTextMarkdown
	ret.rendererFuncs[NodeCodeSpan] = ret.renderCodeSpanMarkdown
	ret.rendererFuncs[NodeCodeBlock] = ret.renderCodeBlockMarkdown
	ret.rendererFuncs[NodeEmphasis] = ret.renderEmphasisMarkdown
	ret.rendererFuncs[NodeStrong] = ret.renderStrongMarkdown
	ret.rendererFuncs[NodeBlockquote] = ret.renderBlockquoteMarkdown
	ret.rendererFuncs[NodeHeading] = ret.renderHeadingMarkdown
	ret.rendererFuncs[NodeList] = ret.renderListMarkdown
	ret.rendererFuncs[NodeListItem] = ret.renderListItemMarkdown
	ret.rendererFuncs[NodeThematicBreak] = ret.renderThematicBreakMarkdown
	ret.rendererFuncs[NodeHardBreak] = ret.renderHardBreakMarkdown
	ret.rendererFuncs[NodeSoftBreak] = ret.renderSoftBreakMarkdown
	ret.rendererFuncs[NodeHTMLBlock] = ret.renderHTMLMarkdown
	ret.rendererFuncs[NodeInlineHTML] = ret.renderInlineHTMLMarkdown
	ret.rendererFuncs[NodeLink] = ret.renderLinkMarkdown
	ret.rendererFuncs[NodeImage] = ret.renderImageMarkdown

	// 注册 GFM 渲染函数

	ret.rendererFuncs[NodeStrikethrough] = ret.renderStrikethroughMarkdown
	ret.rendererFuncs[NodeTaskListItemMarker] = ret.renderTaskListItemMarkerMarkdown
	ret.rendererFuncs[NodeTable] = ret.renderTableMarkdown
	ret.rendererFuncs[NodeTableHead] = ret.renderTableHeadMarkdown
	ret.rendererFuncs[NodeTableRow] = ret.renderTableRowMarkdown
	ret.rendererFuncs[NodeTableCell] = ret.renderTableCellMarkdown

	return
}

// TODO: 表的格式化应该按最宽的单元格对齐内容

func (r *Renderer) renderTableCellMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.writeByte('|')
	}
	return WalkContinue, nil
}

func (r *Renderer) renderTableRowMarkdown(node Node, entering bool) (WalkStatus, error) {
	if !entering {
		r.writeString("|\n")
	}
	return WalkContinue, nil
}

func (r *Renderer) renderTableHeadMarkdown(node Node, entering bool) (WalkStatus, error) {
	if !entering {
		r.writeString("|\n")
		n := node.(*TableHead)
		table := n.Parent().(*Table)
		for i := 0; i < len(table.Aligns); i++ {
			align := table.Aligns[i]
			switch align {
			case 0:
				r.writeString("|---")
			case 1:
				r.writeString("|:---")
			case 2:
				r.writeString("|:---:")
			case 3:
				r.writeString("|---:")
			}
		}
		r.writeString("|\n")
	}
	return WalkContinue, nil
}

func (r *Renderer) renderTableMarkdown(node Node, entering bool) (WalkStatus, error) {
	if !entering {
		r.writeByte('\n')
	}
	return WalkContinue, nil
}

func (r *Renderer) renderStrikethroughMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.writeString("~~")
	} else {
		r.writeString("~~")
	}
	return WalkContinue, nil
}

func (r *Renderer) renderImageMarkdown(node Node, entering bool) (WalkStatus, error) {
	n := node.(*Image)
	if entering {
		r.writeString("![")
		r.write(n.firstChild.Tokens())
		r.writeString("](")
		r.write(n.Destination)
		if nil != n.Title {
			r.writeString(" \"")
			r.write(n.Title)
			r.writeByte('"')
		}
		r.writeByte(')')
	}
	return WalkContinue, nil
}

func (r *Renderer) renderLinkMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		n := node.(*Link)
		r.writeString("[")
		r.write(n.firstChild.Tokens()) // FIXME: 未解决链接嵌套，另外还需要考虑链接引用定义
		r.writeString("](")
		r.write(n.Destination)
		if nil != n.Title {
			r.writeString(" \"")
			r.write(n.Title)
			r.writeByte('"')
		}
		r.writeByte(')')
	}
	return WalkContinue, nil
}

func (r *Renderer) renderHTMLMarkdown(node Node, entering bool) (WalkStatus, error) {
	if !entering {
		return WalkContinue, nil
	}

	r.newline()
	r.write(node.Tokens())
	r.newline()
	return WalkContinue, nil
}

func (r *Renderer) renderInlineHTMLMarkdown(node Node, entering bool) (WalkStatus, error) {
	if !entering {
		return WalkContinue, nil
	}

	r.write(node.Tokens())
	return WalkContinue, nil
}

func (r *Renderer) renderDocumentMarkdown(node Node, entering bool) (WalkStatus, error) {
	return WalkContinue, nil
}

func (r *Renderer) renderParagraphMarkdown(node Node, entering bool) (WalkStatus, error) {
	listPadding := 0
	inList := false
	inTightList := false
	lastListItemLastPara := false
	if parent := node.Parent(); nil != parent {
		if listItem, ok := parent.(*ListItem); ok { // ListItem.Paragraph
			inList = true

			// 必须通过列表（而非列表项）上的紧凑标识判断，因为在设置该标识时仅设置了 List.tight
			// 设置紧凑标识的具体实现可参考函数 List.Finalize()
			inTightList = listItem.Parent().(*List).tight

			firstPara := listItem.firstChild
			if 3 != listItem.listData.typ { // 普通列表
				if firstPara != node {
					listPadding = listItem.padding
				}
			} else { // 任务列表
				if firstPara.Next() != node { // 任务列表要跳过 TaskListItemMarker 即 [X]
					listPadding = listItem.padding
				}
			}

			nextItem := listItem.next
			if nil == nextItem {
				nextPara := node.Next()
				lastListItemLastPara = nil == nextPara
			}
		}
	}

	if entering {
		r.write(bytes.Repeat([]byte{itemSpace}, listPadding))
	} else {
		r.newline()
		if !inList {
			r.writeByte('\n')
		} else {
			if !inTightList || lastListItemLastPara {
				r.writeByte('\n')
			}
		}
	}
	return WalkContinue, nil
}

func (r *Renderer) renderTextMarkdown(node Node, entering bool) (WalkStatus, error) {
	if !entering {
		return WalkContinue, nil
	}

	if typ := node.Parent().Type(); NodeLink != typ && NodeImage != typ {
		r.write(node.Tokens())
	}
	return WalkContinue, nil
}

func (r *Renderer) renderCodeSpanMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.writeByte('`')
		r.write(node.Tokens())
		return WalkSkipChildren, nil
	}

	r.writeByte('`')
	return WalkContinue, nil
}

func (r *Renderer) renderCodeBlockMarkdown(node Node, entering bool) (WalkStatus, error) {
	n := node.(*CodeBlock)
	if !n.isFenced {
		n.fenceLength = 3
	}
	if entering {
		listPadding := 0
		if grandparent := node.Parent().Parent(); nil != grandparent {
			if list, ok := grandparent.(*List); ok { // List.ListItem.CodeBlock
				if node.Parent().FirstChild() != node {
					listPadding = list.padding
				}
			}
		}

		r.newline()
		if 0 < listPadding {
			r.write(bytes.Repeat([]byte{itemSpace}, listPadding))
		}
		r.write(bytes.Repeat([]byte{itemBacktick}, n.fenceLength))
		r.write(n.info)
		r.writeByte('\n')
		if 0 < listPadding {
			lines := bytes.Split(n.tokens, []byte{itemNewline})
			length := len(lines)
			for i, line := range lines {
				r.write(bytes.Repeat([]byte{itemSpace}, listPadding))
				r.write(line)
				if i < length-1 {
					r.writeByte('\n')
				}
			}
		} else {
			r.write(n.tokens)
		}
		return WalkSkipChildren, nil
	}

	r.write(bytes.Repeat([]byte{itemBacktick}, n.fenceLength))
	r.writeString("\n\n")
	return WalkContinue, nil
}

func (r *Renderer) renderEmphasisMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.writeByte('*')
	} else {
		r.writeByte('*')
	}
	return WalkContinue, nil
}

func (r *Renderer) renderStrongMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.writeString("**")
	} else {
		r.writeString("**")
	}
	return WalkContinue, nil
}

func (r *Renderer) renderBlockquoteMarkdown(n Node, entering bool) (WalkStatus, error) {
	if entering {
		r.newline()
		r.writeString("> ") // 带个空格更好一些
	} else {
		r.newline()
	}
	return WalkContinue, nil
}

func (r *Renderer) renderHeadingMarkdown(node Node, entering bool) (WalkStatus, error) {
	n := node.(*Heading)
	if entering {
		r.write(bytes.Repeat([]byte{itemCrosshatch}, n.Level)) // 统一使用 ATX 标题，不使用 Setext 标题
		r.writeByte(itemSpace)
	} else {
		r.newline()
		r.writeByte(itemNewline)
	}
	return WalkContinue, nil
}

func (r *Renderer) renderListMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.listLevel++
	} else {
		n := node.(*List)
		if n.tight {
			r.newline()
		}
		r.listLevel--
	}
	return WalkContinue, nil
}

func (r *Renderer) renderListItemMarkdown(node Node, entering bool) (WalkStatus, error) {
	n := node.(*ListItem)
	if entering {
		r.newline()
		if 1 < r.listLevel {
			parent := n.Parent().Parent().(*ListItem)
			r.write(bytes.Repeat([]byte{itemSpace}, len(parent.marker)+1))
			if 1 == parent.listData.typ {
				r.writeByte(' ') // 有序列表需要加上分隔符 . 或者 ) 的一个字符长度
			}
		}
		if 1 == n.listData.typ {
			r.writeString(strconv.Itoa(n.num) + ".")
		} else {
			r.write(n.marker)
		}
		r.writeByte(' ')
	}
	return WalkContinue, nil
}

func (r *Renderer) renderTaskListItemMarkerMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		n := node.(*TaskListItemMarker)
		r.writeByte('[')
		if n.checked {
			r.writeByte('X')
		} else {
			r.writeByte(' ')
		}
		r.writeByte(']')
	}
	return WalkContinue, nil
}

func (r *Renderer) renderThematicBreakMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.newline()
		r.writeString("---\n\n")
	}
	return WalkContinue, nil
}

func (r *Renderer) renderHardBreakMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		if !r.option.SoftBreak2HardBreak {
			r.writeString("\\\n")
		} else {
			r.writeString("\n")
		}
	}
	return WalkContinue, nil
}

func (r *Renderer) renderSoftBreakMarkdown(node Node, entering bool) (WalkStatus, error) {
	if entering {
		r.newline()
	}
	return WalkContinue, nil
}
