<script setup>
// Fill in a few words without seeing the story, then reveal it. The
// player's words are highlighted inside the story.
import { computed } from 'vue';
import { playFor } from '../../lib/play';
import { fillStory, questPrompts, t } from '../../lib/i18n';
import { play as sound } from '../../lib/sound';

const props = defineProps({ quest: { type: Object, required: true } });
const s = computed(() => playFor(props.quest));
const blanks = computed(() => questPrompts(props.quest));
const filled = computed(() => s.value.answers.every((a) => a.trim()));
const story = computed(() => fillStory(props.quest, s.value.answers));

function reveal() {
  if (!filled.value) return;
  s.value.storyShown = true;
  sound('pop');
}
</script>

<template>
  <div class="madlib">
    <Transition name="swap" mode="out-in">
      <form v-if="!s.storyShown" key="fill" class="fill" @submit.prevent="reveal">
        <label v-for="(b, i) in blanks" :key="i" class="blank">
          <span class="blank__label">{{ b }}</span>
          <input v-model="s.answers[i]" class="blank__input" maxlength="40" :placeholder="'…'" />
        </label>
        <!-- Enter reveals the story; the main button below does too. -->
        <button type="submit" hidden />
      </form>

      <div v-else key="story" class="story">
        <p class="story__text">
          <template v-for="(part, i) in story" :key="i">
            <mark v-if="part.word" class="story__word">{{ part.text }}</mark>
            <template v-else>{{ part.text }}</template>
          </template>
        </p>
        <button class="btn btn--ghost story__edit" @click="s.storyShown = false">✏️ {{ t('play.madlibEdit') }}</button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.fill {
  display: grid;
  gap: 8px;
}
.blank {
  display: grid;
  grid-template-columns: 110px 1fr;
  align-items: center;
  gap: 10px;
}
.blank__label {
  font-size: 12px;
  font-weight: 800;
  color: var(--muted);
  line-height: 1.25;
}
.blank__input {
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
.blank__input:focus {
  border-color: var(--accent, var(--violet));
}
.story {
  padding: 14px 16px 10px;
  border-radius: 16px;
  background: color-mix(in srgb, var(--accent, var(--violet)) 10%, rgb(0 0 0 / 0.2));
  border: 1px solid color-mix(in srgb, var(--accent, var(--violet)) 30%, transparent);
}
.story__text {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.55;
  user-select: text;
}
.story__word {
  padding: 1px 6px;
  border-radius: 6px;
  color: var(--night);
  background: var(--accent, var(--violet));
  font-weight: 900;
}
.story__edit {
  height: 28px;
  margin-top: 6px;
  padding: 0;
}
.swap-enter-active,
.swap-leave-active {
  transition: opacity 0.18s, transform 0.25s var(--ease);
}
.swap-enter-from {
  opacity: 0;
  transform: scale(0.97);
}
.swap-leave-to {
  opacity: 0;
}
</style>
