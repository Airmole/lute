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
)

// CodeBlock 描述了代码块节点结构。
type CodeBlock struct {
	*BaseNode
	isFenced    bool
	fenceChar   byte
	fenceLength int
	fenceOffset int
	info        items
}

func (codeBlock *CodeBlock) Continue(context *Context) int {
	var ln = context.currentLine
	var indent = context.indent
	if codeBlock.isFenced {
		if indent <= 3 && codeBlock.isFencedCodeClose(ln[context.nextNonspace:], codeBlock.fenceChar, codeBlock.fenceLength) {
			// closing fence - we're at end of line, so we can return
			context.finalize(codeBlock)
			return 2
		} else {
			// skip optional spaces of fence offset
			var i = codeBlock.fenceOffset
			var token byte
			for i > 0 {
				token = ln.peek(context.offset)
				if itemSpace != token && itemTab != token {
					break
				}
				context.advanceOffset(1, true)
				i--
			}
		}
	} else { // 缩进代码块
		if indent >= 4 {
			context.advanceOffset(4, true)
		} else if context.blank {
			context.advanceNextNonspace()
		} else {
			return 1
		}
	}
	return 0
}

func (codeBlock *CodeBlock) Finalize(context *Context) {
	if codeBlock.isFenced {
		content := codeBlock.tokens
		length := len(content)
		if 1 > length {
			return
		}

		var i int
		var token byte
		for ; i < length; i++ {
			token = content[i]
			if itemNewline == token {
				break
			}
		}
		firstLine := content[:i]
		rest := content[i+1:]

		codeBlock.info = unescapeString(bytes.TrimSpace(firstLine))
		codeBlock.tokens = rest
	} else { // 缩进代码块
		codeBlock.tokens = codeBlock.tokens.replaceNewlineSpace()
	}
}

func (codeBlock *CodeBlock) AcceptLines() bool {
	return true
}

func (codeBlock *CodeBlock) CanContain(nodeType int) bool {
	return false
}

func (t *Tree) parseFencedCode() (ret *CodeBlock) {
	marker := t.context.currentLine[t.context.nextNonspace]
	if itemBacktick != marker && itemTilde != marker {
		return nil
	}

	fenceChar := marker
	fenceLength := 0
	for i := t.context.nextNonspace; i < t.context.currentLineLen && fenceChar == t.context.currentLine[i]; i++ {
		fenceLength++
	}

	if 3 > fenceLength {
		return nil
	}

	var info items
	infoTokens := t.context.currentLine[t.context.nextNonspace+fenceLength:]
	if itemBacktick == marker && bytes.Contains(infoTokens, []byte{itemBacktick}) {
		return nil // info 部分不能包含 `
	}
	info = bytes.TrimSpace(infoTokens)
	info = unescapeString(info)
	ret = &CodeBlock{&BaseNode{typ: NodeCodeBlock, tokens: make(items, 0, 256)},
		true, fenceChar, fenceLength, t.context.indent, info}

	return
}

func (codeBlock *CodeBlock) isFencedCodeClose(tokens items, openMarker byte, num int) bool {
	closeMarker := tokens[0]
	if closeMarker != openMarker {
		return false
	}
	if num > tokens.accept(closeMarker) {
		return false
	}
	tokens = bytes.TrimSpace(tokens)
	for _, token := range tokens {
		if token != openMarker {
			return false
		}
	}
	return true
}
