<script setup>
// A short first-run intro: pick a language, then three quick cards on how
// SideQuest works. It can be replayed from the backpack.
import { computed, ref } from 'vue';
import Mascot from './Mascot.vue';
import { LANGUAGES, lang, t } from '../lib/i18n';

defineProps({
  look: { type: Object, default: () => ({}) },
});
const emit = defineEmits(['lang', 'done']);

const step = ref(0);
const STEPS = 4;
const last = computed(() => step.value === STEPS - 1);

function next() {
  if (last.value) emit('done');
  else step.value++;
}

const PLAY = [
  ['✏️', 'kind.answer'],
  ['☑️', 'kind.checklist'],
  ['⏱️', 'kind.timer'],
  ['👆', 'kind.counter'],
  ['🖍️', 'kind.doodle'],
];
</script>

<template>
  <section class="intro" :aria-label="t('intro.hi')">
    <div class="intro__dots" aria-hidden="true">
      <i v-for="n in STEPS" :key="n" :class="{ on: n - 1 === step }" />
    </div>

    <Transition name="slide" mode="out-in">
      <div :key="step" class="intro__step">
        <!-- 0: hello + language -->
        <template v-if="step === 0">
          <Mascot :size="88" mood="wink" :hat="look.hat" :extra="look.extra" class="intro__boo" />
          <h2>{{ t('intro.hi') }}</h2>
          <p>{{ t('intro.hiText') }}</p>
          <div class="langs">
            <button
              v-for="l in LANGUAGES"
              :key="l.code"
              class="lang"
              :class="{ on: lang === l.code }"
              :aria-pressed="lang === l.code"
              @click="$emit('lang', l.code)"
            >
              <span aria-hidden="true">{{ l.flag }}</span> {{ l.label }}
            </button>
          </div>
        </template>

        <!-- 1: how it works -->
        <template v-else-if="step === 1">
          <div class="intro__art" aria-hidden="true">🎲</div>
          <h2>{{ t('intro.how') }}</h2>
          <ul class="how">
            <li><span class="how__btn how__btn--go">{{ t('action.go') }}</span>{{ t('intro.howGo') }}</li>
            <li><span class="how__btn">🎲 {{ t('action.reroll') }}</span>{{ t('intro.howReroll') }}</li>
            <li><span class="how__btn">{{ t('action.skip') }}</span>{{ t('intro.howSkip') }}</li>
          </ul>
        </template>

        <!-- 2: play in the app -->
        <template v-else-if="step === 2">
          <div class="intro__art" aria-hidden="true">✍️</div>
          <h2>{{ t('intro.play') }}</h2>
          <p>{{ t('intro.playText') }}</p>
          <div class="kinds">
            <span v-for="[emoji, key] in PLAY" :key="key">{{ emoji }} {{ t(key) }}</span>
          </div>
        </template>

        <!-- 3: XP and levels -->
        <template v-else>
          <div class="intro__art" aria-hidden="true">✨</div>
          <h2>{{ t('intro.grow') }}</h2>
          <p>{{ t('intro.growText') }}</p>
          <div class="grow" aria-hidden="true">
            <Mascot :size="44" hat="hat-violet" />
            <span>→</span>
            <Mascot :size="44" hat="hat-mint" extra="glasses" />
            <span>→</span>
            <Mascot :size="44" hat="crown" extra="scarf" />
          </div>
        </template>
      </div>
    </Transition>

    <div class="intro__foot">
      <button class="btn btn--primary" @click="next">{{ last ? t('intro.start') : t('intro.next') }}</button>
      <button v-if="!last" class="btn btn--ghost" @click="$emit('done')">{{ t('intro.skip') }}</button>
    </div>
  </section>
</template>

<style scoped>
.intro {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
  text-align: center;
}
.intro__dots {
  display: flex;
  justify-content: center;
  gap: 6px;
  padding: 2px 0 6px;
}
.intro__dots i {
  width: 7px;
  height: 7px;
  border-radius: 99px;
  background: rgb(255 255 255 / 0.15);
  transition: width 0.3s var(--ease), background 0.3s;
}
.intro__dots i.on {
  width: 20px;
  background: var(--gold);
}
.intro__step {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  overflow-y: auto;
  scrollbar-width: none;
}
.intro__boo {
  animation: pop 0.6s var(--spring) both;
}
.intro__art {
  font-size: 44px;
  line-height: 1;
  animation: pop 0.5s var(--spring) both;
}
h2 {
  margin: 0;
  font-size: 22px;
  font-weight: 900;
}
p {
  margin: 0;
  max-width: 300px;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.45;
  color: var(--muted);
}
.langs {
  display: grid;
  gap: 8px;
  width: 100%;
  margin-top: 6px;
}
.lang {
  height: 44px;
  border-radius: 14px;
  font-size: 15px;
  font-weight: 900;
  background: var(--night-2);
  border: 1.5px solid var(--line);
  transition: border-color 0.2s, background 0.2s, transform 0.2s var(--spring);
}
.lang:active {
  transform: scale(0.97);
}
.lang.on {
  border-color: var(--gold);
  background: rgb(255 203 87 / 0.1);
}
.how {
  display: grid;
  gap: 8px;
  width: 100%;
  margin: 4px 0 0;
  padding: 0;
  list-style: none;
  text-align: left;
}
.how li {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13.5px;
  font-weight: 700;
  color: var(--muted);
}
.how__btn {
  flex: none;
  min-width: 96px;
  padding: 6px 10px;
  border-radius: 10px;
  text-align: center;
  font-size: 13px;
  font-weight: 900;
  color: var(--ghost);
  background: var(--night-3);
  border: 1px solid var(--line);
}
.how__btn--go {
  color: #1d1240;
  background: var(--gold);
  border-color: var(--gold);
}
.kinds {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 6px;
  margin-top: 4px;
}
.kinds span {
  padding: 4px 10px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 800;
  background: var(--night-2);
  border: 1px solid var(--line);
}
.grow {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  color: var(--faint);
  font-weight: 900;
}
.intro__foot {
  display: grid;
  gap: 4px;
  padding-top: 8px;
}

.slide-enter-active,
.slide-leave-active {
  transition: opacity 0.18s, transform 0.25s var(--ease);
}
.slide-enter-from {
  opacity: 0;
  transform: translateX(24px);
}
.slide-leave-to {
  opacity: 0;
  transform: translateX(-24px);
}
@keyframes pop {
  from { transform: scale(0.5); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
</style>
