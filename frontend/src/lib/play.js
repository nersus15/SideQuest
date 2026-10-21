// What the player has typed, ticked, counted, drawn, picked or timed for the
// current quest. Kept outside the components so closing the panel mid-quest
// loses nothing.
import { reactive } from 'vue';
import { fillStory, questPrompts } from './i18n';
import { answerOK, checkOf } from './checks';

export const play = reactive({
  questId: '',
  answers: [],
  checked: [],
  count: 0,
  doodle: '',
  timerDone: false,
  choice: -1, // choice: picked option
  storyShown: false, // madlib: story revealed
  millis: 0, // stopwatch: last stopped time
  spin: -1, // spinner: landed option
  order: [], // order: option indexes in the player's order
  quiz: -1, // quiz: the right option, once found
  quizWrong: [], // quiz: wrong options tried so far
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
    choice: -1,
    storyShown: false,
    millis: 0,
    spin: -1,
    order: [...Array(n).keys()],
    quiz: -1,
    quizWrong: [],
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
      // Checked quests need every answer right; others need at least one.
      if (checkOf(q)) return s.answers.every((_, i) => answerOK(q, s.answers, i));
      return s.answers.some((a) => a.trim());
    case 'checklist':
      return s.checked.every(Boolean);
    case 'timer':
      return s.timerDone;
    case 'counter':
      return s.count > 0;
    case 'doodle':
      return !!s.doodle;
    case 'choice':
      return s.choice >= 0;
    case 'madlib':
      return s.storyShown;
    case 'stopwatch':
      return s.millis > 0;
    case 'spinner':
      return s.spin >= 0;
    case 'quiz':
      return s.quiz >= 0;
    default:
      return true; // "do it" quests, and order (any order is a fine answer)
  }
}

// Madlibs reveal their story with the main button before completing.
export function storyPending(q) {
  const s = playFor(q);
  return q?.interaction?.kind === 'madlib' && !s.storyShown;
}
export function storyFilled(q) {
  return playFor(q).answers.every((a) => a.trim());
}
export function revealStory(q) {
  if (storyFilled(q)) playFor(q).storyShown = true;
}

// The payload for Complete(), in the language the player is using.
export function playInput(q) {
  const s = playFor(q);
  const kind = q?.interaction?.kind;
  const opts = questPrompts(q);
  const input = { answers: [], count: 0, doodle: '', story: '', millis: 0 };
  switch (kind) {
    case 'answer':
      input.answers = s.answers.map((a) => a.trim());
      break;
    case 'counter':
      input.count = s.count;
      break;
    case 'doodle':
      input.doodle = s.doodle;
      break;
    case 'choice':
      input.answers = [opts[s.choice]];
      break;
    case 'madlib':
      input.answers = s.answers.map((a) => a.trim());
      input.story = fillStory(q, s.answers).map((p) => p.text).join('');
      break;
    case 'stopwatch':
      input.millis = s.millis;
      break;
    case 'spinner':
      input.answers = [opts[s.spin]];
      break;
    case 'order':
      input.answers = s.order.map((i) => opts[i]);
      break;
    case 'quiz':
      input.answers = [opts[s.quiz]];
      input.count = s.quizWrong.length; // tries before the right answer
      break;
  }
  return input;
}
