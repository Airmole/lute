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

// CodeSpan 描述了代码节点结构。
type CodeSpan struct {
	*BaseNode
}

func (t *Tree) parseCodeSpan(ctx *InlineContext) (ret Node) {
	startPos := ctx.pos
	n := 0
	for ; startPos+n < ctx.tokensLen; n++ {
		if itemBacktick != ctx.tokens[startPos+n] {
			break
		}
	}

	backticks := ctx.tokens[startPos : startPos+n]
	if ctx.tokensLen <= startPos+n {
		ctx.pos += n
		ret = &Text{tokens: backticks}
		return
	}

	endPos := t.matchCodeSpanEnd(ctx.tokens[startPos+n:], n)
	if 1 > endPos {
		ctx.pos += n
		ret = &Text{tokens: backticks}
		return
	}
	endPos = startPos + endPos + n

	textTokens := ctx.tokens[startPos+n : endPos]
	textTokens.replaceAll(itemNewline, itemSpace)
	if 2 < len(textTokens) && itemSpace == textTokens[0] && itemSpace == textTokens[len(textTokens)-1] && !textTokens.isBlankLine() {
		// 如果首尾是空格并且整行不是空行时剔除首尾的一个空格
		textTokens = textTokens[1:]
		textTokens = textTokens[:len(textTokens)-1]
	}

	ret = &CodeSpan{&BaseNode{typ: NodeCodeSpan, tokens: textTokens}}
	ctx.pos = endPos + n
	return
}

func (t *Tree) matchCodeSpanEnd(tokens items, num int) (pos int) {
	length := len(tokens)
	for pos < length {
		l := tokens[pos:].accept(itemBacktick)
		if num == l {
			next := pos + l
			if length-1 > next && itemBacktick == tokens[next] {
				continue
			}
			return pos
		}
		if 0 < l {
			pos += l
		} else {
			pos++
		}
	}
	return -1
}
