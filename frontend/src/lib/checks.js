// Answer checks, mirroring internal/quest/check.go so the player gets instant
// feedback. The backend checks again before a quest counts as done.

// Letters and digits only, lower-cased.
const letters = (s) => (s ?? '').toLowerCase().replace(/[^\p{L}\p{N}]/gu, '');
const first = (s) => [...letters(s)][0] ?? '';
const last = (s) => [...letters(s)].at(-1) ?? '';
const reversed = (s) => [...letters(s)].reverse().join('');

export const checkOf = (q) => (q?.interaction?.kind === 'answer' ? q.interaction.check ?? '' : '');

// Whether answer i is right. Empty answers are never right.
export function answerOK(q, answers, i) {
  const a = answers[i] ?? '';
  if (!a.trim()) return false;
  switch (checkOf(q)) {
    case 'initials':
      return first(a) === first(q.interaction.prompts[i]);
    case 'chain':
      return i === 0 ? !!first(a) : !!last(answers[i - 1]) && first(a) === last(answers[i - 1]);
    case 'reverse':
      return i === 0 ? letters(a).length >= 2 : letters(a).length >= 2 && letters(a) === reversed(answers[0]);
    default:
      return true;
  }
}

// What answer i should start with, for hints and placeholders ('' if none).
export function neededStart(q, answers, i) {
  switch (checkOf(q)) {
    case 'initials':
      return first(q.interaction.prompts[i]).toUpperCase();
    case 'chain':
      return i > 0 ? last(answers[i - 1]).toUpperCase() : '';
    default:
      return '';
  }
}
