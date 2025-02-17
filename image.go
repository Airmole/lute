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

// Image 描述了图片节点结构。
type Image struct {
	*BaseNode
	Destination items // 图片链接地址
	Title       items // 图片标题
}

// parseBang 解析 !，可能是图片标记开始 ![ 也可能是普通文本 !。
func (t *Tree) parseBang(ctx *InlineContext) (ret Node) {
	var startPos = ctx.pos
	ctx.pos++
	if ctx.pos < ctx.tokensLen && itemOpenBracket == ctx.tokens[ctx.pos] {
		ctx.pos++
		ret = &Text{tokens: toItems("![")}
		// 将图片开始标记入栈
		t.addBracket(ret, startPos+2, true, ctx)
		return
	}

	ret = &Text{tokens: toItems("!")}
	return
}
