# brainfuck2wasm
A Brainfuck to WebAssembly compiler written in Go.

## Requirements
In order to start developing for the web with Brainfuck, you will need:
 * The `wat2wasm` tool from the [WebAssembly Binary Toolkit](https://github.com/WebAssembly/wabt) which on Windows is just extracting a release to C:\Program Files\wabt\ and adding the bin directory to your PATH.
 * A Go compiler
 * A modern web browser

Clone this repo, open a terminal, and then `cd` to the repo directory.

## Usage
To compile a sample to a WebAssembly module named "brainfuck.wasm":
```
go run . -o brainfuck ./samples/hello-world.b
```

Then you can host the website in this directory using Python:
```
python3 -m http.server
```

(Note: if `python3` isn't a valid program, try `python` instead)

Now go to `localhost:8000` in your web browser and wait for the program to run! Mandelbrot takes a while because it is
a very large program.

You can also install brainfuck2wasm on your computer and run anywhere:
```
go install
brainfuck2wasm -o mandelbrot.wat ./samples/mandelbrot.b
```

It is installed to go/bin in your user home directory.

## More in-depth usage examples
To compile a sample to WebAssembly Text format (.wat):
```
go run . -c -o hello-world.wat ./samples/hello-world.b
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

Takes about 5 seconds for me. Then it will print the results.

Anyone is free to fork this project and make it optimized! There's a lot of opportunity for usability improvements and decreasing the number of generated instructions.

This project was for a pay-wall Medium story which you may find here:
https://medium.com/gitconnected/compiling-brainf-to-webassembly-with-go-8838519e3c8b
