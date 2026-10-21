// Display metadata for quest categories and needs. Keys match the Go side;
// labels come from the i18n dictionaries.
import { t } from './i18n';

export const CATEGORIES = {
  creative: { emoji: '🎨', color: 'var(--c-creative)' },
  chill: { emoji: '🌿', color: 'var(--c-chill)' },
  social: { emoji: '💬', color: 'var(--c-social)' },
  explore: { emoji: '🔭', color: 'var(--c-explore)' },
  play: { emoji: '🎲', color: 'var(--c-play)' },
  wildcard: { emoji: '🦄', color: 'var(--c-wildcard)' },
};

export const CATEGORY_KEYS = Object.keys(CATEGORIES);

// "Right now I can't…" toggles map to needs the quest would require.
export const SITUATIONS = [
  { need: 'people', emoji: '🙈' },
  { need: 'sound', emoji: '🤫' },
  { need: 'move', emoji: '🪑' },
  { need: 'paper', emoji: '📄' },
  { need: 'phone', emoji: '📵' },
  { need: 'outside', emoji: '🪟' },
];

export const difficultyName = (d) => t(`difficulty.${d}`);
export const categoryName = (key) => t(`cat.${key}`);

export function category(key) {
  const c = CATEGORIES[key] ?? { emoji: '✨', color: 'var(--violet)' };
  return { ...c, label: categoryName(key) };
}

export const MOODS = [
  { id: 'low', emoji: '😮‍💨' },
  { id: 'meh', emoji: '😐' },
  { id: 'good', emoji: '🙂' },
  { id: 'great', emoji: '😄' },
];
