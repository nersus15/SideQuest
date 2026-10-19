<script setup>
import { computed, ref } from 'vue';
import { CATEGORIES, CATEGORY_KEYS, SITUATIONS } from '../lib/categories';

const props = defineProps({
  view: { type: Object, required: true },
  // The player's own quests, newest first.
  mine: { type: Array, default: () => [] },
});
const emit = defineEmits(['prefs', 'close', 'hide', 'new', 'edit', 'delete', 'play']);

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

const earned = computed(() => props.view.badges.filter((b) => b.earned).length);
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

<template>
  <section class="pack" aria-label="Backpack">
    <header class="pack__head">
      <h2>🎒 Backpack</h2>
      <button class="icon-btn" aria-label="Close backpack" @click="$emit('close')">
        <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true"><path d="M4 4l8 8M12 4l-8 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" /></svg>
      </button>
    </header>

    <div class="pack__scroll">
      <p class="pack__stats">
        <strong>{{ view.completed }}</strong> quests done · <strong>{{ view.xp }}</strong> XP ·
        <strong>{{ earned }}</strong>/{{ view.badges.length }} badges
      </p>

      <div class="mine__head">
        <h3>My quests</h3>
        <button class="mine__new" @click="$emit('new')">+ Make a quest</button>
      </div>
      <p v-if="!mine.length" class="pack__hint">Got a fun idea? Turn it into a quest and Boo will offer it to you.</p>
      <ul v-else class="mine">
        <li v-for="q in mine" :key="q.id" :style="{ '--accent': CATEGORIES[q.category]?.color }">
          <button class="mine__title" :title="`Edit “${q.title}”`" @click="$emit('edit', q)">
            <span aria-hidden="true">{{ CATEGORIES[q.category]?.emoji }}</span>
            <span class="mine__text">{{ q.title }}</span>
          </button>
          <button class="mine__icon" :aria-label="`Play ${q.title} now`" title="Play now" @click="$emit('play', q.id)">▶</button>
          <button
            class="mine__icon mine__del"
            :class="{ armed: armed === q.id }"
            :aria-label="armed === q.id ? `Really delete ${q.title}?` : `Delete ${q.title}`"
            :title="armed === q.id ? 'Tap again to delete' : 'Delete'"
            @click="remove(q.id)"
            @blur="armed === q.id && (armed = '')"
          >
            {{ armed === q.id ? 'Delete?' : '✕' }}
          </button>
        </li>
      </ul>

      <h3>Badges</h3>
      <ul class="badges">
        <li
          v-for="b in view.badges"
          :key="b.id"
          :class="{ earned: b.earned }"
          :title="b.earned ? b.name : `Locked · ${b.hint}`"
        >
          <span class="badges__emoji">{{ b.earned ? b.emoji : '?' }}</span>
          <span class="badges__name">{{ b.earned ? b.name : b.hint }}</span>
        </li>
      </ul>

      <h3>Quest vibes</h3>
      <p class="pack__hint">Tap to turn kinds of quests on or off.</p>
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
          {{ CATEGORIES[key].emoji }} {{ CATEGORIES[key].label }}
        </button>
      </div>

      <h3>Right now I'm…</h3>
      <div class="chips">
        <button
          v-for="s in SITUATIONS"
          :key="s.need"
          class="toggle toggle--situation"
          :class="{ on: avoid.has(s.need) }"
          :aria-pressed="avoid.has(s.need)"
          @click="toggleSituation(s.need)"
        >
          {{ s.emoji }} {{ s.label }}
        </button>
      </div>

      <div v-if="view.recent.length" class="recent">
        <h3>Recently done</h3>
        <ul>
          <li v-for="(r, i) in view.recent" :key="i">
            <span>{{ CATEGORIES[r.category]?.emoji ?? '✨' }} {{ r.title }}</span>
            <span class="recent__xp">+{{ r.xp }}</span>
          </li>
        </ul>
      </div>

      <button class="btn btn--ghost pack__quit" @click="$emit('hide')">Hide SideQuest for now</button>
      <p class="pack__where">Boo waits in the {{ isMac ? 'menu bar' : 'system tray' }}. Quit SideQuest from there.</p>
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

.recent ul {
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: 13px;
  font-weight: 700;
}
.recent li {
  display: flex;
  justify-content: space-between;
  padding: 5px 0;
  border-bottom: 1px solid var(--line);
  color: var(--muted);
}
.recent__xp {
  color: var(--gold);
  font-weight: 900;
}
.mine__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
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
.pack__where {
  margin: 0 0 4px;
  text-align: center;
  font-size: 11.5px;
  color: var(--faint);
}
</style>
