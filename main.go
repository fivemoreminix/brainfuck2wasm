package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
var watOnly = flag.Bool("c", false, "don't generate a WebAssembly module using wat2wasm")

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
	}
	ext := path.Ext(*outFile)
	*outFile = (*outFile)[:len(*outFile)-len(ext)]

	watPath := *outFile + ".wat"
	watFile, err := os.Create(watPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: unable to open %s", watPath)
		os.Exit(1)
	}

	var g = NewGenerator()
	fmt.Fprintf(watFile, watOutputFormat, g.genInstrs(string(bytes)))
	watFile.Close()

	if !*watOnly {
		wasmPath := *outFile + ".wasm"
		cmd := exec.Command("wat2wasm", watPath, "-o", wasmPath)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			switch typ := err.(type) {
			case *exec.ExitError:
				if !typ.Success() {
					fmt.Fprintln(os.Stderr, "warning: wat2wasm failed to compile the .wat file. Please file an issue."+
						"at https://github.com/fivemoreminix/brainfuck2wasm/issues")
					fmt.Fprintln(os.Stderr, "wat2wasm exit code: %v", typ.ExitCode())
					return
				} else {
					fmt.Println("Compiled module using wat2wasm")
				}
			default:
				fmt.Fprintf(os.Stderr, "warning: unable to run command 'wat2wasm %s'.\n"+
					"Are you sure wat2wasm is available in your PATH?\n\n", watPath)
				fmt.Fprintln(os.Stderr, "If you didn't intend to use wat2wasm, pass the -c flag. See help for more info.")
			}
			return
		}
		err = os.Remove(watPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: error deleting temporary .wat: %v", err)
			return
		}
	}
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
