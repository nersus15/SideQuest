// Thin wrapper over the Wails bindings. In development, when the page runs in
// a plain browser (vite without Wails), it falls back to a tiny in-memory
// mock so the UI can still be previewed. Production builds never use it.
import * as Go from '../../wailsjs/go/main/App';
import { createMock } from './mock';

const hasWails = () => typeof window !== 'undefined' && !!window.go?.main?.App;

let mock;
function backend() {
  if (!import.meta.env.DEV || hasWails()) return Go;
  mock ??= createMock();
  return mock;
}

export const api = {
  state: () => backend().State(),
  reroll: () => backend().Reroll(),
  skip: () => backend().Skip(),
  accept: () => backend().Accept(),
  release: () => backend().Release(),
  setTwist: (on) => backend().SetTwist(on),
  complete: () => backend().Complete(),
  setPrefs: (prefs) => backend().SetPrefs(prefs),
  anchor: () => backend().Anchor(),
  expand: () => backend().Expand(),
  collapse: () => backend().Collapse(),
  hide: () => backend().Hide(),
  customQuests: () => backend().CustomQuests(),
  customReward: (minutes, difficulty) => backend().CustomReward(minutes, difficulty),
  saveCustom: (draft) => backend().SaveCustom(draft),
  deleteCustom: (id) => backend().DeleteCustom(id),
  playQuest: (id) => backend().PlayQuest(id),
};

export function onEvent(name, fn) {
  if (typeof window !== 'undefined' && window.runtime?.EventsOn) {
    return window.runtime.EventsOn(name, fn);
  }
  return () => {};
}
