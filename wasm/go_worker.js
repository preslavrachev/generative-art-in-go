// hijack the console.log function to capture stdout
let oldLog = console.log;
console.log = (line) => {
  postMessage({
    log: line,
  });
};

self.importScripts("wasm_exec.js");
//console.log(self.Go);

// initialize the Go WASM glue
const go = new self.Go();

let mod, inst;
let result;
WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
  .then((result) => {
    mod = result.module;
    inst = result.instance;

    go.run(inst);
    console.log("WASM Loaded");
  })
  .catch((err) => {
    console.error(err);
  });

self.onmessage = async (msg) => {
  switch (msg.data.type) {
    case "call":
      //console.log("Message received: " + msg.data.call);
      console.log("Start Rendering");
      args = msg.data.args || [];
      await self[msg.data.func](...args);
      console.log("Stop Rendering");
      break;
    case "set":
      self[msg.data.prop] = msg.data.value;
      break;
    default:
      console.error("Unavailable message type!");
  }
};

function displayImage(buf) {
  let blob = new Blob([buf], { type: imageType });
  console.log("Bytes received" + blob);
  postMessage({
    image: blob,
  });
}

function trackProgress(progress) {
  postMessage({
    progress: progress,
  });
}
