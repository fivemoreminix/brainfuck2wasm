package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const watOutputFormat = `(module
 (import "js" "mem" (memory 1))
 (import "console" "putChar" (func $putChar (param i32)))
 (import "console" "getChar" (func $getChar (result i32)))

 (export "runBrainfuck" (func $runBrainfuck))
 (func $runBrainfuck (local $ptr i32)
%s )
)
`

var outFile = flag.String("o", "", "output file")

func main() {
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "error: expected an input file")
		os.Exit(1)
	}
	bytes, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: failed to read file")
		os.Exit(1)
	}

	if *outFile == "" {
		// The outFile is the input file with a .wat extension instead
		*outFile = flag.Arg(0)
		ext := path.Ext(*outFile)
		*outFile = strings.Replace(*outFile, ext, ".wat", 1)
	}

	f, err := os.Create(*outFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: unable to open outfile")
		os.Exit(1)
	}
	defer f.Close()

	var g = NewGenerator()
	fmt.Fprintf(f, watOutputFormat, g.genInstrs(string(bytes)))
}

type Generator struct {
	labelIdx           int
	loopDepth          int
	loopDepthLabelIdxs map[int]int
	indentLevel        int

	ptrMovedAmt int
	incDecAmt   int
}

func NewGenerator() *Generator {
	return &Generator{loopDepthLabelIdxs: make(map[int]int)}
}

func (g *Generator) genMove(dir int) string {
	indent := indent(g.indentLevel)
	if dir >= 0 {
		return fmt.Sprintf("%s(i32.add (local.get $ptr) (i32.const %d))\n%[1]s(local.set $ptr)\n", indent, dir)
	} else {
		return fmt.Sprintf("%s(i32.sub (local.get $ptr) (i32.const %d))\n%[1]s(local.set $ptr)\n", indent, -dir)
	}
}

func (g *Generator) genIncDec(val int) string {
	indent := indent(g.indentLevel)
	if val >= 0 {
		return fmt.Sprintf("%s(local.get $ptr)\n%[1]s(i32.store8 (i32.add (i32.load8_u (local.get $ptr)) (i32.const %d)))\n", indent, val)
	} else {
		return fmt.Sprintf("%s(local.get $ptr)\n%[1]s(i32.store8 (i32.sub (i32.load8_u (local.get $ptr)) (i32.const %d)))\n", indent, -val)
	}
}

func (g *Generator) genMoveIfNeeded() (out string) {
	if g.ptrMovedAmt != 0 {
		out = g.genMove(g.ptrMovedAmt)
		g.ptrMovedAmt = 0
	}
	return
}

func (g *Generator) genIncDecIfNeeded() (out string) {
	if g.incDecAmt != 0 {
		out = g.genIncDec(g.incDecAmt)
		g.incDecAmt = 0
	}
	return
}

func (g *Generator) genInstrs(code string) (out string) {
	g.indentLevel = 2
	for _, r := range code {
		indentS := indent(g.indentLevel)
		switch r {
		case '>':
			out += g.genIncDecIfNeeded()
			g.ptrMovedAmt++
		case '<':
			out += g.genIncDecIfNeeded()
			g.ptrMovedAmt--
		case '+':
			out += g.genMoveIfNeeded()
			g.incDecAmt++
		case '-':
			out += g.genMoveIfNeeded()
			g.incDecAmt--
		case '.':
			out += g.genMoveIfNeeded()
			out += g.genIncDecIfNeeded()

			out += fmt.Sprintf("%s(i32.load8_u (local.get $ptr))\n%[1]s(call $putChar)\n", indentS)
		case ',':
			out += g.genMoveIfNeeded()
			out += g.genIncDecIfNeeded()

			out += fmt.Sprintf("%s(i32.store8 (local.get $ptr) (call $getChar))\n", indentS)
		case '[':
			out += g.genMoveIfNeeded()
			out += g.genIncDecIfNeeded()

			out += fmt.Sprintf("%s(block $label$%[2]d\n", indentS, g.labelIdx)
			blockIndentS := indent(1 + g.indentLevel)
			out += fmt.Sprintf("%s(br_if $label$%[2]d (i32.eq (i32.load8_u (local.get $ptr)) (i32.const 0)))\n", blockIndentS, g.labelIdx)

			g.labelIdx++ // Block and Loop sections use different labels
			out += fmt.Sprintf("%s(loop $label$%d\n", blockIndentS, g.labelIdx)
			g.loopDepthLabelIdxs[g.loopDepth] = g.labelIdx // loopDepthLabelIdxs keep records of the labels for Loop sections

			g.labelIdx++
			g.loopDepth++
			g.indentLevel += 2
		case ']':
			out += g.genMoveIfNeeded()
			out += g.genIncDecIfNeeded()

			g.loopDepth--
			loopIndentS := indentS                    // Parent block indent plus the loop's indent
			blockIndentS := indent(g.indentLevel - 1) // Just parent block indent
			baseIndentS := indent(g.indentLevel - 2)
			out += fmt.Sprintf("%s(br_if $label$%[4]d (i32.ne (i32.load8_u (local.get $ptr)) (i32.const 0)))\n%[2]s)\n%[3]s)\n", loopIndentS, blockIndentS, baseIndentS, g.loopDepthLabelIdxs[g.loopDepth])
			g.indentLevel -= 2 // We exited the Loop and Block sections
		}
	}

	out += g.genMoveIfNeeded()
	out += g.genIncDecIfNeeded()

	return
}

func indent(level int) string {
	return strings.Repeat(" ", level)
}
