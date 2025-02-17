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

package test

import (
	"encoding/json"
	"fmt"
	"github.com/b3log/lute"
	"io/ioutil"
	"strconv"
	"testing"
)

type testcase struct {
	EndLine   int    `json:"end_line"`
	Section   string `json:"section"`
	HTML      string `json:"html"`
	Markdown  string `json:"markdown"`
	Example   int    `json:"example"`
	StartLine int    `json:"start_line"`
}

func TestSpec(t *testing.T) {
	bytes, err := ioutil.ReadFile("commonmark-spec.json")
	if nil != err {
		t.Fatalf("read spec test cases failed: " + err.Error())
	}

	var testcases []testcase
	if err = json.Unmarshal(bytes, &testcases); nil != err {
		t.Fatalf("read spec test caes failed: " + err.Error())
	}

	luteEngine := lute.New(lute.GFM(false),
		lute.SoftBreak2HardBreak(false),
		lute.CodeSyntaxHighlight(false),
		lute.AutoSpace(false),
	)

	for _, test := range testcases {
		testName := test.Section + " " + strconv.Itoa(test.Example)
		fmt.Println("Test [" + testName + "]")

		html, err := luteEngine.MarkdownStr(testName, test.Markdown)
		if nil != err {
			t.Fatalf("unexpected: %s", err)
		}

		if test.HTML != html {
			t.Fatalf("test case [%s] failed\nexpected\n\t%q\ngot\n\t%q\noriginal markdown text\n\t%q", testName, test.HTML, html, test.Markdown)
		}
	}
}
