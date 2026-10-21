<script setup>
import { computed } from 'vue';
import { category, difficultyName } from '../lib/categories';
import { questDesc, questTitle, t, twistText } from '../lib/i18n';
import QuestPlay from './QuestPlay.vue';

// How each kind of quest is played, shown before accepting.
const KINDS = {
  answer: '✏️',
  checklist: '☑️',
  timer: '⏱️',
  counter: '👆',
  doodle: '🖍️',
  choice: '👉',
  madlib: '📝',
  stopwatch: '🏅',
  spinner: '🎰',
  order: '↕️',
  quiz: '🧠',
};

const props = defineProps({
  view: { type: Object, required: true },
  busy: { type: Boolean, default: false },
});
defineEmits(['toggle-twist']);

const q = computed(() => props.view.quest);
const cat = computed(() => category(q.value.category));
// Custom packs may use the full allowed text length; shrink type to fit.
const title = computed(() => questTitle(q.value));
const desc = computed(() => questDesc(q.value));
const long = computed(() => title.value.length > 32 || desc.value.length > 130);
const kind = computed(() => q.value.interaction?.kind ?? '');
</script>

<template>
  <article class="card" :class="{ 'card--active': view.accepted, 'card--long': long }" :style="{ '--accent': cat.color }">
    <header class="card__top">
      <span class="chip">
        <span aria-hidden="true">{{ cat.emoji }}</span>
        {{ cat.label }}
      </span>
      <span v-if="view.accepted" class="ribbon">{{ t('card.onQuest') }}</span>
      <span v-else-if="q.source === 'custom'" class="mine">✎ {{ t('card.mine') }}</span>
      <span v-else-if="kind" class="kind">{{ KINDS[kind] }} {{ t(`kind.${kind}`) }}</span>
    </header>

    <h2 class="card__title">{{ title }}</h2>
    <p class="card__desc">{{ desc }}</p>

    <!-- Once accepted, the space goes to playing; XP shows in the celebration. -->
    <ul v-if="!view.accepted" class="meta">
      <li>⏱ {{ t('card.minutes', { n: q.minutes }) }}</li>
      <li>
        <span class="dots" aria-hidden="true">
          <i v-for="n in 3" :key="n" :class="{ on: n <= q.difficulty }" />
        </span>
        {{ difficultyName(q.difficulty) }}
      </li>
      <li class="meta__xp">✦ {{ view.reward }} XP</li>
    </ul>

    <QuestPlay v-if="view.accepted && kind" :quest="q" :best="view.best ?? 0" />

    <p v-if="view.accepted && view.twist && view.twistOn" class="twist-note">
      🌀 {{ twistText(q, view.twist) }} <strong v-if="view.twist.bonus">+{{ view.twist.bonus }}</strong>
    </p>
    <button
      v-else-if="view.twist && !view.accepted"
      class="twist"
      :class="{ 'twist--on': view.twistOn }"
      :aria-pressed="view.twistOn"
      :disabled="busy"
      @click="$emit('toggle-twist')"
    >
      <span class="twist__icon" aria-hidden="true">🌀</span>
      <span class="twist__text">
        <strong>{{ view.twistOn ? t('card.twistOn') : t('card.twistAdd') }}</strong>
        {{ twistText(q, view.twist) }}
      </span>
      <span v-if="view.twist.bonus" class="twist__bonus">+{{ view.twist.bonus }}</span>
    </button>
  </article>
</template>

<style scoped>
.card {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 16px 18px 14px;
  border-radius: 22px;
  background:
    radial-gradient(140% 90% at 100% 0%, color-mix(in srgb, var(--accent) 22%, transparent), transparent 60%),
    var(--night-2);
  border: 1px solid var(--line);
  box-shadow: inset 0 1px 0 rgb(255 255 255 / 0.05);
}
.card--long .card__title {
  font-size: 19px;
}
.card--long .card__desc {
  font-size: 13.5px;
}
.card--active {
  border-color: color-mix(in srgb, var(--accent) 55%, transparent);
}

.card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px 3px 8px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 800;
  letter-spacing: 0.02em;
  color: var(--night);
  background: var(--accent);
}
.mine,
.kind {
  font-size: 11px;
  font-weight: 800;
  color: var(--faint);
}
.ribbon {
  font-size: 11px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--accent);
  animation: pulse 2s ease-in-out infinite;
}

.card__title {
  margin: 4px 0 0;
  font-size: 23px;
  line-height: 1.15;
  font-weight: 900;
  letter-spacing: -0.01em;
  text-wrap: balance;
}
.card__desc {
  margin: 0;
  color: var(--muted);
  font-size: 15px;
  font-weight: 600;
  text-wrap: pretty;
}

.meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  margin: 4px 0 0;
  padding: 0;
  list-style: none;
  font-size: 12.5px;
  font-weight: 800;
  color: var(--faint);
}
.meta li {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.meta__xp {
  color: var(--gold);
}
.dots {
  display: inline-flex;
  gap: 2px;
}
.dots i {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgb(255 255 255 / 0.14);
}
.dots i.on {
  background: var(--accent);
}

.twist-note {
  margin: 2px 0 0;
  font-size: 12.5px;
  font-weight: 700;
  color: var(--muted);
}
.twist-note strong {
  color: var(--gold);
}
.twist {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 6px;
  padding: 9px 12px;
  border-radius: 14px;
  border: 1.5px dashed rgb(201 184 255 / 0.3);
  text-align: left;
  font-size: 13px;
  font-weight: 600;
  color: var(--muted);
  transition: border-color 0.2s, background 0.2s, transform 0.2s var(--spring);
}
.twist:hover {
  border-color: rgb(201 184 255 / 0.55);
  background: rgb(255 255 255 / 0.03);
}
.twist:active {
  transform: scale(0.98);
}
.twist__icon {
  font-size: 18px;
  transition: transform 0.5s var(--spring);
}
.twist__text {
  flex: 1;
  line-height: 1.3;
}
.twist__text strong {
  display: block;
  font-weight: 900;
  color: var(--ghost);
}
.twist__bonus {
  font-weight: 900;
  color: var(--gold);
}
.twist--on {
  border-style: solid;
  border-color: var(--gold);
  background: rgb(255 203 87 / 0.08);
}
.twist--on .twist__icon {
  transform: rotate(360deg);
}

@keyframes pulse {
  50% { opacity: 0.55; }
}
</style>
