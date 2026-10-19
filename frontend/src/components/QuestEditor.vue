<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { CATEGORIES, CATEGORY_KEYS, DIFFICULTY } from '../lib/categories';
import { api } from '../lib/api';

const props = defineProps({
  // The quest being edited, or null for a new one.
  quest: { type: Object, default: null },
});
const emit = defineEmits(['saved', 'cancel', 'play']);

const TIMES = [1, 2, 3, 5, 10, 15];
const DEFAULT_DESC = 'A quest you made yourself. Go for it!';

const draft = reactive({
  id: props.quest?.id ?? '',
  title: props.quest?.title ?? '',
  // Don't show the placeholder text as if the player had typed it.
  description: props.quest?.description === DEFAULT_DESC ? '' : props.quest?.description ?? '',
  category: props.quest?.category ?? 'wildcard',
  minutes: props.quest?.minutes ?? 2,
  difficulty: props.quest?.difficulty ?? 1,
});

const error = ref('');
const saving = ref(false);
const saved = ref(null);
const reward = ref(0);
const titleEl = ref(null);

const canSave = computed(() => draft.title.trim().length > 0 && !saving.value);

watch(
  () => [draft.minutes, draft.difficulty],
  async ([m, d]) => {
    try {
      reward.value = await api.customReward(m, d);
    } catch {
      reward.value = 0;
    }
  },
  { immediate: true },
);

onMounted(() => titleEl.value?.focus());

async function save() {
  if (!canSave.value) return;
  saving.value = true;
  error.value = '';
  try {
    saved.value = await api.saveCustom({ ...draft });
    emit('saved', saved.value);
  } catch (err) {
    error.value = typeof err === 'string' ? err : 'Could not save that quest. Try again?';
  } finally {
    saving.value = false;
  }
}
</script>

<template>
  <section class="editor" aria-label="Quest editor">
    <header class="editor__head">
      <button class="icon-btn" aria-label="Back" @click="$emit('cancel')">
        <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true"><path d="M10 3.5L5.5 8l4.5 4.5" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" /></svg>
      </button>
      <h2>{{ quest ? 'Edit quest' : 'Make a quest' }}</h2>
    </header>

    <div v-if="saved" class="done">
      <p class="done__emoji" aria-hidden="true">✨</p>
      <h3>Quest saved!</h3>
      <p class="done__title">“{{ saved.title }}”</p>
      <p class="done__hint">Boo will mix it into your quests.</p>
      <div class="done__actions">
        <button class="btn btn--primary" @click="$emit('play', saved.id)">Try it now</button>
        <button class="btn btn--ghost" @click="$emit('cancel')">Back to backpack</button>
      </div>
    </div>

    <form v-else class="editor__form" @submit.prevent="save">
      <div class="editor__scroll">
        <label class="field">
          <span class="field__label">What's the quest? <em>{{ draft.title.length }}/60</em></span>
          <input
            ref="titleEl"
            v-model="draft.title"
            class="input"
            maxlength="60"
            placeholder="e.g. Draw my cat as a knight"
            @keydown.esc.stop="$emit('cancel')"
          />
        </label>

        <label class="field">
          <span class="field__label">Any details? <em>optional</em></span>
          <textarea
            v-model="draft.description"
            class="input input--area"
            maxlength="220"
            rows="2"
            placeholder="A tip, a rule, a silly twist…"
            @keydown.esc.stop="$emit('cancel')"
          />
        </label>

        <div class="field">
          <span class="field__label">Kind</span>
          <div class="chips">
            <button
              v-for="key in CATEGORY_KEYS"
              :key="key"
              type="button"
              class="pick"
              :class="{ on: draft.category === key }"
              :style="{ '--accent': CATEGORIES[key].color }"
              :aria-pressed="draft.category === key"
              @click="draft.category = key"
            >
              {{ CATEGORIES[key].emoji }} {{ CATEGORIES[key].label }}
            </button>
          </div>
        </div>

        <div class="field">
          <span class="field__label">How long?</span>
          <div class="chips">
            <button
              v-for="t in TIMES"
              :key="t"
              type="button"
              class="pick pick--small"
              :class="{ on: draft.minutes === t }"
              :aria-pressed="draft.minutes === t"
              @click="draft.minutes = t"
            >
              {{ t }} min
            </button>
          </div>
        </div>

        <div class="field">
          <span class="field__label">How hard?</span>
          <div class="chips">
            <button
              v-for="d in 3"
              :key="d"
              type="button"
              class="pick pick--small"
              :class="{ on: draft.difficulty === d }"
              :aria-pressed="draft.difficulty === d"
              @click="draft.difficulty = d"
            >
              {{ DIFFICULTY[d] }}
            </button>
          </div>
        </div>
      </div>

      <p v-if="error" class="editor__error" role="alert">{{ error }}</p>

      <div class="editor__foot">
        <span class="editor__xp" title="XP comes from how long and how hard it is">✦ {{ reward }} XP</span>
        <button type="submit" class="btn btn--primary" :disabled="!canSave">
          {{ saving ? 'Saving…' : quest ? 'Save changes' : 'Save quest' }}
        </button>
      </div>
    </form>
  </section>
</template>

<style scoped>
.editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}
.editor__head {
  display: flex;
  align-items: center;
  gap: 4px;
  margin: 0 0 6px -6px;
}
.editor__head h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 900;
}
.editor__form {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
}
.editor__scroll {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 4px 5px 4px 4px; /* room for the input focus ring */
  margin: -4px -5px 0 -4px;
  scrollbar-width: thin;
  scrollbar-color: var(--night-3) transparent;
}

.field {
  display: grid;
  gap: 6px;
  margin-bottom: 12px;
}
.field__label {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--faint);
}
.field__label em {
  font-style: normal;
  font-weight: 700;
  text-transform: none;
  letter-spacing: 0;
}
.input {
  width: 100%;
  padding: 9px 12px;
  border-radius: 12px;
  border: 1.5px solid var(--line);
  background: var(--night-2);
  color: var(--ghost);
  font: inherit;
  font-weight: 700;
  font-size: 14px;
  outline: none;
  user-select: text;
  -webkit-user-select: text;
  transition: border-color 0.2s, box-shadow 0.2s;
}
.input::placeholder {
  color: var(--faint);
  font-weight: 600;
}
.input:focus {
  border-color: var(--violet);
  box-shadow: 0 0 0 3px rgb(139 108 255 / 0.2);
}
.input--area {
  resize: none;
  line-height: 1.35;
}

.chips {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.pick {
  padding: 5px 10px;
  border-radius: 99px;
  font-size: 12.5px;
  font-weight: 800;
  color: var(--muted);
  border: 1.5px solid var(--line);
  transition: transform 0.2s var(--spring), background 0.2s, border-color 0.2s, color 0.2s;
}
.pick:active {
  transform: scale(0.94);
}
.pick.on {
  color: var(--ghost);
  border-color: var(--accent, var(--violet));
  background: color-mix(in srgb, var(--accent, var(--violet)) 22%, transparent);
}
.pick--small {
  padding: 4px 10px;
}

.editor__error {
  margin: 6px 0 0;
  padding: 7px 10px;
  border-radius: 10px;
  font-size: 12.5px;
  font-weight: 800;
  color: #2a0f17;
  background: var(--coral);
}
.editor__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding-top: 10px;
}
.editor__xp {
  font-size: 14px;
  font-weight: 900;
  color: var(--gold);
}

.done {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  text-align: center;
}
.done__emoji {
  margin: 18px 0 0;
  font-size: 40px;
  animation: pop 0.5s var(--spring) both;
}
.done h3 {
  margin: 4px 0 0;
  font-size: 22px;
  font-weight: 900;
}
.done__title {
  margin: 6px 0 0;
  font-weight: 800;
  color: var(--muted);
}
.done__hint {
  margin: 4px 0 0;
  font-size: 13px;
  color: var(--faint);
}
.done__actions {
  display: grid;
  gap: 6px;
  width: 100%;
  margin-top: auto;
}

@keyframes pop {
  from { transform: scale(0.4); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
</style>
