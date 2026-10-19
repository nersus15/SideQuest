<script setup>
import { computed } from 'vue';
import Mascot from './Mascot.vue';
import { category } from '../lib/categories';

const props = defineProps({
  view: { type: Object, default: null },
  // true when Boo wants to gently suggest a quest
  nudge: { type: Boolean, default: false },
  // play the pop-in animation (only when the app starts)
  intro: { type: Boolean, default: false },
  // window corner to stay pinned to while the window resizes
  anchor: { type: Object, default: () => ({ right: false, bottom: false }) },
});
const emit = defineEmits(['open']);

// Dragging moves the native window, so compare screen coordinates: if the
// pointer travelled, this "click" is the end of a drag, not a tap.
let down = null;
function onDown(e) {
  down = { x: e.screenX, y: e.screenY };
}
function onClick(e) {
  const moved = down && Math.hypot(e.screenX - down.x, e.screenY - down.y) > 4;
  down = null;
  if (!moved) emit('open');
}

const level = computed(() => props.view?.level ?? { number: 1, xpInLevel: 0, xpToNext: 50 });
const pct = computed(() => Math.min(100, (level.value.xpInLevel / level.value.xpToNext) * 100));

const status = computed(() => {
  const v = props.view;
  if (!v?.quest) return { text: 'Hello!', emoji: '' };
  if (v.accepted) return { text: 'Questing', emoji: category(v.quest.category).emoji };
  if (props.nudge) return { text: 'Quest time?', emoji: '' };
  return { text: 'Quest ready', emoji: '' };
});
</script>

<template>
  <div class="capsule-wrap" :class="{ 'is-right': anchor.right, 'is-bottom': anchor.bottom }">
    <div
      class="capsule"
      :class="{ 'capsule--nudge': nudge, 'capsule--active': view?.accepted, 'capsule--intro': intro }"
      role="button"
      tabindex="0"
      :aria-label="`${status.text}. Open SideQuest`"
      title="Click to open · drag to move"
      @pointerdown="onDown"
      @click="onClick"
      @keydown.enter.prevent="$emit('open')"
      @keydown.space.prevent="$emit('open')"
    >
      <span class="capsule__boo">
        <Mascot :size="34" :mood="nudge ? 'wink' : 'idle'" />
        <span v-if="nudge" class="capsule__spark" aria-hidden="true">✦</span>
      </span>
      <span class="capsule__text">
        <span class="capsule__status">
          {{ status.text }}<span v-if="status.emoji" class="capsule__emoji">{{ status.emoji }}</span>
        </span>
        <span class="capsule__level">
          <span class="capsule__lv">Lv {{ level.number }}</span>
          <span class="capsule__bar"><span :style="{ width: pct + '%' }" /></span>
        </span>
      </span>
    </div>
  </div>
</template>

<style scoped>
.capsule-wrap {
  position: fixed;
  z-index: 2; /* above the panel while they hand over */
  top: 0;
  left: 0;
  padding: var(--gutter);
  transform-origin: top left;
}
.capsule-wrap.is-right {
  left: auto;
  right: 0;
}
.capsule-wrap.is-bottom {
  top: auto;
  bottom: 0;
}
.capsule-wrap.is-right:not(.is-bottom) {
  transform-origin: top right;
}
.capsule-wrap.is-bottom:not(.is-right) {
  transform-origin: bottom left;
}
.capsule-wrap.is-right.is-bottom {
  transform-origin: bottom right;
}
.capsule {
  --wails-draggable: drag;
  position: relative;
  width: 140px; /* window is 156×62 incl. an 8px gutter, see app.go */
  height: 46px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 12px 0 6px;
  border-radius: 999px;
  background:
    radial-gradient(120% 140% at 0% 0%, rgb(139 108 255 / 0.45), transparent 60%),
    linear-gradient(180deg, var(--night-3), var(--night));
  border: 1px solid var(--line);
  box-shadow:
    0 3px 8px rgb(10 4 30 / 0.45),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  cursor: pointer;
  outline: none;
  transition: transform 0.25s var(--spring), box-shadow 0.25s var(--ease);
}
.capsule--intro {
  animation: capsule-in 0.45s var(--spring) both;
}
.capsule:hover {
  transform: scale(1.03);
  box-shadow: 0 3px 8px rgb(10 4 30 / 0.5), 0 0 0 2px rgb(139 108 255 / 0.3);
}
.capsule:active {
  transform: scale(0.97);
}
.capsule:focus-visible {
  box-shadow: 0 0 0 2px var(--gold);
}

.capsule__boo {
  position: relative;
  flex: none;
  margin-top: -4px;
}
.capsule:hover .capsule__boo {
  animation: wiggle 0.5s var(--ease);
}
.capsule__spark {
  position: absolute;
  top: -3px;
  right: -4px;
  color: var(--gold);
  font-size: 10px;
  animation: twinkle 1.6s ease-in-out infinite;
}

.capsule__text {
  display: grid;
  gap: 3px;
  min-width: 0;
}
.capsule__status {
  font-weight: 800;
  font-size: 12.5px;
  line-height: 1.1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.capsule__emoji {
  margin-left: 3px;
}
.capsule__level {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 10px;
  line-height: 1;
  font-weight: 800;
  color: var(--muted);
}
.capsule__bar {
  flex: 1;
  height: 4px;
  border-radius: 99px;
  background: rgb(255 255 255 / 0.1);
  overflow: hidden;
}
.capsule__bar span {
  display: block;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--violet), var(--gold));
  transition: width 0.8s var(--ease);
}

.capsule--nudge {
  animation: nudge 2.4s ease-in-out 0.5s infinite;
}
.capsule--active {
  border-color: rgb(255 203 87 / 0.35);
}

@keyframes capsule-in {
  from { transform: scale(0.6); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
@keyframes nudge {
  0%, 100% { box-shadow: 0 3px 8px rgb(10 4 30 / 0.45), 0 0 0 0 rgb(255 203 87 / 0); }
  50% { box-shadow: 0 3px 8px rgb(10 4 30 / 0.45), 0 0 0 3px rgb(255 203 87 / 0.3); }
}
@keyframes wiggle {
  0%, 100% { transform: rotate(0); }
  30% { transform: rotate(-8deg); }
  65% { transform: rotate(6deg); }
}
@keyframes twinkle {
  0%, 100% { transform: scale(0.8); opacity: 0.6; }
  50% { transform: scale(1.2); opacity: 1; }
}
</style>
