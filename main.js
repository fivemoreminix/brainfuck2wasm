var memory = new WebAssembly.Memory({initial: 1}); // Allocate 1 64KB page of memory
const cellptr = new WebAssembly.Global({value: 'i32', mutable: true})

const container = document.getElementById('container');
const input = document.getElementById('input');
var submittedInput = "Test input.\0";
var inputCharIdx = 0;

function keyPress(event) {
    if (event.keyCode === 13) {
        alert("Started running");
        submittedInput = input.value + '\0';
        input.disabled = true;
        WebAssembly.instantiateStreaming(fetch('brainfuck.wasm'), importObject)
            .then(({instance}) => {
                instance.exports.runBrainfuck();
                // console.log(new Uint8Array(memory.buffer, 0, 8)); // Print first 8 cells of memory
            });
    }
}

var importObject = {
    console: {
        putChar: function(ch) {
            container.innerHTML += "&#" + ch + ";";
        },
        getChar: function() {
            const result = submittedInput.charCodeAt(inputCharIdx);
            if (inputCharIdx + 1 < submittedInput.length) {
                inputCharIdx++;
            }
            return result;
        }
    },
    js: {
        mem: memory,
        cellptr,
    }
};
