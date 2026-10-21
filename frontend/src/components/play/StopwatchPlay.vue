<script setup>
// "How long can you…?" One big button starts and stops; the player's best
// time sits underneath, and beating it gets a little celebration.
import { computed, onBeforeUnmount, ref } from 'vue';
import { playFor } from '../../lib/play';
import { formatTime, t } from '../../lib/i18n';
import { play as sound } from '../../lib/sound';

const props = defineProps({
  quest: { type: Object, required: true },
  best: { type: Number, default: 0 }, // personal best in ms
});
const s = computed(() => playFor(props.quest));

const running = ref(false);
const now = ref(0);
let started = 0;
let raf = 0;

function frame() {
  now.value = Date.now() - started;
  raf = requestAnimationFrame(frame);
}
function toggle() {
  if (running.value) {
    cancelAnimationFrame(raf);
    running.value = false;
    s.value.millis = Date.now() - started;
    sound(s.value.millis > props.best ? 'ding' : 'tick');
    return;
  }
  started = Date.now();
  now.value = 0;
  running.value = true;
  sound('tick');
  raf = requestAnimationFrame(frame);
}
onBeforeUnmount(() => cancelAnimationFrame(raf));

const shown = computed(() => (running.value ? now.value : s.value.millis));
const newBest = computed(() => !running.value && s.value.millis > 0 && s.value.millis > props.best);
</script>

<template>
  <div class="sw">
    <button class="sw__btn" :class="{ 'sw__btn--stop': running }" @click="toggle">
      {{ running ? t('play.swStop') : s.millis ? t('play.swAgain') : t('play.swStart') }}
    </button>
    <div class="sw__side">
      <span class="sw__time" :class="{ 'sw__time--best': newBest }">{{ formatTime(shown) }}</span>
      <span class="sw__note">
        <template v-if="running">{{ t('play.swGo') }}</template>
        <template v-else-if="newBest">{{ t('play.swNewBest') }}</template>
        <template v-else-if="best">🏅 {{ t('play.swBest', { t: formatTime(best) }) }}</template>
        <template v-else>{{ t('play.swNoBest') }}</template>
      </span>
    </div>
  </div>
</template>

<style scoped>
.sw {
  display: flex;
  align-items: center;
  gap: 16px;
}
.sw__btn {
  display: grid;
  place-items: center;
  flex: none;
  width: 82px;
  height: 82px;
  border-radius: 50%;
  font-size: 14px;
  font-weight: 900;
  color: var(--night);
  background: radial-gradient(circle at 35% 30%, color-mix(in srgb, var(--accent, var(--violet)) 70%, white), var(--accent, var(--violet)));
  box-shadow: 0 4px 0 color-mix(in srgb, var(--accent, var(--violet)) 60%, black);
  transition: transform 0.15s, box-shadow 0.15s, background 0.2s;
}
.sw__btn:active {
  transform: translateY(3px);
  box-shadow: 0 1px 0 color-mix(in srgb, var(--accent, var(--violet)) 60%, black);
}
.sw__btn--stop {
  background: radial-gradient(circle at 35% 30%, #ffb3a6, var(--coral));
  box-shadow: 0 4px 0 #c0584a;
  animation: pulse 1.2s ease-in-out infinite;
}
.sw__side {
  display: grid;
  gap: 4px;
  min-width: 0;
}
.sw__time {
  font-size: 30px;
  font-weight: 1000;
  line-height: 1.1;
  font-variant-numeric: tabular-nums;
  transition: color 0.3s;
}
.sw__time--best {
  color: var(--gold);
  animation: pop 0.45s var(--spring);
}
.sw__note {
  font-size: 12.5px;
  font-weight: 800;
  color: var(--muted);
}
@keyframes pulse {
  50% { transform: scale(1.04); }
}
@keyframes pop {
  40% { transform: scale(1.12); }
}
</style>
