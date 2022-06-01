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
 (global $cellptr (import "js" "cellptr") (mut i32))

 (export "runBrainfuck" (func $runBrainfuck))
 (func $runBrainfuck
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
	fmt.Println(flag.Arg(0))
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

	fmt.Fprintf(f, watOutputFormat, generateInstructions(string(bytes)))
}

func generateInstructions(code string) (out string) {
	labelIdx := 0
	loopDepth := 0
	loopDepthLabelIdxs := make(map[int]int)
	indentLevel := 0
	for _, r := range code {
		indentS := indent(2 + indentLevel)
		switch r {
		case '>':
			out += fmt.Sprintf("%s(i32.add (global.get $cellptr) (i32.const 1))\n%[1]s(global.set $cellptr)\n", indentS)
		case '<':
			out += fmt.Sprintf("%s(i32.sub (global.get $cellptr) (i32.const 1))\n%[1]s(global.set $cellptr)\n", indentS)
		case '+':
			out += fmt.Sprintf("%s(global.get $cellptr)\n%[1]s(i32.store8 (i32.add (i32.load8_u (global.get $cellptr)) (i32.const 1)))\n", indentS)
		case '-':
			out += fmt.Sprintf("%s(global.get $cellptr)\n%[1]s(i32.store8 (i32.sub (i32.load8_u (global.get $cellptr)) (i32.const 1)))\n", indentS)
		case '.':
			out += fmt.Sprintf("%s(i32.load8_u (global.get $cellptr))\n%[1]s(call $putChar)\n", indentS)
		case ',':
			out += fmt.Sprintf("%s(i32.store8 (global.get $cellptr) (call $getChar))\n", indentS)
		case '[':
			out += fmt.Sprintf("%s(block $label$%[2]d\n", indentS, labelIdx)
			blockIndentS := indent(3 + indentLevel)
			out += fmt.Sprintf("%s(br_if $label$%[2]d (i32.eq (i32.load8_u (global.get $cellptr)) (i32.const 0)))\n", blockIndentS, labelIdx)

			labelIdx++ // Block and Loop sections use different labels
			out += fmt.Sprintf("%s(loop $label$%d\n", blockIndentS, labelIdx)
			loopDepthLabelIdxs[loopDepth] = labelIdx // loopDepthLabelIdxs keep records of the labels for Loop sections

			labelIdx++
			loopDepth++
			indentLevel += 2
		case ']':
			loopDepth--
			loopIndentS := indent(2 + indentLevel)      // Parent block indent plus the loop's indent
			blockIndentS := indent(2 + indentLevel - 1) // Just parent block indent
			out += fmt.Sprintf("%s(br_if $label$%[4]d (i32.ne (i32.load8_u (global.get $cellptr)) (i32.const 0)))\n%[2]s)\n%[3]s)\n", loopIndentS, blockIndentS, indent(indentLevel), loopDepthLabelIdxs[loopDepth])
			indentLevel -= 2 // We exited the Loop and Block sections
		}
	}
	return
}

func indent(level int) string {
	return strings.Repeat(" ", level)
}
