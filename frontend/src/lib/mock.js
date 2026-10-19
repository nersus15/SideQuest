// Browser-only stand-in for the Go backend. It mirrors the shape of
// game.View closely enough to develop the UI; the real rules live in Go.
const QUESTS = [
  { id: 'pocket-monster', title: 'Pocket monster', description: 'Doodle a tiny creature that lives in your pocket. Give it a name and one weird habit.', category: 'creative', minutes: 3, difficulty: 1, needs: ['paper'], xp: 15, twists: [{ id: 'other-hand', text: 'Use your other hand.', bonus: 5 }] },
  { id: 'square-breath', title: 'Square breathing', description: 'Breathe in for 4, hold for 4, out for 4, hold for 4. Four rounds.', category: 'chill', minutes: 1, difficulty: 1, xp: 10, twists: [] },
  { id: 'two-truths', title: 'Two truths, one lie', description: 'Play one round of two truths and a lie with someone.', category: 'social', minutes: 4, difficulty: 3, needs: ['people'], xp: 30, twists: [] },
  { id: 'color-quest', title: 'Color quest', description: 'Pick a color, then spot 5 things in that color around you.', category: 'explore', minutes: 2, difficulty: 1, xp: 12, twists: [{ id: 'ten', text: 'Find 10 instead of 5.', bonus: 8 }] },
  { id: 'pen-flip', title: 'Pen flip', description: 'Flip a pen so it lands standing up. You get ten tries.', category: 'play', minutes: 2, difficulty: 2, xp: 18, twists: [] },
  { id: 'documentary', title: 'Nature documentary', description: 'Narrate what you are doing for 30 seconds like a whispering wildlife documentary host.', category: 'wildcard', minutes: 1, difficulty: 2, needs: ['sound'], xp: 18, twists: [] },
];

const BADGES = [
  { id: 'first-step', name: 'First Step', emoji: '🐣', hint: 'Finish your first quest' },
  { id: 'doodler', name: 'Doodle Wizard', emoji: '🎨', hint: '3 creative quests' },
  { id: 'zen-bean', name: 'Zen Bean', emoji: '🌿', hint: '3 chill quests' },
  { id: 'spicy', name: 'Spice Taster', emoji: '🌶️', hint: 'A spicy quest' },
];

const TITLES = ['Curious Blob', 'Wanderer', 'Side Kick', 'Trailblazer'];

function level(xp) {
  let n = 1;
  let rest = xp;
  while (rest >= 50 + 25 * (n - 1)) {
    rest -= 50 + 25 * (n - 1);
    n++;
  }
  return { number: n, title: TITLES[Math.min(n - 1, TITLES.length - 1)], xpInLevel: rest, xpToNext: 50 + 25 * (n - 1) };
}

export function createMock() {
  const mine = [];
  const s = { xp: 0, completed: 0, badges: new Set(), prefs: { muted: [], avoid: [] }, cur: null, recent: [] };

  const pick = (exclude) => {
    const pool = QUESTS.filter((q) => q.id !== exclude && !s.prefs.muted.includes(q.category));
    const q = pool[Math.floor(Math.random() * pool.length)] ?? QUESTS[0];
    s.cur = { q, twist: q.twists[0] ?? null, twistOn: false, accepted: false };
  };

  const view = () => {
    if (!s.cur) pick();
    const c = s.cur;
    return {
      quest: c.q,
      twist: c.twist,
      twistOn: c.twistOn,
      accepted: c.accepted,
      reward: c.q.xp + (c.twistOn && c.twist ? c.twist.bonus : 0),
      xp: s.xp,
      level: level(s.xp),
      completed: s.completed,
      badges: BADGES.map((b) => ({ ...b, earned: s.badges.has(b.id) })),
      prefs: s.prefs,
      recent: s.recent,
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
    SetPrefs: (p) => ((s.prefs = p), s.prefs.muted.includes(s.cur?.q.category) && pick(s.cur.q.id), later(view())),
    Complete: () => {
      if (!s.cur?.accepted) return Promise.reject('there is no accepted quest to complete');
      const gained = view().reward;
      const before = level(s.xp).number;
      s.xp += gained;
      s.completed++;
      s.recent.unshift({ questId: s.cur.q.id, title: s.cur.q.title, category: s.cur.q.category, xp: gained });
      const newBadges = [];
      if (!s.badges.has('first-step')) {
        s.badges.add('first-step');
        newBadges.push(BADGES[0]);
      }
      pick(s.cur.q.id);
      return later({ view: view(), gained, leveledUp: level(s.xp).number > before, newBadges });
    },
    // ?anchor=br (or tr, bl) previews other screen corners.
    Anchor: () => {
      const a = new URLSearchParams(location.search).get('anchor') ?? '';
      return later({ right: a.includes('r'), bottom: a.includes('b') });
    },
    Expand: () => later(),
    Collapse: () => later(),
    Hide: () => later(),
    CustomQuests: () => later([...mine].reverse()),
    CustomReward: (m, d) => later(Math.min(40, 8 + 2 * m + 6 * (d - 1))),
    SaveCustom: (d) => {
      if (!d.title.trim()) return Promise.reject('Give your quest a name first.');
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
      if (!q) return Promise.reject('That quest no longer exists.');
      s.cur = { q, twist: null, twistOn: false, accepted: false };
      return later(view());
    },
  };
}
