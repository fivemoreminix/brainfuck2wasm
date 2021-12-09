# brainfuck2wasm
A Brainfuck to WebAssembly compiler written in Go.

## Usage
In order to start developing for the web with Brainfuck, you will need:
 * The `wat2wasm` tool from the [WebAssembly Binary Toolkit](https://github.com/WebAssembly/wabt) which on Windows is just extracting a release to C:\Program Files\wabt\ and adding the bin directory to your PATH.
 * A Go compiler
 * A modern web browser

Clone this repo, open a terminal, and then `cd` to the repo directory.

To compile a sample:
```
go run . -out hello-world.wat ./samples/hello-world.b
```

You can also build the compiler to a native executable:
```
go build .
brainfuck2wasm -out mandelbrot.wat ./samples/mandelbrot.b
```

After compiling a Brainfuck sample to .wat, you're going to need to compile the .wat (WebAssembly text) to a .wasm (WebAssembly) which is actually interpreted in web browsers.

```
wat2wasm -o brainfuck.wasm mandelbrot.wat
```

This generates a binary file called brainfuck.wasm which you cannot read but the browser can. It is also called "brainfuck.wasm" for a reason -- the javascript code is written to find the file named "brainfuck.wasm" in the same folder as itself.

To run this example, you cannot simply open index.html in a browser unfortunately. But it's easy to start a webserver if you have Python installed:

```
python -m http.server
```

Now open your browser and enter the address `localhost:8000` which will get you to your directory being served from your Python webserver.

Immediately, the brainfuck program will be running in the background. Mandelbrot takes a while because a) it is not a simple brainfuck program, and b) I didn't say this was an *optimizing* compiler.

Takes about 5 seconds for me. Then it will just print almost instantaneously.

Anyone is free to fork this project and make it optimized! There's a lot of opportunity for usability improvements and decreasing the number of generated instructions.

This project was for a pay-wall Medium story that I am still writing.
