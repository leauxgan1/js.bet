

// const audio = document.querySelector("#audio-players");

// const players = {
//   "attack": audio.querySelector("#attack"),
//   "block": audio.querySelector("#block"),
//   "crit": audio.querySelector("#crit"),
//   "dodge": audio.querySelector("#dodge"),
//   "winner": audio.querySelector("#winner"),
// };

// window.addEventListener("htmx:after:swap", (event) => {
//   const sides = event.detail.ctx.target.querySelector("#fighter-sides")
//   const audiosToPlay = sides.dataset.audios;
//   console.log(audiosToPlay);
//   if (!audiosToPlay) {
//     return;
//   }

//   const names = audiosToPlay.split(",");
//   for (const name of names) {
//     const foundPlayer = players[name];
//     if (!foundPlayer) continue;
//     console.log("playing: " + name);
//     foundPlayer.play();
//   }
// });
