<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import Capsule from './components/Capsule.vue';
import QuestCard from './components/QuestCard.vue';
import Celebration from './components/Celebration.vue';
import Backpack from './components/Backpack.vue';
import QuestEditor from './components/QuestEditor.vue';
import Mascot from './components/Mascot.vue';
import Onboarding from './components/Onboarding.vue';
import { api, onEvent } from './lib/api';
import { errorText, lang, levelTitle, setLang, t } from './lib/i18n';
import { clearPlay, isReady, playInput } from './lib/play';

// The capsule morphs into the panel and back:
//   capsule → opening (capsule pinned to its corner while the window makes room)
//           → panel (pill-shaped clip grows into the full panel)
//           → closing (panel shrinks back into the pill) → capsule
const mode = ref('capsule'); // capsule | opening | panel | closing
// The window corner the panel grows from; set by the backend.
const anchor = ref({ right: false, bottom: false });
// The pop-in animation only plays when the app starts, not after closing.
const capsuleIntro = ref(true);

const PILL = { w: 140, h: 46, r: 23 }; // keep in sync with Capsule.vue
// Clip shape matching the capsule's spot inside the panel, for the morph.
const pillClip = computed(() => {
  const { right, bottom } = anchor.value;
  const x = `calc(100% - ${PILL.w}px)`;
  const y = `calc(100% - ${PILL.h}px)`;
  return `inset(${bottom ? y : '0px'} ${right ? '0px' : x} ${bottom ? '0px' : y} ${right ? x : '0px'} round ${PILL.r}px)`;
});

// Next painted frame, with a timeout fallback for when the webview is hidden.
const frame = () =>
  new Promise((done) => {
    const t = setTimeout(done, 50);
    requestAnimationFrame(() => {
      clearTimeout(t);
      done();
    });
  });
const sheet = ref(null); // null | 'backpack' | 'editor' | 'intro'
const mine = ref([]); // the player's own quests
const editing = ref(null); // quest being edited, null for a new one
const editorKey = ref(0);
const view = ref(null);
const celebration = ref(null);
const busy = ref(false);
const toast = ref('');
const nudge = ref(''); // '' | 'quest' (now and then) | 'break' (break reminder)
// True while a fullscreen app, game or presentation is in front.
const screenBusy = ref(false);
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
let hideAfterClose = false;

const level = computed(() => view.value?.level ?? { number: 1, title: '', xpInLevel: 0, xpToNext: 50 });
const look = computed(() => view.value?.look ?? {});
// The saved language wins over the system guess.
watch(view, (v) => v?.lang && setLang(v.lang));
// Quests with an in-app part must be played before they can be completed.
const ready = computed(() => isReady(view.value?.quest));
const notReadyHint = computed(() => (ready.value ? '' : t(`action.notYet.${view.value?.quest?.interaction?.kind}`)));
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
    say(t('panel.error'));
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
  nudge.value = '';
  capsuleIntro.value = false;
  mode.value = 'opening';
  if (view.value && !view.value.onboarded) sheet.value = 'intro';
  refresh();
  try {
    // Pin the capsule to the corner the window will grow from, so it stays
    // put on screen while the window resizes underneath it.
    anchor.value = await api.anchor();
    await frame();
    await api.expand();
  } catch (err) {
    console.error(err); // the panel still works in a smaller window
  }
  // Let the webview pick up its new size before the morph starts.
  await frame();
  mode.value = 'panel';
}

function close() {
  if (mode.value !== 'panel') return;
  sheet.value = null;
  mode.value = 'closing'; // the panel shrinks into the pill, then afterPanelLeave
}

async function afterPanelLeave() {
  // The capsule has faded in over the shrinking pill; now the window
  // shrinks around it.
  mode.value = 'capsule';
  celebration.value = null;
  await frame();
  await api.collapse().catch(console.error);
  scheduleNudge();
  if (hideAfterClose) {
    hideAfterClose = false;
    api.hide();
  }
}

// Tuck Boo away; the menu bar / tray icon brings it back.
function hideBoo() {
  if (mode.value === 'panel') {
    hideAfterClose = true;
    close();
  } else if (mode.value === 'capsule') {
    api.hide();
  }
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
  if (!ready.value) return;
  const input = playInput(view.value.quest);
  const res = await run(() => api.complete(input));
  if (!res) return;
  clearPlay();
  celebration.value = res;
  view.value = res.view;
}

function rateDone(rating) {
  const id = celebration.value?.done?.id;
  if (id) api.rateQuest(id, rating).catch(console.error);
}

function setMood(mood) {
  if (celebration.value?.entry) api.setMood(celebration.value.entry, mood).catch(console.error);
}

// --- language, intro and outfit ---

async function chooseLanguage(code) {
  setLang(code); // instant; the backend also renames the tray menu
  const v = await run(() => api.setLanguage(code));
  if (v) view.value = v;
}

async function finishIntro() {
  sheet.value = null;
  try {
    view.value = await api.finishOnboarding();
    if (!view.value.lang) chooseLanguage(lang.value);
  } catch (err) {
    console.error(err);
  }
}

async function setLook(l) {
  const v = await run(() => api.setLook(l));
  if (v) view.value = v;
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

// Boo perks up to suggest a quest: every N minutes when the break reminder
// is on, otherwise now and then at random. No sounds, no popups, and never
// while a fullscreen app, game or presentation is in front.
let nudgeTimer;
function scheduleNudge(minutes) {
  clearTimeout(nudgeTimer);
  const every = view.value?.nudges?.breakMinutes ?? 0;
  minutes ??= every || 20 + Math.random() * 25;
  nudgeTimer = setTimeout(() => {
    if (screenBusy.value) {
      scheduleNudge(5); // try again once the screen is free
      return;
    }
    if (mode.value === 'capsule' && !view.value?.accepted) nudge.value = every ? 'break' : 'quest';
  }, minutes * 60_000);
}

async function setNudges(n) {
  const v = await run(() => api.setNudges(n));
  if (!v) return;
  view.value = v;
  scheduleNudge(); // restart with the new interval
}

function onKey(e) {
  if (mode.value !== 'panel') return;
  if (e.key === 'Escape') {
    if (sheet.value === 'editor') sheet.value = 'backpack';
    else if (sheet.value) sheet.value = null;
    else close();
  }
}

// --- the player's own quests ---

async function loadMine() {
  try {
    mine.value = await api.customQuests();
  } catch (err) {
    console.error(err);
  }
}

function toggleBackpack() {
  if (sheet.value) {
    sheet.value = null;
  } else {
    sheet.value = 'backpack';
    loadMine();
  }
}

function editQuest(q) {
  editing.value = q;
  editorKey.value++;
  sheet.value = 'editor';
}

async function afterCustomChange() {
  const before = view.value?.quest?.id;
  await Promise.all([loadMine(), refresh()]);
  if (view.value?.quest?.id !== before) cardKey.value++;
}

async function deleteQuest(id) {
  try {
    await api.deleteCustom(id);
  } catch (err) {
    console.error(err);
    say(t('mine.deleteFailed'));
  }
  afterCustomChange();
}

async function playQuest(id) {
  try {
    view.value = await api.playQuest(id);
    celebration.value = null;
    sheet.value = null;
    rollDir.value = 'next';
    cardKey.value++;
  } catch (err) {
    console.error(err);
    say(errorText(err, 'mine.playFailed'));
  }
}


let offPoke;
let offHide;
let offBusy;
let offAnchor;
onMounted(async () => {
  await refresh();
  // First launch: open straight into the intro.
  if (view.value && !view.value.onboarded) open();
  scheduleNudge();
  window.addEventListener('keydown', onKey);
  offPoke = onEvent('sidequest:poke', open);
  offHide = onEvent('sidequest:hide', hideBoo);
  offBusy = onEvent('sidequest:busy', (b) => {
    screenBusy.value = !!b;
    if (b) nudge.value = '';
  });
  api.busy().then((b) => (screenBusy.value = !!b)).catch(() => {});
  // On Windows the window is always panel-sized and trimmed to the capsule,
  // so the capsule must sit in the right corner from the start.
  offAnchor = onEvent('sidequest:anchor', (a) => a && (anchor.value = a));
  api.anchor().then((a) => a && (anchor.value = a)).catch(() => {});
});
onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey);
  clearTimeout(nudgeTimer);
  offPoke?.();
  offHide?.();
  offBusy?.();
  offAnchor?.();
});
</script>

<template>
  <Transition name="capsule">
    <Capsule
      v-if="mode !== 'panel'"
      :view="view"
      :nudge="nudge"
      :intro="capsuleIntro"
      :anchor="anchor"
      @open="open"
    />
  </Transition>

  <Transition name="panel" :duration="{ enter: 520, leave: 340 }" @after-leave="afterPanelLeave">
    <div v-if="mode === 'panel'" class="panel-wrap" :style="{ '--pill-clip': pillClip }">
      <main class="panel">
        <header class="panel__head">
          <Mascot :size="34" :mood="celebration ? 'happy' : view?.accepted ? 'wink' : 'idle'" :hat="look.hat" :extra="look.extra" />
          <div class="panel__level">
            <div class="panel__lvrow">
              <strong>Lv {{ level.number }}</strong>
              <span>{{ levelTitle(level.number) }}</span>
            </div>
            <div class="bar" :title="`${level.xpInLevel} / ${level.xpToNext} XP`">
              <span :style="{ width: levelPct + '%' }" />
            </div>
          </div>
          <button
            class="icon-btn"
            :class="{ 'icon-btn--on': sheet }"
            :aria-label="t('panel.backpack')"
            :title="t('panel.backpack')"
            @click="toggleBackpack"
          >
            <span aria-hidden="true" style="font-size: 17px">🎒</span>
          </button>
          <button class="icon-btn" :aria-label="t('panel.shrink')" :title="t('panel.shrink')" @click="close">
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
              :mine="mine"
              @prefs="setPrefs"
              @close="sheet = null"
              @hide="hideBoo"
              @new="editQuest(null)"
              @edit="editQuest"
              @delete="deleteQuest"
              @play="playQuest"
              @look="setLook"
              @lang="chooseLanguage"
              @tutorial="sheet = 'intro'"
              @nudges="setNudges"
            />
          </Transition>
          <Transition name="sheet">
            <Onboarding
              v-if="sheet === 'intro'"
              class="panel__sheet"
              :look="look"
              @lang="chooseLanguage"
              @done="finishIntro"
            />
          </Transition>
          <Transition name="sheet">
            <QuestEditor
              v-if="sheet === 'editor'"
              :key="editorKey"
              class="panel__sheet"
              :quest="editing"
              @saved="afterCustomChange"
              @cancel="sheet = 'backpack'"
              @play="playQuest"
            />
          </Transition>

          <Celebration v-if="celebration" :result="celebration" @next="nextQuest" @close="close" @mood="setMood" @rate="rateDone" />

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
                {{ t('action.go') }}
              </button>
              <div class="actions__row">
                <button class="btn btn--soft" :disabled="busy" :title="t('action.rerollHint')" @click="reroll">
                  <span class="dice" :class="{ 'dice--spin': busy }" aria-hidden="true">🎲</span> {{ t('action.reroll') }}
                </button>
                <button class="btn btn--soft" :disabled="busy" :title="t('action.skipHint')" @click="skip">
                  {{ t('action.skip') }}
                </button>
              </div>
            </div>

            <div v-else class="actions">
              <p v-if="notReadyHint" class="actions__hint">{{ notReadyHint }} ☝️</p>
              <button class="btn btn--primary actions__main" :disabled="busy || !ready" :title="notReadyHint" @click="complete">
                {{ t('action.done') }}
              </button>
              <button class="btn btn--ghost" :disabled="busy" @click="release">{{ t('action.later') }}</button>
            </div>
          </template>

          <div v-else class="panel__loading">
            <Mascot :size="56" />
            <p>{{ t('panel.loading') }}</p>
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
  position: fixed;
  inset: 0;
  padding: var(--gutter);
  /* Follows the clip shape during the morph, unlike box-shadow. */
  filter: drop-shadow(0 4px 7px rgb(8 3 24 / 0.5));
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
.actions__hint {
  margin: -4px 0 -4px;
  text-align: center;
  font-size: 12px;
  font-weight: 800;
  color: var(--faint);
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

/* Morph: the panel is revealed from a pill-shaped clip at the capsule's
   corner, so the capsule appears to stretch open (and shrink back). The
   contents fade in once there is room for them. */
.panel-enter-from .panel,
.panel-leave-to .panel {
  clip-path: var(--pill-clip);
}
.panel-enter-to .panel,
.panel-leave-from .panel {
  clip-path: inset(0px 0px 0px 0px round var(--radius));
}
.panel-enter-active .panel {
  transition: clip-path 0.46s cubic-bezier(0.2, 0.9, 0.25, 1);
}
.panel-leave-active .panel {
  transition: clip-path 0.3s cubic-bezier(0.55, 0, 0.3, 1) 0.04s;
}
.panel-enter-active .panel__head,
.panel-enter-active .panel__body {
  transition: opacity 0.24s ease-out 0.16s, transform 0.34s var(--ease) 0.16s;
}
.panel-leave-active .panel__head,
.panel-leave-active .panel__body {
  transition: opacity 0.12s ease-in, transform 0.16s ease-in;
}
.panel-enter-from .panel__head,
.panel-enter-from .panel__body,
.panel-leave-to .panel__head,
.panel-leave-to .panel__body {
  opacity: 0;
  transform: translateY(6px) scale(0.98);
}

/* The capsule sits above the panel: it fades out over the growing panel,
   and fades back in over the last part of the shrink, so there is never a
   frame with nothing on screen. */
.capsule-enter-active {
  transition: opacity 0.16s ease-out 0.2s, transform 0.2s var(--ease) 0.2s;
}
.capsule-leave-active {
  transition: opacity 0.14s ease-in;
}
.capsule-enter-from {
  opacity: 0;
  transform: scale(0.97);
}
.capsule-leave-to {
  opacity: 0;
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
