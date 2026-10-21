// What the player has typed, ticked, counted or drawn for the current quest.
// Kept outside the components so closing the panel mid-quest loses nothing.
import { reactive } from 'vue';

export const play = reactive({
  questId: '',
  answers: [],
  checked: [],
  count: 0,
  doodle: '',
  timerDone: false,
});

// Resets the state when a different quest comes up.
export function playFor(q) {
  if (play.questId === q?.id) return play;
  const n = q?.interaction?.prompts?.length ?? 0;
  Object.assign(play, {
    questId: q?.id ?? '',
    answers: Array(n).fill(''),
    checked: Array(n).fill(false),
    count: 0,
    doodle: '',
    timerDone: false,
  });
  return play;
}

export function clearPlay() {
  play.questId = '';
}

// Whether the quest has been "played" enough to complete it.
export function isReady(q) {
  const s = playFor(q);
  switch (q?.interaction?.kind) {
    case 'answer':
      return s.answers.some((a) => a.trim());
    case 'checklist':
      return s.checked.every(Boolean);
    case 'timer':
      return s.timerDone;
    case 'counter':
      return s.count > 0;
    case 'doodle':
      return !!s.doodle;
    default:
      return true;
  }
}

// The payload for Complete().
export function playInput(q) {
  const s = playFor(q);
  return {
    answers: q?.interaction?.kind === 'answer' ? s.answers.map((a) => a.trim()) : [],
    count: q?.interaction?.kind === 'counter' ? s.count : 0,
    doodle: q?.interaction?.kind === 'doodle' ? s.doodle : '',
  };
}
