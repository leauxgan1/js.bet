const audioCache = {};

async function preloadAudioBuffers() {
  const audioFiles = ['attack', 'block', 'crit', 'dodge', 'winner'];
  const audioContext = new (window.AudioContext || window.webkitAudioContext)();
  
  for (const name of audioFiles) {
    try {
      const response = await fetch(`./audio/${name}.wav`);
      const arrayBuffer = await response.arrayBuffer();
      const audioBuffer = await audioContext.decodeAudioData(arrayBuffer);
      audioCache[name] = audioBuffer;
      console.log(`Loaded: ${name}`);
    } catch (error) {
      console.error(`Failed to load ${name}:`, error);
    }
  }
  
  window.audioContext = audioContext;
}

function playAudioInstant(name) {
  if (!window.audioContext) {
    console.error("AudioContext not initialized");
    return;
  }
  
  const buffer = audioCache[name];
  if (!buffer) {
    console.warn(`Audio ${name} not loaded`);
    return;
  }
  
  if (window.audioContext.state === 'suspended') {
    window.audioContext.resume();
  }
  
  const source = window.audioContext.createBufferSource();
  source.buffer = buffer;
  source.connect(window.audioContext.destination);
  source.start(0); // Play immediately
  
  console.log(`Playing: ${name} (Web Audio)`);
}

window.addEventListener("htmx:after:swap", (event) => {
  const sides = event.detail.ctx.target.querySelector("#fighter-sides");
  const audiosToPlay = sides?.dataset?.audios;
  
  if (!audiosToPlay) return;
  
  const names = audiosToPlay.split(",");
  for (const name of names) {
    playAudioInstant(name.trim());
  }
});

preloadAudioBuffers();
