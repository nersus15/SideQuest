<script setup>
// Let chance decide: a slot-machine style reel that slows down and lands on
// one option. Big readable text instead of a crowded wheel.
import { computed, onBeforeUnmount, ref } from 'vue';
import { playFor } from '../../lib/play';
import { questPrompts, t } from '../../lib/i18n';
import { play as sound } from '../../lib/sound';

const props = defineProps({ quest: { type: Object, required: true } });
const s = computed(() => playFor(props.quest));
const options = computed(() => questPrompts(props.quest));

const spinning = ref(false);
const shown = ref(s.value.spin);
let timer = 0;

function spin() {
  if (spinning.value) return;
  const n = options.value.length;
  spinning.value = true;
  s.value.spin = -1;
  let idx = shown.value < 0 ? 0 : shown.value;
  let steps = 16 + Math.floor(Math.random() * n);
  let delay = 55;
  const step = () => {
    idx = (idx + 1) % n;
    shown.value = idx;
    sound('tick');
    if (--steps <= 0) {
      spinning.value = false;
      s.value.spin = idx;
      sound('pop');
      return;
    }
    delay = Math.min(delay * 1.13, 340); // slow down towards the end
    timer = setTimeout(step, delay);
  };
  timer = setTimeout(step, delay);
}
onBeforeUnmount(() => clearTimeout(timer));

const landed = computed(() => !spinning.value && s.value.spin >= 0);
</script>

<template>
  <div class="spinner">
    <div class="reel" :class="{ 'reel--spinning': spinning, 'reel--landed': landed }" aria-live="polite">
      <Transition name="roll">
        <span v-if="shown >= 0" :key="`${shown}-${spinning}`" class="reel__text">{{ options[shown] }}</span>
        <span v-else key="idle" class="reel__idle">{{ t('play.spinIdle') }}</span>
      </Transition>
    </div>
    <button class="btn btn--soft spinner__btn" :disabled="spinning" @click="spin">
      <span class="dice" :class="{ 'dice--roll': spinning }" aria-hidden="true">🎲</span>
      {{ landed ? t('play.spinAgain') : t('play.spin') }}
    </button>
  </div>
</template>

<style scoped>
.spinner {
  display: grid;
  gap: 10px;
}
.reel {
  position: relative;
  display: grid;
  place-items: center;
  height: 64px;
  padding: 0 14px;
  border-radius: 16px;
  overflow: hidden;
  background: rgb(0 0 0 / 0.22);
  border: 1.5px dashed var(--line);
  text-align: center;
  transition: border-color 0.25s, background 0.25s;
}
.reel--spinning {
  border-style: solid;
  border-color: color-mix(in srgb, var(--accent, var(--violet)) 50%, transparent);
}
.reel--landed {
  border-style: solid;
  border-color: var(--accent, var(--violet));
  background: color-mix(in srgb, var(--accent, var(--violet)) 14%, rgb(0 0 0 / 0.2));
  animation: land 0.45s var(--spring);
}
.reel__text {
  font-size: 17px;
  font-weight: 900;
  line-height: 1.25;
}
.reel__idle {
  font-size: 13px;
  font-weight: 700;
  color: var(--faint);
}
.spinner__btn {
  justify-self: center;
  height: 38px;
  padding: 0 20px;
  font-size: 14px;
}
.dice {
  display: inline-block;
}
.dice--roll {
  animation: roll 0.4s linear infinite;
}
.roll-enter-active,
.roll-leave-active {
  transition: transform 0.06s linear, opacity 0.06s linear;
}
/* Old and new text slide at the same time, so fast spins never lag. */
.roll-leave-active {
  position: absolute;
}
.roll-enter-from {
  transform: translateY(14px);
  opacity: 0;
}
.roll-leave-to {
  transform: translateY(-14px);
  opacity: 0;
}
@keyframes land {
  40% { transform: scale(1.04); }
}
@keyframes roll {
  to { transform: rotate(360deg); }
}
</style>
