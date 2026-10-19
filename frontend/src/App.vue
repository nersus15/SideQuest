<script setup>
import { ref, onMounted } from 'vue';
import {
  GetState,
  NextQuest,
  AcceptCurrent,
  SkipCurrent,
  CompleteCurrent,
  RerollQuest,
} from '../wailsjs/go/main/App';
import * as runtime from '../wailsjs/runtime';

const expanded = ref(false);
const quest = ref(null);
const xp = ref(0);
const level = ref(1);
const badges = ref([]);
const completedCount = ref(0);
const status = ref('Ready for a tiny adventure');
const dragState = ref({ active: false, pointerId: null, offsetX: 0, offsetY: 0 });

async function refreshState() {
  try {
    const state = await GetState();
    xp.value = state.xp ?? 0;
    level.value = state.level ?? 1;
    badges.value = state.badges ?? [];
    completedCount.value = state.completedCount ?? 0;
    if (state.quest) {
      quest.value = state.quest;
      expanded.value = true;
    }
  } catch (error) {
    console.error('GetState failed:', error);
  }
}

async function fetchNextQuest() {
  try {
    const result = await NextQuest();
    quest.value = result;
    expanded.value = true;
    status.value = 'Fresh quest unlocked';
  } catch (error) {
    console.error('NextQuest failed:', error);
  }
}

async function acceptQuest() {
  try {
    const result = await AcceptCurrent();
    quest.value = result;
    status.value = 'Quest accepted';
  } catch (error) {
    console.error('AcceptCurrent failed:', error);
  }
}

async function rerollQuest() {
  try {
    const result = await RerollQuest();
    quest.value = result;
    status.value = 'Another little challenge';
  } catch (error) {
    console.error('RerollQuest failed:', error);
  }
}

async function skipQuest() {
  try {
    const result = await SkipCurrent();
    quest.value = result;
    status.value = 'Quest passed for now';
  } catch (error) {
    console.error('SkipCurrent failed:', error);
  }
}

async function completeQuest() {
  try {
    const result = await CompleteCurrent();
    xp.value = result.progress?.xp ?? xp.value;
    level.value = result.progress?.level ?? level.value;
    badges.value = result.badges ?? badges.value;
    completedCount.value = Object.keys(result.progress?.completed ?? {}).length;
    quest.value = result.quest ?? quest.value;
    status.value = `Quest complete +${result.xp} XP`;
    expanded.value = false;
    await refreshState();
  } catch (error) {
    console.error('CompleteCurrent failed:', error);
  }
}

function collapse() {
  expanded.value = false;
  runtime.WindowSetSize(220, 92);
}

function expand() {
  if (!quest.value) {
    return fetchNextQuest();
  }
  expanded.value = true;
  runtime.WindowSetSize(370, 240);
}

function startDrag(event) {
  const position = runtime.WindowGetPosition();
  dragState.value = {
    active: true,
    pointerId: event.pointerId,
    offsetX: event.clientX - position[0],
    offsetY: event.clientY - position[1],
  };
  event.preventDefault();
  window.addEventListener('pointermove', handleDrag);
  window.addEventListener('pointerup', stopDrag);
}

function handleDrag(event) {
  if (!dragState.value.active) return;
  const nextX = event.clientX - dragState.value.offsetX;
  const nextY = event.clientY - dragState.value.offsetY;
  runtime.WindowSetPosition(Math.max(0, nextX), Math.max(0, nextY));
}

function stopDrag() {
  dragState.value.active = false;
  window.removeEventListener('pointermove', handleDrag);
  window.removeEventListener('pointerup', stopDrag);
}

onMounted(async () => {
  runtime.WindowSetAlwaysOnTop(true);
  runtime.WindowSetDarkTheme();
  runtime.WindowSetSize(220, 92);
  runtime.WindowSetMinSize(200, 80);
  await refreshState();
  if (!quest.value) {
    await fetchNextQuest();
  }
});
</script>

<template>
  <div class="app-shell">
    <div v-if="!expanded" class="collapsed-shell" @dblclick="expand">
      <div class="capsule-shell" @pointerdown="startDrag" @click="expand">
        <div class="capsule-dot" aria-hidden="true"></div>
        <div class="capsule-text">
          <strong>SideQuest</strong>
          <span>{{ xp }} XP</span>
        </div>
      </div>
    </div>

    <div v-else class="quest-card" @pointerdown.stop="() => {}">
      <div class="quest-header">
        <span class="quest-pill">{{ quest?.category ?? 'Quest' }}</span>
        <span class="level-badge">Lvl {{ level }}</span>
      </div>

      <div class="quest-body">
        <h2 class="quest-title">{{ quest?.title ?? 'Tiny mission' }}</h2>
        <p class="quest-description">{{ quest?.description ?? 'A little spark of fun is waiting.' }}</p>
        <div class="quest-meta">
          <span>{{ quest?.durationMinutes ?? 2 }} min</span>
          <span>{{ quest?.difficulty ?? 1 }} difficulty</span>
          <span>+{{ quest?.xp ?? 15 }} XP</span>
        </div>
      </div>

      <div class="quest-actions">
        <button class="action-btn primary" @click="completeQuest">Complete</button>
        <button class="action-btn secondary" @click="acceptQuest">Accept</button>
        <button class="action-btn secondary" @click="rerollQuest">Reroll</button>
        <button class="action-btn warning" @click="skipQuest">Skip</button>
      </div>

      <div class="quest-footer">
        <small>{{ status }}</small>
      </div>
    </div>
  </div>
</template>
