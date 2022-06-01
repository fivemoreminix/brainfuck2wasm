var memory = new WebAssembly.Memory({initial: 1}); // Allocate 1 64KB page of memory

const container = document.getElementById('container');
const input = document.getElementById('input');
var submittedInput = "Test input.\0";
var inputCharIdx = 0;

function keyPress(event) {
    if (event.keyCode === 13) {
        submittedInput = input.value + '\0';
        input.disabled = true;

        const runningText = document.getElementById("running-text");
        runningText.innerHTML = "running...";

        WebAssembly.instantiateStreaming(fetch('brainfuck.wasm'), importObject)
            .then(({instance}) => {
                instance.exports.runBrainfuck();
                // console.log(new Uint8Array(memory.buffer, 0, 8)); // Print first 8 cells of memory
                runningText.innerHTML = "";
            });
    }
}

var importObject = {
    console: {
        putChar: function(ch) {
            switch (ch) {
                case 10: ch = "<br>"; break;
                case 32: ch = "&nbsp;"; break;
                default: ch = String.fromCharCode(ch);
            }
            container.innerHTML += ch;
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
    }
};
