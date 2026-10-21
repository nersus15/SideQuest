// Thin wrapper over the Wails bindings. In development, when the page runs in
// a plain browser (vite without Wails), it falls back to a tiny in-memory
// mock so the UI can still be previewed. Production builds never use it.
import * as Go from '../../wailsjs/go/main/App';

const hasWails = () => typeof window !== 'undefined' && !!window.go?.main?.App;

// The mock is imported lazily and only in development, so it (and the quest
// data it reads) never ships in a release build.
let mock;
async function backend() {
  if (!import.meta.env.DEV || hasWails()) return Go;
  mock ??= import('./mock').then((m) => m.createMock());
  return mock;
}

export const api = {
  state: async () => (await backend()).State(),
  reroll: async () => (await backend()).Reroll(),
  skip: async () => (await backend()).Skip(),
  accept: async () => (await backend()).Accept(),
  release: async () => (await backend()).Release(),
  setTwist: async (on) => (await backend()).SetTwist(on),
  complete: async (input) => (await backend()).Complete(input),
  setPrefs: async (prefs) => (await backend()).SetPrefs(prefs),
  anchor: async () => (await backend()).Anchor(),
  expand: async () => (await backend()).Expand(),
  collapse: async () => (await backend()).Collapse(),
  hide: async () => (await backend()).Hide(),
  customQuests: async () => (await backend()).CustomQuests(),
  customReward: async (minutes, difficulty) => (await backend()).CustomReward(minutes, difficulty),
  saveCustom: async (draft) => (await backend()).SaveCustom(draft),
  deleteCustom: async (id) => (await backend()).DeleteCustom(id),
  playQuest: async (id) => (await backend()).PlayQuest(id),
  journal: async () => (await backend()).Journal(),
  setMood: async (entry, mood) => (await backend()).SetMood(entry, mood),
  rateQuest: async (id, rating) => (await backend()).RateQuest(id, rating),
  setLanguage: async (lang) => (await backend()).SetLanguage(lang),
  finishOnboarding: async () => (await backend()).FinishOnboarding(),
  setLook: async (look) => (await backend()).SetLook(look),
};

export function onEvent(name, fn) {
  if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
    return window.runtime.EventsOn(name, fn);
  }
  return () => {};
}
