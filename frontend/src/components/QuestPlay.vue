<script setup>
// The in-app part of a quest: short answers, a checklist, a timer (with an
// optional breathing guide), a tap counter or a small doodle canvas.
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { playFor } from '../lib/play';
import { questPrompts, t } from '../lib/i18n';
import { play } from '../lib/sound';

const props = defineProps({
  quest: { type: Object, required: true },
});

const s = computed(() => playFor(props.quest));
const kind = computed(() => props.quest.interaction?.kind ?? '');
const prompts = computed(() => questPrompts(props.quest));

// --- timer ---
const seconds = computed(() => props.quest.interaction?.seconds ?? 0);
const left = ref(seconds.value);
const running = ref(false);
let tick;
function startTimer() {
  clearInterval(tick);
  left.value = seconds.value;
  running.value = true;
  s.value.timerDone = false;
  const end = Date.now() + seconds.value * 1000;
  tick = setInterval(() => {
    left.value = Math.max(0, Math.ceil((end - Date.now()) / 1000));
    if (left.value === 0) {
      clearInterval(tick);
      running.value = false;
      s.value.timerDone = true;
      play('ding');
    }
  }, 200);
}
const elapsed = computed(() => seconds.value - left.value);
const ring = computed(() => (running.value || s.value.timerDone ? elapsed.value / seconds.value : 0));
// Box breathing: 4s in, 4s hold, 4s out, 4s hold.
const phase = computed(() => (running.value ? Math.floor(elapsed.value / 4) % 4 : -1));
const breath = computed(() => (phase.value < 0 ? null : ['in', 'hold', 'out', 'hold'][phase.value]));
// The circle grows while breathing in, stays big while holding, then shrinks.
const dial = computed(() => (phase.value < 0 || props.quest.interaction?.guide !== 'breath' ? '' : phase.value < 2 ? 'in' : 'out'));
const breathLabel = computed(() => (breath.value === 'in' ? t('play.breathIn') : breath.value === 'out' ? t('play.breathOut') : t('play.breathHold')));

function toggleCheck(i) {
  s.value.checked[i] = !s.value.checked[i];
  play('tick');
}

// A calm tone at each breath in and out.
watch(breath, (b, prev) => b && b !== prev && b !== 'hold' && play('breath'));

// --- counter ---
const target = computed(() => props.quest.interaction?.target ?? 0);
const bump = ref(0);
function count() {
  s.value.count = Math.min(999, s.value.count + 1);
  play('tick');
  bump.value++;
}

// --- doodle ---
const canvas = ref(null);
const COLORS = ['#f6f1ff', '#ffcb57', '#ff8fb8', '#67e2bd', '#6cc6ff'];
const pen = ref(COLORS[0]);
let ctx;
let drawing = false;
let last = null;

function setupCanvas() {
  const c = canvas.value;
  if (!c) return;
  const dpr = window.devicePixelRatio || 1;
  const { width, height } = c.getBoundingClientRect();
  c.width = Math.round(width * dpr);
  c.height = Math.round(height * dpr);
  ctx = c.getContext('2d');
  ctx.scale(dpr, dpr);
  ctx.lineCap = 'round';
  ctx.lineJoin = 'round';
  ctx.lineWidth = 3.5;
  // Bring back an earlier drawing for this quest.
  if (s.value.doodle) {
    const img = new Image();
    img.onload = () => ctx.drawImage(img, 0, 0, width, height);
    img.src = s.value.doodle;
  }
}

function point(e) {
  const r = canvas.value.getBoundingClientRect();
  return { x: e.clientX - r.left, y: e.clientY - r.top };
}
function down(e) {
  drawing = true;
  try {
    canvas.value.setPointerCapture(e.pointerId); // keep drawing past the edge
  } catch {
    /* not every pointer can be captured */
  }
  last = point(e);
  ctx.strokeStyle = pen.value;
  ctx.beginPath();
  ctx.arc(last.x, last.y, 1.2, 0, Math.PI * 2);
  ctx.stroke();
}
function move(e) {
  if (!drawing) return;
  const p = point(e);
  ctx.beginPath();
  ctx.moveTo(last.x, last.y);
  ctx.lineTo(p.x, p.y);
  ctx.stroke();
  last = p;
}
function up() {
  if (!drawing) return;
  drawing = false;
  save();
}
// Saved at 1x so journal entries stay small.
function save() {
  const c = canvas.value;
  const { width, height } = c.getBoundingClientRect();
  const out = document.createElement('canvas');
  out.width = Math.round(width);
  out.height = Math.round(height);
  out.getContext('2d').drawImage(c, 0, 0, out.width, out.height);
  s.value.doodle = out.toDataURL('image/png');
}
function clearDoodle() {
  const c = canvas.value;
  ctx.clearRect(0, 0, c.width, c.height);
  s.value.doodle = '';
}

onMounted(() => {
  if (kind.value === 'doodle') setupCanvas();
});
onBeforeUnmount(() => clearInterval(tick));
</script>

<template>
  <div class="play" :class="`play--${kind}`">
    <!-- answer -->
    <div v-if="kind === 'answer'" class="answers" :class="{ 'answers--grid': prompts.length >= 4 }">
      <input
        v-for="(p, i) in prompts"
        :key="i"
        v-model="s.answers[i]"
        class="answer__input"
        maxlength="140"
        :placeholder="p"
        :aria-label="p"
      />
    </div>

    <!-- checklist -->
    <ul v-else-if="kind === 'checklist'" class="checks">
      <li v-for="(p, i) in prompts" :key="i">
        <button class="check" :class="{ on: s.checked[i] }" :aria-pressed="s.checked[i]" @click="toggleCheck(i)">
          <span class="check__box" aria-hidden="true">{{ s.checked[i] ? '✓' : '' }}</span>
          {{ p }}
        </button>
      </li>
    </ul>

    <!-- timer -->
    <div v-else-if="kind === 'timer'" class="timer">
      <div class="timer__dial" :class="dial && `timer__dial--${dial}`">
        <svg viewBox="0 0 100 100" aria-hidden="true">
          <circle cx="50" cy="50" r="44" class="timer__track" />
          <circle cx="50" cy="50" r="44" class="timer__fill" :style="{ strokeDashoffset: 276.5 * (1 - ring) }" />
        </svg>
        <span class="timer__text">
          <template v-if="s.timerDone && !running">✓</template>
          <span v-else>{{ running ? left : seconds }}<small>s</small></span>
        </span>
      </div>
      <div class="timer__side">
        <p v-if="dial" class="timer__breath">{{ breathLabel }}</p>
        <p v-else-if="s.timerDone" class="timer__done">{{ t('play.timerDone') }}</p>
        <button v-if="!running" class="btn btn--soft timer__btn" @click="startTimer">
          {{ s.timerDone ? t('play.timerAgain') : t('play.timerStart') }}
        </button>
      </div>
    </div>

    <!-- counter -->
    <div v-else-if="kind === 'counter'" class="counter">
      <button :key="bump" class="counter__tap" @click="count">
        <strong>{{ s.count }}</strong>
        <span>{{ t('play.tap') }}</span>
      </button>
      <div class="counter__side">
        <div v-if="target" class="counter__dots" :aria-label="`${s.count}/${target}`">
          <i v-for="n in target" :key="n" :class="{ on: n <= s.count }" />
        </div>
        <button v-if="s.count" class="btn btn--ghost counter__undo" @click="s.count--">{{ t('play.undo') }}</button>
      </div>
    </div>

    <!-- doodle -->
    <div v-else-if="kind === 'doodle'" class="doodle">
      <canvas
        ref="canvas"
        class="doodle__canvas"
        :aria-label="t('play.drawHere')"
        @pointerdown="down"
        @pointermove="move"
        @pointerup="up"
        @pointercancel="up"
      />
      <span v-if="!s.doodle" class="doodle__hint" aria-hidden="true">{{ t('play.drawHere') }} ✏️</span>
      <div class="doodle__tools">
        <button
          v-for="c in COLORS"
          :key="c"
          class="swatch"
          :class="{ on: pen === c }"
          :style="{ background: c }"
          :aria-label="c"
          @click="pen = c"
        />
        <button class="btn btn--ghost doodle__clear" :disabled="!s.doodle" @click="clearDoodle">{{ t('play.clear') }}</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.play {
  margin-top: 4px;
  user-select: text;
  -webkit-user-select: text;
}

/* answers */
.answers {
  display: grid;
  gap: 6px;
}
.answers--grid {
  grid-template-columns: 1fr 1fr;
}
.answer__input {
  width: 100%;
  height: 34px;
  padding: 0 10px;
  border-radius: 10px;
  border: 1.5px solid var(--line);
  background: rgb(0 0 0 / 0.2);
  color: var(--ghost);
  font: inherit;
  font-size: 14px;
  font-weight: 700;
  outline: none;
  transition: border-color 0.2s;
}
.answer__input:focus {
  border-color: var(--accent, var(--violet));
}
.answer__input::placeholder {
  color: var(--faint);
  font-weight: 600;
}

/* checklist */
.checks {
  display: grid;
  gap: 4px;
  margin: 0;
  padding: 0;
  list-style: none;
}
.check {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  padding: 6px 8px;
  border-radius: 10px;
  text-align: left;
  font-size: 13.5px;
  font-weight: 700;
  color: var(--muted);
  transition: background 0.2s, color 0.2s;
}
.check:hover {
  background: rgb(255 255 255 / 0.04);
}
.check__box {
  display: grid;
  place-items: center;
  flex: none;
  width: 20px;
  height: 20px;
  border-radius: 6px;
  border: 1.5px solid var(--line);
  font-size: 12px;
  font-weight: 900;
  color: var(--night);
  transition: background 0.2s, border-color 0.2s, transform 0.25s var(--spring);
}
.check.on {
  color: var(--ghost);
}
.check.on .check__box {
  background: var(--accent, var(--violet));
  border-color: var(--accent, var(--violet));
  transform: scale(1.08);
}

/* timer */
.timer {
  display: flex;
  align-items: center;
  gap: 14px;
}
.timer__dial {
  position: relative;
  flex: none;
  width: 82px;
  height: 82px;
  transition: transform 4s ease-in-out;
}
.timer__dial--in {
  transform: scale(1.12);
}
.timer__dial--out {
  transform: scale(0.88);
}
.timer__dial svg {
  width: 100%;
  height: 100%;
  transform: rotate(-90deg);
}
.timer__track,
.timer__fill {
  fill: none;
  stroke-width: 8;
}
.timer__track {
  stroke: rgb(255 255 255 / 0.08);
}
.timer__fill {
  stroke: var(--accent, var(--violet));
  stroke-linecap: round;
  stroke-dasharray: 276.5;
  transition: stroke-dashoffset 0.25s linear;
}
.timer__text {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  font-size: 22px;
  font-weight: 1000;
  font-variant-numeric: tabular-nums;
}
.timer__text small {
  font-size: 12px;
  color: var(--faint);
}
.timer__side {
  display: grid;
  gap: 6px;
  justify-items: start;
}
.timer__breath {
  margin: 0;
  font-size: 17px;
  font-weight: 900;
  color: var(--accent, var(--violet));
}
.timer__done {
  margin: 0;
  font-weight: 900;
  color: var(--gold);
}
.timer__btn {
  height: 36px;
  font-size: 14px;
}

/* counter */
.counter {
  display: flex;
  align-items: center;
  gap: 14px;
}
.counter__tap {
  display: grid;
  place-items: center;
  align-content: center;
  flex: none;
  width: 82px;
  height: 82px;
  border-radius: 50%;
  background: radial-gradient(circle at 35% 30%, color-mix(in srgb, var(--accent, var(--violet)) 70%, white), var(--accent, var(--violet)));
  color: var(--night);
  box-shadow: 0 4px 0 color-mix(in srgb, var(--accent, var(--violet)) 60%, black);
  animation: tap-pop 0.25s var(--spring);
}
.counter__tap:active {
  transform: translateY(3px);
  box-shadow: 0 1px 0 color-mix(in srgb, var(--accent, var(--violet)) 60%, black);
}
.counter__tap strong {
  font-size: 26px;
  line-height: 1;
  font-weight: 1000;
}
.counter__tap span {
  font-size: 11px;
  font-weight: 900;
}
.counter__side {
  display: grid;
  gap: 6px;
  justify-items: start;
}
.counter__dots {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  max-width: 150px;
}
.counter__dots i {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: rgb(255 255 255 / 0.1);
  transition: background 0.2s, transform 0.25s var(--spring);
}
.counter__dots i.on {
  background: var(--gold);
  transform: scale(1.1);
}
.counter__undo {
  height: 28px;
  padding: 0 4px;
}

/* doodle */
.doodle {
  position: relative;
}
.doodle__canvas {
  display: block;
  width: 100%;
  height: 112px;
  border-radius: 14px;
  background: rgb(0 0 0 / 0.25);
  border: 1.5px dashed var(--line);
  touch-action: none;
  cursor: crosshair;
}
.doodle__hint {
  position: absolute;
  top: 44px;
  left: 0;
  right: 0;
  text-align: center;
  font-size: 13px;
  font-weight: 800;
  color: var(--faint);
  pointer-events: none;
}
.doodle__tools {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
}
.swatch {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid transparent;
  box-shadow: inset 0 0 0 1px rgb(0 0 0 / 0.2);
  transition: transform 0.2s var(--spring);
}
.swatch.on {
  border-color: var(--ghost);
  transform: scale(1.15);
}
.doodle__clear {
  height: 26px;
  margin-left: auto;
  padding: 0 6px;
}

@keyframes tap-pop {
  from { transform: scale(0.92); }
  to { transform: scale(1); }
}
</style>
