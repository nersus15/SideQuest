// Display metadata for quest categories and needs. Keys match the Go side.
export const CATEGORIES = {
  creative: { label: 'Creative', emoji: '🎨', color: 'var(--c-creative)' },
  chill: { label: 'Chill', emoji: '🌿', color: 'var(--c-chill)' },
  social: { label: 'Social', emoji: '💬', color: 'var(--c-social)' },
  explore: { label: 'Explore', emoji: '🔭', color: 'var(--c-explore)' },
  play: { label: 'Play', emoji: '🎲', color: 'var(--c-play)' },
  wildcard: { label: 'Wildcard', emoji: '🦄', color: 'var(--c-wildcard)' },
};

export const CATEGORY_KEYS = Object.keys(CATEGORIES);

// "Right now I can't…" toggles map to needs the quest would require.
export const SITUATIONS = [
  { need: 'people', label: 'Flying solo', emoji: '🙈' },
  { need: 'sound', label: 'Keep it quiet', emoji: '🤫' },
  { need: 'move', label: 'Staying seated', emoji: '🪑' },
  { need: 'paper', label: 'No paper nearby', emoji: '📄' },
  { need: 'phone', label: 'No phone', emoji: '📵' },
  { need: 'outside', label: 'No window view', emoji: '🪟' },
];

export const DIFFICULTY = ['', 'Easy', 'Medium', 'Spicy'];

export function category(key) {
  return CATEGORIES[key] ?? { label: key, emoji: '✨', color: 'var(--violet)' };
}
