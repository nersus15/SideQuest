<script setup>
// Pick one of 2–4 options. Big, friendly tiles; one tap and done.
import { computed } from 'vue';
import { playFor } from '../../lib/play';
import { questPrompts } from '../../lib/i18n';
import { play as sound } from '../../lib/sound';

const props = defineProps({ quest: { type: Object, required: true } });
const s = computed(() => playFor(props.quest));
const options = computed(() => questPrompts(props.quest));

function pick(i) {
  s.value.choice = i;
  sound('tick');
}
</script>

<template>
  <div class="choices" :class="{ 'choices--two': options.length === 2 }" role="radiogroup">
    <button
      v-for="(o, i) in options"
      :key="i"
      class="choice"
      :class="{ on: s.choice === i, dim: s.choice >= 0 && s.choice !== i }"
      role="radio"
      :aria-checked="s.choice === i"
      @click="pick(i)"
    >
      <span class="choice__text">{{ o }}</span>
      <span v-if="s.choice === i" class="choice__check" aria-hidden="true">✓</span>
    </button>
  </div>
</template>

<style scoped>
.choices {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.choice {
  position: relative;
  display: grid;
  place-items: center;
  min-height: 62px;
  padding: 10px 12px;
  border-radius: 16px;
  background: rgb(255 255 255 / 0.04);
  border: 1.5px solid var(--line);
  text-align: center;
  transition: transform 0.25s var(--spring), border-color 0.2s, background 0.2s, opacity 0.2s;
}
.choices--two .choice {
  min-height: 78px;
}
.choice:hover {
  background: rgb(255 255 255 / 0.07);
}
.choice:active {
  transform: scale(0.96);
}
.choice__text {
  font-size: 14px;
  font-weight: 800;
  line-height: 1.3;
}
.choice.on {
  border-color: var(--accent, var(--violet));
  background: color-mix(in srgb, var(--accent, var(--violet)) 16%, transparent);
  animation: pop 0.35s var(--spring);
}
.choice.dim {
  opacity: 0.5;
}
.choice__check {
  position: absolute;
  top: -7px;
  right: -7px;
  display: grid;
  place-items: center;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 900;
  color: var(--night);
  background: var(--accent, var(--violet));
}
@keyframes pop {
  40% { transform: scale(1.05); }
}
</style>
