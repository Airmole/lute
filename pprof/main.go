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

package main

import (
	"io/ioutil"
	"os"
	"runtime/pprof"

	"github.com/b3log/lute"
)

func main() {
	spec := "test/commonmark-spec"
	bytes, err := ioutil.ReadFile(spec + ".md")
	if nil != err {
		panic(err)
	}

	luteEngine := lute.New(lute.GFM(true),
		lute.CodeSyntaxHighlight(false),
		lute.SoftBreak2HardBreak(false),
		lute.AutoSpace(false),
		lute.FixTermTypo(false),
	)

	cpuProfile, _ := os.Create("pprof/cpu_profile")
	pprof.StartCPUProfile(cpuProfile)
	for i := 0; i < 100; i++ {
		_, err := luteEngine.Markdown("pprof "+spec, bytes)
		if nil != err {
			panic(err)
		}
	}
	pprof.StopCPUProfile()
}
