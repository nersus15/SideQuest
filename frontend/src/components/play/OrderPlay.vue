<script setup>
// Rank a few things your way. Drag rows (mouse or touch), or focus a row
// and use the arrow keys. Rows glide into place.
import { computed, ref } from 'vue';
import { playFor } from '../../lib/play';
import { questPrompts, t } from '../../lib/i18n';
import { play as sound } from '../../lib/sound';

const props = defineProps({ quest: { type: Object, required: true } });
const s = computed(() => playFor(props.quest));
const options = computed(() => questPrompts(props.quest));

const list = ref(null);
const dragging = ref(-1); // option index being dragged
let rows = []; // row middles at drag start, by position

function move(from, to) {
  if (to < 0 || to >= s.value.order.length || to === from) return;
  const order = [...s.value.order];
  const [item] = order.splice(from, 1);
  order.splice(to, 0, item);
  s.value.order = order;
  sound('tick');
}

// The TransitionGroup ref is a component; its root is the <ol>.
const listEl = () => list.value?.$el ?? list.value;

function down(e, item) {
  dragging.value = item;
  try {
    listEl().setPointerCapture(e.pointerId); // keep tracking outside the list
  } catch {
    /* not every pointer can be captured */
  }
  rows = [...listEl().children].map((el) => {
    const r = el.getBoundingClientRect();
    return r.top + r.height / 2;
  });
}
function over(e) {
  if (dragging.value < 0) return;
  const from = s.value.order.indexOf(dragging.value);
  // The slot whose middle is closest to the pointer.
  let to = 0;
  rows.forEach((mid, i) => {
    if (Math.abs(e.clientY - mid) < Math.abs(e.clientY - rows[to])) to = i;
  });
  move(from, to);
}
function up() {
  dragging.value = -1;
}
function key(e, pos) {
  if (e.key === 'ArrowUp') {
    e.preventDefault();
    move(pos, pos - 1);
  } else if (e.key === 'ArrowDown') {
    e.preventDefault();
    move(pos, pos + 1);
  }
}
</script>

<template>
  <div class="order">
    <p class="order__hint">↕ {{ t('play.orderHint') }}</p>
    <TransitionGroup
      ref="list"
      tag="ol"
      name="row"
      class="rows"
      @pointermove="over"
      @pointerup="up"
      @pointercancel="up"
    >
      <li
        v-for="(item, pos) in s.order"
        :key="item"
        class="row"
        :class="{ 'row--drag': dragging === item }"
        tabindex="0"
        :aria-label="`${pos + 1}. ${options[item]}`"
        @pointerdown.prevent="down($event, item)"
        @keydown="key($event, pos)"
      >
        <span class="row__num">{{ pos + 1 }}</span>
        <span class="row__text">{{ options[item] }}</span>
        <span class="row__grip" aria-hidden="true">⠿</span>
      </li>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.order__hint {
  margin: 0 0 6px;
  font-size: 12px;
  font-weight: 700;
  color: var(--faint);
}
.rows {
  display: grid;
  gap: 5px;
  margin: 0;
  padding: 0;
  list-style: none;
  touch-action: none;
}
.row {
  display: flex;
  align-items: center;
  gap: 10px;
  height: 34px;
  padding: 0 10px 0 8px;
  border-radius: 12px;
  background: rgb(255 255 255 / 0.05);
  border: 1.5px solid var(--line);
  cursor: grab;
  user-select: none;
  transition: box-shadow 0.2s, border-color 0.2s, background 0.2s;
  outline: none;
}
.row:focus-visible {
  border-color: var(--gold);
}
.row--drag {
  cursor: grabbing;
  border-color: var(--accent, var(--violet));
  background: color-mix(in srgb, var(--accent, var(--violet)) 16%, var(--night-2));
  box-shadow: 0 6px 14px rgb(0 0 0 / 0.35);
  z-index: 1;
}
.row__num {
  display: grid;
  place-items: center;
  flex: none;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  font-size: 11.5px;
  font-weight: 900;
  color: var(--night);
  background: var(--accent, var(--violet));
}
.row__text {
  flex: 1;
  min-width: 0;
  font-size: 14px;
  font-weight: 700;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.row__grip {
  font-size: 16px;
  color: var(--faint);
}
.row-move {
  transition: transform 0.22s var(--ease);
}
</style>
