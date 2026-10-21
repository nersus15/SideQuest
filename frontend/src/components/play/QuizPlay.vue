<script setup>
// Guess, then see the answer. Wrong guesses are part of the fun: same XP,
// a friendly "almost", and a short explanation.
import { computed, nextTick, ref } from 'vue';
import { playFor } from '../../lib/play';
import { questExplain, questPrompts, t } from '../../lib/i18n';
import { play as sound } from '../../lib/sound';

const props = defineProps({ quest: { type: Object, required: true } });
const s = computed(() => playFor(props.quest));
const options = computed(() => questPrompts(props.quest));
const right = computed(() => (props.quest.interaction?.answer ?? 1) - 1);
const answered = computed(() => s.value.quiz >= 0);
const correct = computed(() => s.value.quiz === right.value);
const LETTERS = ['A', 'B', 'C', 'D'];
const result = ref(null);

async function pick(i) {
  if (answered.value) return;
  s.value.quiz = i;
  sound(i === right.value ? 'ding' : 'breath');
  await nextTick();
  result.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}
</script>

<template>
  <div class="quiz">
    <button
      v-for="(o, i) in options"
      :key="i"
      class="opt"
      :class="{
        right: answered && i === right,
        wrong: answered && i === s.quiz && !correct,
        dim: answered && i !== right && i !== s.quiz,
      }"
      :disabled="answered"
      @click="pick(i)"
    >
      <span class="opt__letter">{{ answered && i === right ? '✓' : answered && i === s.quiz ? '✗' : LETTERS[i] }}</span>
      <span class="opt__text">{{ o }}</span>
    </button>

    <Transition name="reveal">
      <div v-if="answered" ref="result" class="result" :class="correct ? 'result--right' : 'result--wrong'">
        <strong>{{ correct ? t('play.quizRight') : t('play.quizWrong', { a: LETTERS[right] }) }}</strong>
        <p v-if="questExplain(quest)">{{ questExplain(quest) }}</p>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.quiz {
  display: grid;
  gap: 8px;
}
.opt {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  padding: 9px 12px;
  border-radius: 14px;
  background: rgb(255 255 255 / 0.04);
  border: 1.5px solid var(--line);
  text-align: left;
  transition: transform 0.2s var(--spring), border-color 0.2s, background 0.2s, opacity 0.2s;
}
.opt:not(:disabled):hover {
  background: rgb(255 255 255 / 0.07);
}
.opt:not(:disabled):active {
  transform: scale(0.98);
}
.opt:disabled {
  cursor: default;
}
.opt__letter {
  display: grid;
  place-items: center;
  flex: none;
  width: 26px;
  height: 26px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 900;
  color: var(--muted);
  background: rgb(255 255 255 / 0.08);
}
.opt__text {
  font-size: 14px;
  font-weight: 700;
}
.opt.right {
  border-color: var(--c-chill);
  background: rgb(103 226 189 / 0.12);
}
.opt.right .opt__letter {
  color: var(--night);
  background: var(--c-chill);
}
.opt.wrong {
  border-color: var(--coral);
  background: rgb(255 138 122 / 0.1);
}
.opt.wrong .opt__letter {
  color: var(--night);
  background: var(--coral);
}
.opt.dim {
  opacity: 0.45;
}
.result {
  margin-top: 4px;
  padding: 10px 12px;
  border-radius: 14px;
  font-size: 13px;
  line-height: 1.4;
}
.result strong {
  display: block;
  font-size: 14px;
  font-weight: 900;
}
.result p {
  margin: 4px 0 0;
  color: var(--muted);
  font-weight: 600;
}
.result--right {
  background: rgb(103 226 189 / 0.1);
}
.result--right strong {
  color: var(--c-chill);
}
.result--wrong {
  background: rgb(255 203 87 / 0.08);
}
.result--wrong strong {
  color: var(--gold);
}
.reveal-enter-active {
  transition: opacity 0.25s, transform 0.35s var(--spring);
}
.reveal-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
</style>
