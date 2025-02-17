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
	"strings"
	"unicode/utf8"
	"unsafe"
)

// lexer 描述了词法分析器结构。
type lexer struct {
	input   items // 输入的文本字符数组
	length  int   // 输入的文本字符数组的长度
	offset  int   // 当前读取位置
	lineNum int   // 当前行号
	width   int   // 最新一个 token 的宽度（字节数）
}

// nextLine 返回下一行。
func (l *lexer) nextLine() (line items) {
	if l.offset >= l.length {
		return
	}

	var b, nb byte
	i := l.offset
	for ; i < l.length; i += l.width {
		b = l.input[i]
		if itemNewline == b {
			i++
			break
		} else if itemCarriageReturn == b {
			// 按照规范定义的 line ending (https://spec.commonmark.org/0.29/#line-ending) 处理 \r
			if i < l.length-1 {
				nb = l.input[i+1]
				if itemNewline == nb {
					l.input = append(l.input[:i], l.input[i+1:]...) // 移除 \r，依靠下一个的 \n 切行
					l.length--                                      // 重新计算总长
				}
			}
			i++
			break
		} else if '\u0000' == b {
			// 将 \u0000 替换为 \uFFFD https://spec.commonmark.org/0.29/#insecure-characters

			l.input = append(l.input, 0, 0)
			copy(l.input[i+2:], l.input[i:])
			// \uFFFD 的 UTF-8 编码为 \xEF\xBF\xBD 共三个字节
			l.input[i] = '\xEF'
			l.input[i+1] = '\xBF'
			l.input[i+2] = '\xBD'
			l.length += 2 // 重新计算总长
			l.width = 3
			continue
		}

		if utf8.RuneSelf <= b { // 说明占用多个字节
			_, l.width = utf8.DecodeRune(l.input[i:])
		} else {
			l.width = 1
		}
	}
	line = l.input[l.offset:i]
	l.offset = i
	l.lineNum++
	return
}

// newLexer 创建一个词法分析器。
func newLexer(input items) (ret *lexer) {
	ret = &lexer{}
	// 动态构造一次，因为后续有可能会对字节数组进行赋值
	// 不构造的话会报错 fatal error: fault
	builder := strings.Builder{}

	builder.Write(input)
	ret.input = items(builder.String())
	ret.length = len(ret.input)

	if 0 < ret.length && itemNewline != ret.input[ret.length-1] {
		// 以 \n 结尾预处理
		ret.input = append(ret.input, itemNewline)
		ret.length++
	}

	return
}

// fromItems 快速转换 items 为 string。
func fromItems(items items) string {
	return *(*string)(unsafe.Pointer(&items))
}

// toItems 快速转换 str 为 items。
func toItems(str string) items {
	x := (*[2]uintptr)(unsafe.Pointer(&str))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*items)(unsafe.Pointer(&h))
}
