<script setup>
import { computed } from 'vue';
import Mascot from './Mascot.vue';
import { category } from '../lib/categories';

const props = defineProps({
  view: { type: Object, default: null },
  // true when Boo wants to gently suggest a quest
  nudge: { type: Boolean, default: false },
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
  if (v.accepted) return { text: 'On a quest', emoji: category(v.quest.category).emoji };
  if (props.nudge) return { text: 'Psst… quest?', emoji: '' };
  return { text: 'Quest ready', emoji: '' };
});
</script>

<template>
  <div class="capsule-wrap">
    <div
      class="capsule"
      :class="{ 'capsule--nudge': nudge, 'capsule--active': view?.accepted }"
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
        <Mascot :size="44" :mood="nudge ? 'wink' : 'idle'" />
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
  padding: var(--gutter);
}
.capsule {
  --wails-draggable: drag;
  position: relative;
  width: 184px; /* window is 208×84 incl. gutter, see app.go */
  height: 60px;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px 0 8px;
  border-radius: 999px;
  background:
    radial-gradient(120% 140% at 0% 0%, rgb(139 108 255 / 0.45), transparent 60%),
    linear-gradient(180deg, var(--night-3), var(--night));
  border: 1px solid var(--line);
  box-shadow:
    0 6px 16px rgb(10 4 30 / 0.45),
    inset 0 1px 0 rgb(255 255 255 / 0.08);
  cursor: pointer;
  outline: none;
  transition: transform 0.25s var(--spring), box-shadow 0.25s var(--ease);
  animation: capsule-in 0.45s var(--spring) both;
}
.capsule:hover {
  transform: translateY(-1px) scale(1.02);
  box-shadow: 0 8px 20px rgb(10 4 30 / 0.5), 0 0 0 3px rgb(139 108 255 / 0.25);
}
.capsule:active {
  transform: scale(0.97);
}
.capsule:focus-visible {
  box-shadow: 0 0 0 3px var(--gold);
}

.capsule__boo {
  position: relative;
  flex: none;
  margin-top: -6px;
}
.capsule:hover .capsule__boo {
  animation: wiggle 0.5s var(--ease);
}
.capsule__spark {
  position: absolute;
  top: -2px;
  right: -4px;
  color: var(--gold);
  font-size: 12px;
  animation: twinkle 1.6s ease-in-out infinite;
}

.capsule__text {
  display: grid;
  gap: 4px;
  min-width: 0;
}
.capsule__status {
  font-weight: 800;
  font-size: 14px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.capsule__emoji {
  margin-left: 4px;
}
.capsule__level {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 800;
  color: var(--muted);
}
.capsule__bar {
  width: 64px;
  height: 5px;
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
  animation: capsule-in 0.45s var(--spring) both, nudge 2.4s ease-in-out 0.5s infinite;
}
.capsule--active {
  border-color: rgb(255 203 87 / 0.35);
}

@keyframes capsule-in {
  from { transform: scale(0.6); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
@keyframes nudge {
  0%, 100% { box-shadow: 0 6px 16px rgb(10 4 30 / 0.45), 0 0 0 0 rgb(255 203 87 / 0); }
  50% { box-shadow: 0 6px 16px rgb(10 4 30 / 0.45), 0 0 0 4px rgb(255 203 87 / 0.28); }
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
