// Tiny sound effects, synthesised with Web Audio: no audio files, works
// offline, adds nothing to the app size. Every sound is short and soft.
import { reactive } from 'vue';

export const soundSettings = reactive({ volume: 60, muted: false, nudge: false });
// Set while a fullscreen app, game or presentation is in front.
export const soundQuiet = { value: false };

let ctx;
function audio() {
  if (typeof window === 'undefined') return null;
  try {
    ctx ??= new (window.AudioContext || window.webkitAudioContext)();
    if (ctx.state === 'suspended') ctx.resume();
    return ctx;
  } catch {
    return null; // no audio device: stay silent
  }
}

// Webviews only allow audio after a user gesture, so warm up on the first one.
if (typeof window !== 'undefined') {
  window.addEventListener('pointerdown', () => audio(), { once: true, capture: true });
}

// One soft note: frequency (optionally gliding to `to`), start offset and
// length in seconds.
function note(ac, out, freq, at, len, { type = 'sine', gain = 0.25, to } = {}) {
  const osc = ac.createOscillator();
  const env = ac.createGain();
  const t0 = ac.currentTime + at;
  osc.type = type;
  osc.frequency.setValueAtTime(freq, t0);
  if (to) osc.frequency.exponentialRampToValueAtTime(to, t0 + len);
  env.gain.setValueAtTime(0.0001, t0);
  env.gain.exponentialRampToValueAtTime(gain, t0 + 0.012);
  env.gain.exponentialRampToValueAtTime(0.0001, t0 + len);
  osc.connect(env).connect(out);
  osc.start(t0);
  osc.stop(t0 + len + 0.02);
}

const SOUNDS = {
  // Accepting a quest: a bubbly pop.
  pop: (ac, out) => note(ac, out, 420, 0, 0.12, { to: 880, type: 'triangle' }),
  // Ticking a box, tapping the counter, toggling a twist.
  tick: (ac, out) => note(ac, out, 1250, 0, 0.05, { type: 'square', gain: 0.07 }),
  // Reroll or skip: a quick soft swoosh.
  whoosh: (ac, out) => note(ac, out, 700, 0, 0.16, { to: 260, gain: 0.12 }),
  // Finishing a quest: a little two-note chime.
  chime: (ac, out) => {
    note(ac, out, 784, 0, 0.35, { gain: 0.2 });
    note(ac, out, 1175, 0.09, 0.45, { gain: 0.18 });
  },
  // Levelling up: a cheerful rising arpeggio.
  levelup: (ac, out) => {
    [523, 659, 784, 1047].forEach((f, i) => note(ac, out, f, i * 0.08, 0.32, { type: 'triangle', gain: 0.18 }));
    note(ac, out, 1568, 0.34, 0.5, { gain: 0.12 });
  },
  // A timer finished.
  ding: (ac, out) => note(ac, out, 988, 0, 0.6, { gain: 0.16 }),
  // Breathing guide: a low, calm tone on each breath in and out.
  breath: (ac, out) => note(ac, out, 330, 0, 0.9, { gain: 0.1 }),
  // Boo speaks up from the capsule (only if the player opted in).
  nudge: (ac, out) => {
    note(ac, out, 660, 0, 0.18, { gain: 0.12 });
    note(ac, out, 880, 0.14, 0.26, { gain: 0.1 });
  },
};

// play('chime'). Respects mute, volume and fullscreen quiet time.
export function play(name, { force = false } = {}) {
  if (!force && (soundSettings.muted || soundQuiet.value)) return;
  if (name === 'nudge' && !soundSettings.nudge && !force) return;
  const make = SOUNDS[name];
  const ac = make && audio();
  if (!ac) return;
  const out = ac.createGain();
  out.gain.value = Math.max(0.1, Math.min(1, soundSettings.volume / 100));
  out.connect(ac.destination);
  make(ac, out);
}
