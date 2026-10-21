// Browser-only stand-in for the Go backend, used by `npm run dev` without
// Wails. It plays the real built-in quests with simplified rules; the real
// rules live in Go. Preview tricks: ?anchor=br|tr|bl, ?level=5, ?fresh (intro).
import pack from '../../../internal/quest/builtin/quests.json';

const QUESTS = pack.quests.map((q) => ({ ...q, twists: q.twists ?? [], source: 'builtin' }));

const BADGES = [
  ['first-step', '🐣'], ['doodler', '🎨'], ['zen-bean', '🌿'], ['butterfly', '🦋'], ['pathfinder', '🔭'],
  ['player-one', '🎮'], ['gremlin', '🦄'], ['twisty', '🌀'], ['spicy', '🦁'], ['all-rounder', '🌈'], ['main-character', '⭐'],
].map(([id, emoji]) => ({ id, emoji, name: id, hint: '' }));

const LOOKS = [
  ['hat-violet', 'hat', 1], ['hat-mint', 'hat', 2], ['glasses', 'extra', 3], ['hat-coral', 'hat', 4],
  ['scarf', 'extra', 5], ['hat-gold', 'hat', 6], ['flower', 'extra', 7], ['crown', 'hat', 8],
].map(([id, kind, level]) => ({ id, kind, level }));

const span = (n) => 50 + 25 * (n - 1);
function level(xp) {
  let n = 1;
  let rest = xp;
  while (rest >= span(n)) rest -= span(n++);
  return { number: n, title: '', xpInLevel: rest, xpToNext: span(n) };
}
const maxDifficulty = (n) => (n >= 5 ? 3 : n >= 3 ? 2 : 1);
const xpForLevel = (n) => Array.from({ length: n - 1 }, (_, i) => span(i + 1)).reduce((a, b) => a + b, 0);

export function createMock() {
  const params = new URLSearchParams(location.search);
  const mine = [];
  const journal = [];
  const s = {
    xp: xpForLevel(Number(params.get('level')) || 1),
    completed: 0,
    badges: new Set(),
    prefs: { muted: [], avoid: [] },
    ratings: {},
    nudges: { breakMinutes: 0, stayWhenBusy: false },
    cur: null,
    lang: params.has('fresh') ? '' : 'id',
    onboarded: !params.has('fresh'),
    look: { hat: 'hat-violet', extra: '' },
  };

  const pick = (exclude) => {
    const max = maxDifficulty(level(s.xp).number);
    const pool = QUESTS.filter(
      (q) => q.id !== exclude && q.difficulty <= max && !s.prefs.muted.includes(q.category)
        && !(q.needs ?? []).some((n) => s.prefs.avoid.includes(n)),
    );
    const q = pool[Math.floor(Math.random() * pool.length)] ?? QUESTS[0];
    s.cur = { q, twist: q.twists[Math.floor(Math.random() * q.twists.length)] ?? null, twistOn: false, accepted: false };
  };
  // ?quest=<id> starts on a specific quest.
  const first = QUESTS.find((q) => q.id === params.get('quest'));
  if (first) s.cur = { q: first, twist: first.twists[0] ?? null, twistOn: false, accepted: false };

  const view = () => {
    if (!s.cur) pick();
    const c = s.cur;
    const lvl = level(s.xp);
    const max = maxDifficulty(lvl.number);
    return {
      quest: c.q,
      twist: c.twist,
      twistOn: c.twistOn,
      accepted: c.accepted,
      reward: c.q.xp + (c.twistOn && c.twist ? c.twist.bonus : 0),
      xp: s.xp,
      level: lvl,
      completed: s.completed,
      badges: BADGES.map((b) => ({ ...b, earned: s.badges.has(b.id) })),
      prefs: s.prefs,
      recent: [],
      lang: s.lang,
      onboarded: s.onboarded,
      look: s.look,
      nudges: s.nudges,
      looks: LOOKS.map((l) => ({ ...l, unlocked: l.level <= lvl.number })),
      maxDifficulty: max,
      nextDifficultyAt: max === 1 ? 3 : max === 2 ? 5 : 0,
    };
  };

  const later = (v) => new Promise((r) => setTimeout(() => r(v), 60));

  return {
    State: () => later(view()),
    Reroll: () => (pick(s.cur?.q.id), later(view())),
    Skip: () => (pick(s.cur?.q.id), later(view())),
    Accept: () => ((s.cur.accepted = true), later(view())),
    Release: () => ((s.cur.accepted = false), later(view())),
    SetTwist: (on) => ((s.cur.twistOn = on), later(view())),
    SetPrefs: (p) => ((s.prefs = p), pick(s.cur?.q.id), later(view())),
    Complete: (input) => {
      if (!s.cur?.accepted) return Promise.reject('there is no accepted quest to complete');
      const done = s.cur.q;
      const gained = view().reward;
      const before = level(s.xp).number;
      s.xp += gained;
      s.completed++;
      const after = level(s.xp).number;
      const newBadges = [];
      if (!s.badges.has('first-step')) {
        s.badges.add('first-step');
        newBadges.push(BADGES[0]);
      }
      const entry = {
        id: String(Date.now()),
        questId: done.id,
        titles: { en: done.title, id: done.i18n?.id?.title ?? done.title },
        category: done.category,
        kind: done.interaction?.kind ?? '',
        xp: gained,
        at: new Date().toISOString(),
        answers: (input?.answers ?? []).filter(Boolean),
        count: input?.count ?? 0,
        doodle: input?.doodle ?? '',
        mood: '',
      };
      journal.unshift(entry);
      pick(done.id);
      return later({
        view: view(),
        done,
        rating: s.ratings[done.id] ?? '',
        gained,
        leveledUp: after > before,
        newBadges,
        newLooks: LOOKS.filter((l) => l.level > before && l.level <= after),
        newDifficulty: maxDifficulty(after) > maxDifficulty(before) ? maxDifficulty(after) : 0,
        entry: entry.id,
      });
    },
    Journal: () => {
      const moods = journal.filter((e) => e.mood);
      return later({ entries: journal, insight: { category: moods.length >= 3 ? moods[0].category : '', moods: moods.length } });
    },
    SetMood: (id, mood) => {
      const e = journal.find((j) => j.id === id);
      if (e) e.mood = mood;
      return later();
    },
    RateQuest: (id, rating) => {
      if (rating) s.ratings[id] = rating;
      else delete s.ratings[id];
      return later();
    },
    SetNudges: (n) => ((s.nudges = n), later(view())),
    Busy: () => later(params.has('busy')),
    SetLanguage: (lang) => ((s.lang = lang), later(view())),
    FinishOnboarding: () => ((s.onboarded = true), later(view())),
    SetLook: (look) => ((s.look = look), later(view())),
    Anchor: () => {
      const a = params.get('anchor') ?? '';
      return later({ right: a.includes('r'), bottom: a.includes('b') });
    },
    Expand: () => later(),
    Collapse: () => later(),
    Hide: () => later(),
    CustomQuests: () => later([...mine].reverse()),
    CustomReward: (m, d) => later(Math.min(40, 8 + 2 * m + 6 * (d - 1))),
    SaveCustom: (d) => {
      if (!d.title.trim()) return Promise.reject('custom.noTitle');
      const q = {
        ...d,
        id: d.id || `my-${Date.now()}`,
        title: d.title.trim(),
        description: d.description.trim() || 'A quest you made yourself. Go for it!',
        xp: Math.min(40, 8 + 2 * d.minutes + 6 * (d.difficulty - 1)),
        source: 'custom',
        twists: [],
      };
      const i = mine.findIndex((m) => m.id === q.id);
      i >= 0 ? (mine[i] = q) : mine.push(q);
      return later(q);
    },
    DeleteCustom: (id) => {
      const i = mine.findIndex((m) => m.id === id);
      if (i >= 0) mine.splice(i, 1);
      return later();
    },
    PlayQuest: (id) => {
      const q = mine.find((m) => m.id === id);
      if (!q) return Promise.reject('custom.notFound');
      s.cur = { q, twist: null, twistOn: false, accepted: false };
      return later(view());
    },
  };
}
