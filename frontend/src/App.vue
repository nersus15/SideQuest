<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import Capsule from './components/Capsule.vue';
import QuestCard from './components/QuestCard.vue';
import Celebration from './components/Celebration.vue';
import Backpack from './components/Backpack.vue';
import Mascot from './components/Mascot.vue';
import { api, onEvent } from './lib/api';

// capsule → (resize) → panel, and back. "moving" hides everything while the
// native window changes size so nothing visibly jumps.
const mode = ref('capsule'); // capsule | moving | panel
const sheet = ref(null); // null | 'backpack'
const view = ref(null);
const celebration = ref(null);
const busy = ref(false);
const toast = ref('');
const nudge = ref(false);
const cardKey = ref(0);
const cardEl = ref(null);
const cardOverflows = ref(false);

// Fade the bottom edge while there is more card to scroll to.
function checkOverflow() {
  const el = cardEl.value;
  cardOverflows.value = !!el && el.scrollTop + el.clientHeight < el.scrollHeight - 2;
}
watch([view, mode], () => nextTick(checkOverflow));
const rollDir = ref('next');

const level = computed(() => view.value?.level ?? { number: 1, title: '', xpInLevel: 0, xpToNext: 50 });
const levelPct = computed(() => (level.value.xpInLevel / level.value.xpToNext) * 100);

let toastTimer;
function say(msg) {
  toast.value = msg;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => (toast.value = ''), 2600);
}

// Runs a backend call, guarding against double clicks and surfacing errors gently.
async function run(fn) {
  if (busy.value) return null;
  busy.value = true;
  try {
    return await fn();
  } catch (err) {
    console.error(err);
    say('Hmm, that did not work. Try again?');
    refresh(); // resync in case the backend state moved on
    return null;
  } finally {
    busy.value = false;
  }
}

// Read-only, so it does not take the busy lock (buttons stay usable).
async function refresh() {
  try {
    view.value = await api.state();
  } catch (err) {
    console.error(err);
  }
}

async function open() {
  if (mode.value !== 'capsule') return;
  nudge.value = false;
  mode.value = 'moving';
  try {
    await api.expand();
  } catch (err) {
    console.error(err); // the panel still works in a smaller window
  }
  // Give the webview a frame to pick up its new size before animating in.
  requestAnimationFrame(() => (mode.value = 'panel'));
  refresh();
}

function close() {
  if (mode.value !== 'panel') return;
  sheet.value = null;
  mode.value = 'moving'; // panel leave transition runs, then collapse()
}

async function afterPanelLeave() {
  await api.collapse().catch(console.error);
  celebration.value = null;
  mode.value = 'capsule';
  scheduleNudge();
}

async function swap(fn, dir) {
  const v = await run(fn);
  if (!v) return;
  rollDir.value = dir;
  view.value = v;
  cardKey.value++;
}

const reroll = () => swap(api.reroll, 'next');
const skip = () => swap(api.skip, 'away');

async function accept() {
  const v = await run(api.accept);
  if (v) view.value = v;
}

async function release() {
  const v = await run(api.release);
  if (v) view.value = v;
}

async function toggleTwist() {
  const v = await run(() => api.setTwist(!view.value.twistOn));
  if (v) view.value = v;
}

async function complete() {
  const res = await run(api.complete);
  if (!res) return;
  celebration.value = res;
  view.value = res.view;
}

function nextQuest() {
  celebration.value = null;
  cardKey.value++;
}

async function setPrefs(prefs) {
  const v = await run(() => api.setPrefs(prefs));
  if (!v) return;
  if (v.quest?.id !== view.value?.quest?.id) cardKey.value++;
  view.value = v;
}

// Boo occasionally perks up to suggest a quest. No sounds, no popups.
let nudgeTimer;
function scheduleNudge() {
  clearTimeout(nudgeTimer);
  const minutes = 20 + Math.random() * 25;
  nudgeTimer = setTimeout(() => {
    if (mode.value === 'capsule' && !view.value?.accepted) nudge.value = true;
  }, minutes * 60_000);
}

function onKey(e) {
  if (mode.value !== 'panel') return;
  if (e.key === 'Escape') {
    if (sheet.value) sheet.value = null;
    else close();
  }
}

let offPoke;
onMounted(() => {
  refresh();
  scheduleNudge();
  window.addEventListener('keydown', onKey);
  offPoke = onEvent('sidequest:poke', open);
});
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  clearTimeout(nudgeTimer);
  offPoke?.();
});
</script>

<template>
  <Capsule v-if="mode === 'capsule'" :view="view" :nudge="nudge" @open="open" />

  <Transition name="panel" @after-leave="afterPanelLeave">
    <div v-if="mode === 'panel'" class="panel-wrap">
      <main class="panel">
        <header class="panel__head">
          <Mascot :size="34" :mood="celebration ? 'happy' : view?.accepted ? 'wink' : 'idle'" />
          <div class="panel__level">
            <div class="panel__lvrow">
              <strong>Lv {{ level.number }}</strong>
              <span>{{ level.title }}</span>
            </div>
            <div class="bar" :title="`${level.xpInLevel} / ${level.xpToNext} XP`">
              <span :style="{ width: levelPct + '%' }" />
            </div>
          </div>
          <button
            class="icon-btn"
            :class="{ 'icon-btn--on': sheet }"
            aria-label="Open backpack"
            title="Backpack"
            @click="sheet = sheet ? null : 'backpack'"
          >
            <span aria-hidden="true" style="font-size: 17px">🎒</span>
          </button>
          <button class="icon-btn" aria-label="Shrink to capsule" title="Shrink (Esc)" @click="close">
            <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true">
              <path d="M3.5 6l4.5 4.5L12.5 6" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" />
            </svg>
          </button>
        </header>

        <div class="panel__body">
          <Transition name="sheet">
            <Backpack
              v-if="sheet === 'backpack' && view"
              class="panel__sheet"
              :view="view"
              @prefs="setPrefs"
              @close="sheet = null"
              @quit="api.quit"
            />
          </Transition>

          <Celebration v-if="celebration" :result="celebration" @next="nextQuest" @close="close" />

          <template v-else-if="view?.quest">
            <div
              ref="cardEl"
              class="panel__card"
              :class="{ 'panel__card--more': cardOverflows }"
              @scroll.passive="checkOverflow"
            >
              <Transition :name="`card-${rollDir}`" mode="out-in" @after-enter="checkOverflow">
                <QuestCard :key="cardKey" :view="view" :busy="busy" @toggle-twist="toggleTwist" />
              </Transition>
            </div>

            <div v-if="!view.accepted" class="actions">
              <button class="btn btn--primary actions__main" :disabled="busy" @click="accept">
                Let's go!
              </button>
              <div class="actions__row">
                <button class="btn btn--soft" :disabled="busy" title="Show me a different quest" @click="reroll">
                  <span class="dice" :class="{ 'dice--spin': busy }" aria-hidden="true">🎲</span> Reroll
                </button>
                <button class="btn btn--soft" :disabled="busy" title="Not my thing. Hide it for a while." @click="skip">
                  Skip
                </button>
              </div>
            </div>

            <div v-else class="actions">
              <button class="btn btn--primary actions__main" :disabled="busy" @click="complete">
                I did it! 🎉
              </button>
              <button class="btn btn--ghost" :disabled="busy" @click="release">Maybe later</button>
            </div>
          </template>

          <div v-else class="panel__loading">
            <Mascot :size="56" />
            <p>Boo is looking for a quest…</p>
          </div>
        </div>

        <Transition name="toast">
          <p v-if="toast" class="toast" role="status">{{ toast }}</p>
        </Transition>
      </main>
    </div>
  </Transition>
</template>

<style scoped>
.panel-wrap {
  width: 100%;
  height: 100%;
  padding: var(--gutter);
}
.panel {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  border-radius: var(--radius);
  overflow: hidden;
  background:
    radial-gradient(2px 2px at 14% 22%, rgb(255 255 255 / 0.5), transparent),
    radial-gradient(1.5px 1.5px at 82% 12%, rgb(255 255 255 / 0.4), transparent),
    radial-gradient(1.5px 1.5px at 64% 88%, rgb(255 255 255 / 0.3), transparent),
    radial-gradient(120% 60% at 0% 0%, rgb(139 108 255 / 0.28), transparent 60%),
    linear-gradient(180deg, #1d1636, var(--night));
  border: 1px solid var(--line);
  box-shadow: 0 10px 24px rgb(8 3 24 / 0.55);
}

.panel__head {
  --wails-draggable: drag;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 10px 8px 14px;
  cursor: grab;
}
.panel__level {
  flex: 1;
  min-width: 0;
}
.panel__lvrow {
  display: flex;
  align-items: baseline;
  gap: 6px;
  margin-bottom: 4px;
  font-size: 13px;
  white-space: nowrap;
}
.panel__lvrow strong {
  font-weight: 900;
}
.panel__lvrow span {
  color: var(--muted);
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
}
.icon-btn--on {
  background: rgb(255 255 255 / 0.1);
}

.panel__body {
  position: relative;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 4px 14px 14px;
}
.panel__card {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  scrollbar-width: none;
  border-radius: 22px;
}
.panel__card--more {
  mask-image: linear-gradient(180deg, #000 82%, transparent);
}
.panel__sheet {
  position: absolute;
  inset: 0;
  z-index: 2;
  padding: 4px 14px 12px;
  background: linear-gradient(180deg, #1d1636, var(--night));
}
.panel__loading {
  display: grid;
  place-items: center;
  gap: 6px;
  margin: auto;
  color: var(--muted);
  font-weight: 700;
}

.actions {
  display: grid;
  gap: 10px;
  padding-top: 12px;
}
.actions__main {
  height: 48px;
  font-size: 17px;
}
.actions__row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 8px;
}
.dice {
  display: inline-block;
}
.dice--spin {
  animation: spin 0.5s linear infinite;
}

.toast {
  position: absolute;
  left: 50%;
  bottom: 14px;
  z-index: 5;
  margin: 0;
  padding: 8px 14px;
  border-radius: 12px;
  transform: translateX(-50%);
  white-space: nowrap;
  font-size: 13px;
  font-weight: 800;
  background: var(--coral);
  color: #2a0f17;
}

/* Panel opens with a soft spring, closes quickly. */
.panel-enter-active {
  transition: opacity 0.25s var(--ease), transform 0.4s var(--spring);
}
.panel-leave-active {
  transition: opacity 0.16s ease-in, transform 0.18s ease-in;
}
.panel-enter-from,
.panel-leave-to {
  opacity: 0;
  transform: scale(0.92) translateY(8px);
}

.sheet-enter-active,
.sheet-leave-active {
  transition: transform 0.3s var(--ease), opacity 0.2s;
}
.sheet-enter-from,
.sheet-leave-to {
  transform: translateY(24px);
  opacity: 0;
}

/* Reroll: the card slides sideways. Skip: it drops away. */
.card-next-enter-active,
.card-next-leave-active,
.card-away-enter-active,
.card-away-leave-active {
  transition: transform 0.22s var(--ease), opacity 0.18s;
}
.card-next-leave-to {
  transform: translateX(-30px) rotate(-3deg);
  opacity: 0;
}
.card-next-enter-from {
  transform: translateX(30px) rotate(3deg);
  opacity: 0;
}
.card-away-leave-to {
  transform: translateY(26px) scale(0.95);
  opacity: 0;
}
.card-away-enter-from {
  transform: translateY(-12px) scale(0.98);
  opacity: 0;
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.2s, transform 0.3s var(--spring);
}
.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translate(-50%, 8px);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
