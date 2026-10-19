<script setup>
import { computed } from 'vue';
import { CATEGORIES, CATEGORY_KEYS, SITUATIONS } from '../lib/categories';

const props = defineProps({
  view: { type: Object, required: true },
});
const emit = defineEmits(['prefs', 'close', 'quit']);

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

      <button class="btn btn--ghost pack__quit" @click="$emit('quit')">Close SideQuest</button>
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
.pack__quit {
  display: flex;
  margin: 14px auto 0;
}
</style>
