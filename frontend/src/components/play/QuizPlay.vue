<script setup>
// Guess until you get it. A wrong pick is marked and you try another; only
// the right answer completes the quest, then a short explanation shows.
import { computed, nextTick, ref } from 'vue';
import { playFor } from '../../lib/play';
import { questExplain, questPrompts, t } from '../../lib/i18n';
import { play as sound } from '../../lib/sound';

const props = defineProps({ quest: { type: Object, required: true } });
const s = computed(() => playFor(props.quest));
const options = computed(() => questPrompts(props.quest));
const right = computed(() => (props.quest.interaction?.answer ?? 1) - 1);
const solved = computed(() => s.value.quiz === right.value);
const LETTERS = ['A', 'B', 'C', 'D'];
const shake = ref(-1);
const result = ref(null);

async function pick(i) {
  if (solved.value || s.value.quizWrong.includes(i)) return;
  if (i === right.value) {
    s.value.quiz = i;
    sound('ding');
    await nextTick();
    result.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    return;
  }
  s.value.quizWrong = [...s.value.quizWrong, i];
  sound('breath');
  shake.value = i;
  setTimeout(() => (shake.value = -1), 400);
}

const tried = (i) => s.value.quizWrong.includes(i);
</script>

<template>
  <div class="quiz">
    <button
      v-for="(o, i) in options"
      :key="i"
      class="opt"
      :class="{
        right: solved && i === right,
        wrong: tried(i),
        shake: shake === i,
        dim: solved && i !== right && !tried(i),
      }"
      :disabled="solved || tried(i)"
      @click="pick(i)"
    >
      <span class="opt__letter">{{ solved && i === right ? '✓' : tried(i) ? '✗' : LETTERS[i] }}</span>
      <span class="opt__text">{{ o }}</span>
    </button>

    <Transition name="reveal" mode="out-in">
      <div v-if="solved" key="right" ref="result" class="result result--right">
        <strong>{{ s.quizWrong.length ? t('play.quizRightAfter') : t('play.quizRight') }}</strong>
        <p v-if="questExplain(quest)">{{ questExplain(quest) }}</p>
      </div>
      <p v-else-if="s.quizWrong.length" key="again" class="again">{{ t('play.quizTryAgain') }}</p>
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
  animation: pop 0.35s var(--spring);
}
.opt.right .opt__letter {
  color: var(--night);
  background: var(--c-chill);
}
.opt.wrong {
  opacity: 0.55;
  border-color: rgb(255 138 122 / 0.6);
  background: rgb(255 138 122 / 0.06);
}
.opt.wrong .opt__letter {
  color: var(--night);
  background: var(--coral);
}
.opt.shake {
  animation: shake 0.35s ease-in-out;
}
.opt.dim {
  opacity: 0.45;
}
.again {
  margin: 2px 0 0;
  font-size: 12.5px;
  font-weight: 800;
  color: var(--coral);
}
.result {
  margin-top: 4px;
  padding: 10px 12px;
  border-radius: 14px;
  font-size: 13px;
  line-height: 1.4;
  background: rgb(103 226 189 / 0.1);
}
.result strong {
  display: block;
  font-size: 14px;
  font-weight: 900;
  color: var(--c-chill);
}
.result p {
  margin: 4px 0 0;
  color: var(--muted);
  font-weight: 600;
}
.reveal-enter-active {
  transition: opacity 0.25s, transform 0.35s var(--spring);
}
.reveal-enter-from {
  opacity: 0;
  transform: translateY(6px);
}
@keyframes shake {
  20%, 60% { transform: translateX(-5px); }
  40%, 80% { transform: translateX(5px); }
}
@keyframes pop {
  40% { transform: scale(1.03); }
}
</style>
