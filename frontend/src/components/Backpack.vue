<script setup>
import { computed, ref, watch } from 'vue';
import Mascot from './Mascot.vue';
import { CATEGORIES, CATEGORY_KEYS, MOODS, SITUATIONS, categoryName, difficultyName } from '../lib/categories';
import { LANGUAGES, badgeText, entryTitle, lang, levelTitle, questTitle, t } from '../lib/i18n';
import { api } from '../lib/api';

const props = defineProps({
  view: { type: Object, required: true },
  // The player's own quests, newest first.
  mine: { type: Array, default: () => [] },
});
const emit = defineEmits(['prefs', 'close', 'hide', 'new', 'edit', 'delete', 'play', 'look', 'lang', 'tutorial', 'nudges', 'schedule']);

// The tab survives closing the backpack.
const TABS = [
  { id: 'boo', emoji: '👻' },
  { id: 'journal', emoji: '📓' },
  { id: 'mine', emoji: '✏️' },
  { id: 'settings', emoji: '⚙️' },
];
const tab = ref(lastTab);
watch(tab, (v) => (lastTab = v));

// Deleting takes two taps so a quest is never lost by accident.
const armed = ref('');
function remove(id) {
  if (armed.value === id) {
    armed.value = '';
    emit('delete', id);
  } else {
    armed.value = id;
  }
}
const isMac = /Mac/i.test(navigator.platform || navigator.userAgent);

// --- Boo tab ---
const level = computed(() => props.view.level);
const levelPct = computed(() => (level.value.xpInLevel / level.value.xpToNext) * 100);
const earned = computed(() => props.view.badges.filter((b) => b.earned).length);
const look = computed(() => props.view.look ?? { hat: 'hat-violet', extra: '' });
const hats = computed(() => (props.view.looks ?? []).filter((l) => l.kind === 'hat'));
const extras = computed(() => (props.view.looks ?? []).filter((l) => l.kind === 'extra'));
const HAT_SWATCH = {
  'hat-violet': '#8b6cff',
  'hat-mint': '#4fd8ad',
  'hat-coral': '#ff8a7a',
  'hat-gold': '#ffcb57',
  crown: 'linear-gradient(135deg, #ffe08a, #d99a14)',
};
const EXTRA_EMOJI = { glasses: '👓', scarf: '🧣', flower: '🌸' };
function wear(part, id) {
  emit('look', { ...look.value, [part]: id });
}
const difficulties = computed(() =>
  [1, 2, 3].map((d) => ({ d, open: d <= props.view.maxDifficulty, at: d === 1 ? 1 : d === 2 ? 3 : 5 })),
);

// --- Journal tab ---
const journal = ref(null);
async function loadJournal() {
  try {
    journal.value = await api.journal();
  } catch (err) {
    console.error(err);
    journal.value = { entries: [], insight: { category: '', moods: 0 } };
  }
}
watch(tab, (v) => v === 'journal' && loadJournal(), { immediate: true });
const moodEmoji = (id) => MOODS.find((m) => m.id === id)?.emoji ?? '';

function when(at) {
  const d = new Date(at);
  const today = new Date();
  const yesterday = new Date(Date.now() - 864e5);
  const same = (a, b) => a.toDateString() === b.toDateString();
  const time = d.toLocaleTimeString(lang.value === 'id' ? 'id-ID' : 'en-US', { hour: '2-digit', minute: '2-digit' });
  if (same(d, today)) return `${t('journal.today')} ${time}`;
  if (same(d, yesterday)) return `${t('journal.yesterday')} ${time}`;
  return d.toLocaleDateString(lang.value === 'id' ? 'id-ID' : 'en-US', { day: 'numeric', month: 'short' });
}

// --- Settings tab ---
const nudges = computed(() => props.view.nudges ?? { breakMinutes: 0, stayWhenBusy: false });
const BREAKS = [0, 60, 90];
const setBreak = (m) => emit('nudges', { ...nudges.value, breakMinutes: m });
const sched = computed(() => props.view.schedule ?? { times: [], days: 'daily', notify: false });
const DAYS = ['daily', 'weekdays', 'weekends'];
const newTime = ref('10:00');
const saveSched = (patch) => emit('schedule', { ...sched.value, ...patch });
function addTime() {
  const tm = newTime.value;
  if (!/^\d{2}:\d{2}$/.test(tm) || sched.value.times.includes(tm) || sched.value.times.length >= 4) return;
  saveSched({ times: [...sched.value.times, tm] });
}
const removeTime = (tm) => saveSched({ times: sched.value.times.filter((x) => x !== tm) });
const toggleStay = () => emit('nudges', { ...nudges.value, stayWhenBusy: !nudges.value.stayWhenBusy });
const muted = computed(() => new Set(props.view.prefs.muted));
const avoid = computed(() => new Set(props.view.prefs.avoid));

function toggle(list, value) {
  const next = new Set(list);
  next.has(value) ? next.delete(value) : next.add(value);
  return [...next];
}

function toggleCategory(key) {
  const next = toggle(muted.value, key);
  // Keep at least one category switched on.
  if (next.length >= CATEGORY_KEYS.length) return;
  emit('prefs', { ...props.view.prefs, muted: next });
}

function toggleSituation(need) {
  emit('prefs', { ...props.view.prefs, avoid: toggle(avoid.value, need) });
}
</script>

<script>
let lastTab = 'boo';
</script>

<template>
  <section class="pack" :aria-label="t('pack.title')">
    <header class="pack__head">
      <h2>🎒 {{ t('pack.title') }}</h2>
      <button class="icon-btn" :aria-label="t('pack.close')" @click="$emit('close')">
        <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true"><path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
      </button>
    </header>

    <nav class="tabs" role="tablist">
      <button
        v-for="tb in TABS"
        :key="tb.id"
        class="tabs__btn"
        :class="{ on: tab === tb.id }"
        role="tab"
        :aria-selected="tab === tb.id"
        @click="tab = tb.id"
      >
        <span aria-hidden="true">{{ tb.emoji }}</span> {{ t(`pack.tab.${tb.id}`) }}
      </button>
    </nav>

    <div class="pack__scroll">
      <!-- Boo: progress, outfit, quest levels, badges -->
      <template v-if="tab === 'boo'">
        <div class="hero">
          <Mascot :size="64" :hat="look.hat" :extra="look.extra" />
          <div class="hero__text">
            <strong>Lv {{ level.number }} · {{ levelTitle(level.number) }}</strong>
            <div class="bar"><span :style="{ width: levelPct + '%' }" /></div>
            <small>{{ level.xpInLevel }}/{{ level.xpToNext }} XP</small>
          </div>
        </div>
        <p class="pack__stats">
          {{ t('pack.stats', { done: view.completed, xp: view.xp, badges: earned, total: view.badges.length }) }}
        </p>
        <p class="explain">{{ t('pack.xpExplain') }}</p>

        <h3>{{ t('pack.outfit') }}</h3>
        <div class="looks">
          <button
            v-for="h in hats"
            :key="h.id"
            class="look"
            :class="{ on: look.hat === h.id, locked: !h.unlocked }"
            :disabled="!h.unlocked"
            :title="h.unlocked ? t(`look.${h.id}`) : t('pack.lockedAt', { n: h.level })"
            @click="wear('hat', h.id)"
          >
            <span class="look__swatch" :style="{ background: HAT_SWATCH[h.id] }">{{ h.id === 'crown' ? '👑' : '' }}</span>
            <small>{{ h.unlocked ? t(`look.${h.id}`) : `🔒 ${t('pack.lockedAt', { n: h.level })}` }}</small>
          </button>
        </div>
        <div class="looks">
          <button class="look" :class="{ on: !look.extra }" @click="wear('extra', '')">
            <span class="look__swatch look__swatch--none">–</span>
            <small>{{ t('pack.none') }}</small>
          </button>
          <button
            v-for="x in extras"
            :key="x.id"
            class="look"
            :class="{ on: look.extra === x.id, locked: !x.unlocked }"
            :disabled="!x.unlocked"
            :title="x.unlocked ? t(`look.${x.id}`) : t('pack.lockedAt', { n: x.level })"
            @click="wear('extra', x.id)"
          >
            <span class="look__swatch look__swatch--emoji">{{ EXTRA_EMOJI[x.id] }}</span>
            <small>{{ x.unlocked ? t(`look.${x.id}`) : `🔒 ${t('pack.lockedAt', { n: x.level })}` }}</small>
          </button>
        </div>

        <h3>{{ t('pack.difficulty') }}</h3>
        <ul class="levels">
          <li v-for="d in difficulties" :key="d.d" :class="{ open: d.open }">
            <span class="dots" aria-hidden="true"><i v-for="n in 3" :key="n" :class="{ on: n <= d.d }" /></span>
            <strong>{{ difficultyName(d.d) }}</strong>
            <span class="levels__state">{{ d.open ? `✓ ${t('pack.difficultyOpen')}` : `🔒 ${t('pack.lockedAt', { n: d.at })}` }}</span>
          </li>
        </ul>

        <h3>{{ t('pack.badges') }}</h3>
        <ul class="badges">
          <li
            v-for="b in view.badges"
            :key="b.id"
            :class="{ earned: b.earned }"
            :title="b.earned ? badgeText(b.id, b).name : t('pack.badgeLocked', { hint: badgeText(b.id, b).hint })"
          >
            <span class="badges__emoji">{{ b.earned ? b.emoji : '?' }}</span>
            <span class="badges__name">{{ b.earned ? badgeText(b.id, b).name : badgeText(b.id, b).hint }}</span>
          </li>
        </ul>
      </template>

      <!-- Journal -->
      <template v-else-if="tab === 'journal'">
        <div v-if="journal" class="insight">
          <p v-if="journal.insight.category">
            {{ t('journal.insight', { cat: categoryName(journal.insight.category), emoji: CATEGORIES[journal.insight.category]?.emoji }) }}
          </p>
          <p v-else>{{ t('journal.insightSoon') }}</p>
          <small>🔒 {{ t('journal.private') }}</small>
        </div>
        <p v-if="journal && !journal.entries.length" class="pack__hint">{{ t('journal.empty') }}</p>
        <ul v-else-if="journal" class="entries">
          <li v-for="e in journal.entries" :key="e.id" :style="{ '--accent': CATEGORIES[e.category]?.color }">
            <div class="entry__head">
              <span class="entry__title">{{ CATEGORIES[e.category]?.emoji ?? '✨' }} {{ entryTitle(e) }}</span>
              <span class="entry__mood" :title="e.mood ? t(`mood.${e.mood}`) : ''">{{ moodEmoji(e.mood) }}</span>
            </div>
            <div class="entry__meta">{{ when(e.at) }} · +{{ e.xp }} XP</div>
            <div v-if="e.answers?.length" class="entry__answers">
              <span v-for="(a, i) in e.answers" :key="i">{{ a }}</span>
            </div>
            <div v-if="e.count" class="entry__meta">{{ t('journal.count', { n: e.count }) }}</div>
            <img v-if="e.doodle" class="entry__doodle" :src="e.doodle" alt="" />
          </li>
        </ul>
      </template>

      <!-- My quests -->
      <template v-else-if="tab === 'mine'">
        <div class="mine__head">
          <p class="pack__hint">{{ mine.length ? '' : t('mine.empty') }}</p>
          <button class="mine__new" @click="$emit('new')">{{ t('mine.new') }}</button>
        </div>
        <ul v-if="mine.length" class="mine">
          <li v-for="q in mine" :key="q.id" :style="{ '--accent': CATEGORIES[q.category]?.color }">
            <button class="mine__title" :title="t('mine.edit', { title: questTitle(q) })" @click="$emit('edit', q)">
              <span aria-hidden="true">{{ CATEGORIES[q.category]?.emoji }}</span>
              <span class="mine__text">{{ questTitle(q) }}</span>
            </button>
            <button class="mine__icon" :aria-label="t('mine.play')" :title="t('mine.play')" @click="$emit('play', q.id)">▶</button>
            <button
              class="mine__icon mine__del"
              :class="{ armed: armed === q.id }"
              :aria-label="armed === q.id ? t('mine.deleteConfirm') : t('mine.delete')"
              :title="armed === q.id ? t('mine.deleteHint') : t('mine.delete')"
              @click="remove(q.id)"
              @blur="armed === q.id && (armed = '')"
            >
              {{ armed === q.id ? t('mine.deleteConfirm') : '✕' }}
            </button>
          </li>
        </ul>
      </template>

      <!-- Settings -->
      <template v-else>
        <h3>{{ t('pack.language') }}</h3>
        <div class="chips">
          <button
            v-for="l in LANGUAGES"
            :key="l.code"
            class="toggle toggle--situation"
            :class="{ on: lang === l.code }"
            :aria-pressed="lang === l.code"
            @click="$emit('lang', l.code)"
          >
            {{ l.flag }} {{ l.label }}
          </button>
        </div>

        <h3>{{ t('pack.vibes') }}</h3>
        <p class="pack__hint">{{ t('pack.vibesHint') }}</p>
        <div class="chips">
          <button
            v-for="key in CATEGORY_KEYS"
            :key="key"
            class="toggle"
            :class="{ off: muted.has(key) }"
            :style="{ '--accent': CATEGORIES[key].color }"
            :aria-pressed="!muted.has(key)"
            @click="toggleCategory(key)"
          >
            {{ CATEGORIES[key].emoji }} {{ categoryName(key) }}
          </button>
        </div>

        <h3>{{ t('pack.rightNow') }}</h3>
        <div class="chips">
          <button
            v-for="s in SITUATIONS"
            :key="s.need"
            class="toggle toggle--situation"
            :class="{ on: avoid.has(s.need) }"
            :aria-pressed="avoid.has(s.need)"
            @click="toggleSituation(s.need)"
          >
            {{ s.emoji }} {{ t(`need.${s.need}`) }}
          </button>
        </div>

        <h3>{{ t('pack.schedule') }}</h3>
        <p class="pack__hint">{{ t('pack.scheduleHint') }}</p>
        <div class="chips">
          <span v-for="tm in sched.times" :key="tm" class="time-chip">
            ⏰ {{ tm }}
            <button class="time-chip__x" :aria-label="t('pack.scheduleRemove', { time: tm })" @click="removeTime(tm)">✕</button>
          </span>
          <span v-if="sched.times.length < 4" class="time-add">
            <input v-model="newTime" type="time" class="time-add__input" :aria-label="t('pack.scheduleTime')" />
            <button class="mine__new time-add__btn" @click="addTime">+ {{ t('pack.scheduleAdd') }}</button>
          </span>
        </div>
        <div v-if="sched.times.length" class="chips chips--gap">
          <button
            v-for="d in DAYS"
            :key="d"
            class="toggle toggle--situation"
            :class="{ on: sched.days === d }"
            :aria-pressed="sched.days === d"
            @click="saveSched({ days: d })"
          >
            {{ t(`pack.days.${d}`) }}
          </button>
          <button
            class="toggle toggle--situation"
            :class="{ on: sched.notify }"
            :aria-pressed="sched.notify"
            @click="saveSched({ notify: !sched.notify })"
          >
            🔔 {{ t('pack.scheduleNotify') }}
          </button>
        </div>

        <h3>{{ t('pack.breaks') }}</h3>
        <p class="pack__hint">{{ t('pack.breaksHint') }}</p>
        <div class="chips">
          <button
            v-for="m in BREAKS"
            :key="m"
            class="toggle toggle--situation"
            :class="{ on: nudges.breakMinutes === m }"
            :aria-pressed="nudges.breakMinutes === m"
            @click="setBreak(m)"
          >
            {{ m ? t('pack.breakEvery', { n: m }) : t('pack.breakOff') }}
          </button>
        </div>
        <template v-if="!isMac">
          <h3>{{ t('pack.busy') }}</h3>
          <div class="chips">
            <button
              class="toggle toggle--situation"
              :class="{ on: !nudges.stayWhenBusy }"
              :aria-pressed="!nudges.stayWhenBusy"
              @click="toggleStay"
            >
              🙈 {{ t('pack.busyHide') }}
            </button>
          </div>
          <p class="pack__hint pack__hint--after">{{ t('pack.busyHint') }}</p>
        </template>

        <button class="btn btn--ghost pack__quit" @click="$emit('tutorial')">👋 {{ t('pack.tutorial') }}</button>
        <button class="btn btn--ghost pack__quit pack__quit--tight" @click="$emit('hide')">{{ t('pack.hide') }}</button>
        <p class="pack__where">{{ isMac ? t('pack.whereMac') : t('pack.whereWin') }}</p>
      </template>
    </div>
  </section>
</template>

<style scoped>
.pack {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}
.pack__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2px 6px;
}
.pack__head h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 900;
}
.pack__scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0 2px 4px;
  scrollbar-width: thin;
  scrollbar-color: var(--night-3) transparent;
}
.pack__stats {
  margin: 0 0 4px;
  font-size: 13px;
  color: var(--muted);
}
.pack__stats strong {
  color: var(--ghost);
}
h3 {
  margin: 14px 0 6px;
  font-size: 12px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--faint);
}
.chips--gap {
  margin-top: 6px;
}
.time-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 4px 4px 10px;
  border-radius: 99px;
  font-size: 13px;
  font-weight: 800;
  background: rgb(139 108 255 / 0.2);
  border: 1.5px solid var(--violet);
  font-variant-numeric: tabular-nums;
}
.time-chip__x {
  display: grid;
  place-items: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  font-size: 10px;
  color: var(--muted);
}
.time-chip__x:hover {
  background: rgb(255 255 255 / 0.1);
  color: var(--ghost);
}
.time-add {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}
.time-add__input {
  height: 30px;
  padding: 0 6px;
  border-radius: 10px;
  border: 1.5px solid var(--line);
  background: rgb(0 0 0 / 0.2);
  color: var(--ghost);
  font: inherit;
  font-size: 13px;
  font-weight: 800;
  color-scheme: dark;
}
.time-add__btn {
  margin-top: 0;
}
.pack__hint--after {
  margin: 6px 0 0;
}
.pack__hint {
  margin: -2px 0 8px;
  font-size: 12px;
  color: var(--faint);
}

.badges {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}
.badges li {
  display: grid;
  justify-items: center;
  gap: 2px;
  padding: 8px 4px;
  border-radius: 14px;
  background: var(--night-2);
  border: 1px dashed var(--line);
  text-align: center;
}
.badges li.earned {
  border-style: solid;
  border-color: rgb(255 203 87 / 0.35);
  background: rgb(255 203 87 / 0.07);
}
.badges__emoji {
  font-size: 22px;
  line-height: 1.2;
  color: var(--faint);
  font-weight: 900;
}
.badges__name {
  font-size: 10px;
  font-weight: 800;
  line-height: 1.15;
  color: var(--faint);
}
.earned .badges__name {
  color: var(--ghost);
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.toggle {
  padding: 6px 11px;
  border-radius: 99px;
  font-size: 13px;
  font-weight: 800;
  background: color-mix(in srgb, var(--accent, var(--violet)) 18%, transparent);
  border: 1.5px solid var(--accent, var(--violet));
  transition: opacity 0.2s, transform 0.2s var(--spring), background 0.2s;
}
.toggle:active {
  transform: scale(0.94);
}
.toggle.off {
  opacity: 0.4;
  background: transparent;
  border-style: dashed;
  text-decoration: line-through;
}
.toggle--situation {
  --accent: var(--line);
  background: transparent;
  border-color: var(--line);
  color: var(--muted);
}
.toggle--situation.on {
  color: var(--ghost);
  border-color: var(--violet);
  background: rgb(139 108 255 / 0.2);
}

.mine__head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}
.mine__head .pack__hint {
  margin: 8px 0 0;
}
.mine__new {
  margin-top: 8px;
  padding: 4px 10px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 900;
  color: var(--gold);
  border: 1.5px solid rgb(255 203 87 / 0.45);
  transition: background 0.2s, transform 0.2s var(--spring);
}
.mine__new:hover {
  background: rgb(255 203 87 / 0.1);
}
.mine__new:active {
  transform: scale(0.94);
}
.mine {
  display: grid;
  gap: 4px;
  margin: 0;
  padding: 0;
  list-style: none;
}
.mine li {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 2px 4px 2px 0;
  border-radius: 12px;
  background: var(--night-2);
  border-left: 3px solid var(--accent, var(--violet));
}
.mine__title {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: 1;
  min-width: 0;
  padding: 6px 8px;
  text-align: left;
  font-size: 13px;
  font-weight: 800;
}
.mine__text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.mine__title:hover .mine__text {
  text-decoration: underline;
  text-decoration-color: var(--faint);
}
.mine__icon {
  flex: none;
  min-width: 28px;
  height: 28px;
  padding: 0 6px;
  border-radius: 8px;
  font-size: 11px;
  font-weight: 900;
  color: var(--faint);
  transition: background 0.2s, color 0.2s;
}
.mine__icon:hover {
  background: rgb(255 255 255 / 0.07);
  color: var(--ghost);
}
.mine__del.armed {
  color: #2a0f17;
  background: var(--coral);
}
.pack__quit {
  display: flex;
  margin: 14px auto 0;
}
.pack__quit--tight {
  margin-top: 0;
}

/* tabs */
.tabs {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 4px;
  margin-bottom: 8px;
  padding: 3px;
  border-radius: 12px;
  background: rgb(0 0 0 / 0.2);
}
.tabs__btn {
  padding: 5px 2px;
  border-radius: 9px;
  font-size: 11.5px;
  font-weight: 800;
  color: var(--faint);
  white-space: nowrap;
  transition: background 0.2s, color 0.2s;
}
.tabs__btn.on {
  background: var(--night-3);
  color: var(--ghost);
}

/* Boo tab */
.hero {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 4px 4px;
}
.hero__text {
  flex: 1;
  display: grid;
  gap: 4px;
}
.hero__text strong {
  font-size: 15px;
  font-weight: 900;
}
.hero__text small {
  font-size: 11px;
  font-weight: 800;
  color: var(--faint);
}
.explain {
  margin: 6px 0 0;
  padding: 8px 10px;
  border-radius: 12px;
  font-size: 12px;
  line-height: 1.4;
  color: var(--muted);
  background: rgb(139 108 255 / 0.1);
}
.looks {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 4px;
  margin-bottom: 6px;
}
.look {
  display: grid;
  justify-items: center;
  gap: 3px;
  padding: 6px 2px;
  border-radius: 12px;
  border: 1.5px solid transparent;
  transition: border-color 0.2s, background 0.2s;
}
.look:hover:not(:disabled) {
  background: rgb(255 255 255 / 0.04);
}
.look.on {
  border-color: var(--gold);
  background: rgb(255 203 87 / 0.08);
}
.look.locked {
  opacity: 0.45;
  cursor: default;
}
.look small {
  font-size: 9.5px;
  font-weight: 800;
  line-height: 1.15;
  text-align: center;
  color: var(--muted);
}
.look__swatch {
  display: grid;
  place-items: center;
  width: 26px;
  height: 26px;
  border-radius: 50%;
  font-size: 14px;
  box-shadow: inset 0 0 0 1.5px rgb(255 255 255 / 0.2);
}
.look__swatch--none,
.look__swatch--emoji {
  background: var(--night-2);
  color: var(--faint);
  font-weight: 900;
}
.levels {
  display: grid;
  gap: 4px;
  margin: 0;
  padding: 0;
  list-style: none;
}
.levels li {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 10px;
  border-radius: 10px;
  background: var(--night-2);
  font-size: 13px;
  color: var(--faint);
}
.levels li.open {
  color: var(--ghost);
}
.levels__state {
  margin-left: auto;
  font-size: 11.5px;
  font-weight: 800;
}
.open .levels__state {
  color: var(--c-chill);
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
  background: var(--gold);
}

/* journal */
.insight {
  padding: 10px 12px;
  margin-bottom: 8px;
  border-radius: 14px;
  background: linear-gradient(135deg, rgb(139 108 255 / 0.18), rgb(103 226 189 / 0.12));
  border: 1px solid var(--line);
}
.insight p {
  margin: 0 0 4px;
  font-size: 13px;
  font-weight: 800;
  line-height: 1.35;
}
.insight small {
  font-size: 11px;
  color: var(--faint);
}
.entries {
  display: grid;
  gap: 6px;
  margin: 0;
  padding: 0;
  list-style: none;
}
.entries li {
  padding: 8px 10px;
  border-radius: 12px;
  background: var(--night-2);
  border-left: 3px solid var(--accent, var(--violet));
}
.entry__head {
  display: flex;
  align-items: center;
  gap: 6px;
}
.entry__title {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  font-weight: 900;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.entry__mood {
  font-size: 16px;
}
.entry__meta {
  font-size: 11px;
  font-weight: 700;
  color: var(--faint);
}
.entry__answers {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 5px;
}
.entry__answers span {
  padding: 2px 8px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 700;
  background: rgb(255 255 255 / 0.06);
  user-select: text;
}
.entry__doodle {
  display: block;
  width: 100%;
  margin-top: 6px;
  border-radius: 10px;
  background: rgb(0 0 0 / 0.25);
}
.pack__where {
  margin: 0 0 4px;
  text-align: center;
  font-size: 11.5px;
  color: var(--faint);
}
</style>
