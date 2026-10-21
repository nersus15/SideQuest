// Tiny i18n: a reactive language and two dictionaries. Indonesian copy is
// casual and friendly ("kamu", "nggak", "yuk"), never stiff or over the top.
import { ref } from 'vue';

const en = {
  // capsule
  'capsule.hello': 'Hello!',
  'capsule.ready': 'Quest ready',
  'capsule.questing': 'Questing',
  'capsule.nudge': 'Quest time?',
  'capsule.open': 'Open SideQuest',
  'capsule.hint': 'Click to open · drag to move',

  // panel
  'panel.backpack': 'Backpack',
  'panel.shrink': 'Shrink (Esc)',
  'panel.loading': 'Boo is looking for a quest…',
  'panel.error': 'Hmm, that did not work. Try again?',
  'action.go': "Let's go!",
  'action.reroll': 'Reroll',
  'action.rerollHint': 'Show me a different quest',
  'action.skip': 'Skip',
  'action.skipHint': 'Not my thing. Hide it for a while.',
  'action.done': 'I did it! 🎉',
  'action.later': 'Maybe later',
  'action.notYet.answer': 'Write at least one answer first',
  'action.notYet.checklist': 'Tick everything off first',
  'action.notYet.timer': 'Let the timer finish first',
  'action.notYet.counter': 'Tap at least once first',
  'action.notYet.doodle': 'Draw something first',

  // card
  'card.onQuest': 'On quest',
  'card.minutes': '{n} min',
  'card.twistAdd': 'Add a twist?',
  'card.twistOn': 'Twist on!',
  'card.mine': 'Made by you',
  'kind.answer': 'Write it here',
  'kind.checklist': 'Checklist',
  'kind.timer': 'With a timer',
  'kind.counter': 'Tap to count',
  'kind.doodle': 'Doodle it',
  'difficulty.1': 'Easy',
  'difficulty.2': 'Medium',
  'difficulty.3': 'Bold',

  // interactions
  'play.answerHint': 'Type here…',
  'play.timerStart': 'Start timer',
  'play.timerAgain': 'Again',
  'play.timerDone': 'Done! Nice.',
  'play.breathIn': 'Breathe in',
  'play.breathHold': 'Hold',
  'play.breathOut': 'Breathe out',
  'play.tap': 'Tap!',
  'play.undo': 'Undo',
  'play.clear': 'Clear',
  'play.drawHere': 'Draw here',

  // celebration
  'cheer.1': 'Nailed it!',
  'cheer.2': 'Legendary!',
  'cheer.3': 'Look at you go!',
  'cheer.4': 'Quest complete!',
  'cheer.5': 'Boo is proud!',
  'cheer.6': 'Tiny victory!',
  'party.levelUp': 'Level up! Hi, {title}!',
  'party.newBadge': 'New badge',
  'party.newBadges': '{n} new badges',
  'party.unlocked': 'Unlocked',
  'party.newDifficulty': '{name} quests are now open!',
  'party.next': 'Another one!',
  'party.enough': "I'm good for now",
  'party.moodAsk': 'How do you feel now?',
  'party.moodThanks': 'Noted. Thanks for sharing!',
  'party.rateAsk': 'Want more quests like this?',
  'party.rateUp': 'More like this',
  'party.rateDown': 'Less like this',
  'party.rateThanks': 'Got it, Boo will remember.',
  'mood.low': 'Drained',
  'mood.meh': 'Meh',
  'mood.good': 'Good',
  'mood.great': 'Great',

  // backpack
  'pack.title': 'Backpack',
  'pack.close': 'Close backpack',
  'pack.tab.boo': 'Boo',
  'pack.tab.journal': 'Journal',
  'pack.tab.mine': 'My quests',
  'pack.tab.settings': 'Settings',
  'pack.stats': '{done} quests · {xp} XP · {badges}/{total} badges',
  'pack.xpExplain': 'Finish quests to earn XP. Levelling up unlocks outfits for Boo and bolder quests. No streaks, no pressure.',
  'pack.outfit': "Boo's outfit",
  'pack.hats': 'Hats',
  'pack.extras': 'Extras',
  'pack.none': 'Nothing',
  'pack.lockedAt': 'Level {n}',
  'pack.difficulty': 'Quest levels',
  'pack.difficultyOpen': 'open',
  'pack.badges': 'Badges',
  'pack.badgeLocked': 'Locked · {hint}',
  'pack.vibes': 'Quest vibes',
  'pack.vibesHint': 'Tap to turn kinds of quests on or off.',
  'pack.rightNow': "Right now I'm…",
  'pack.language': 'Language',
  'pack.tutorial': 'Show the intro again',
  'pack.hide': 'Hide SideQuest for now',
  'pack.whereMac': 'Boo waits in the menu bar. Quit SideQuest from there.',
  'pack.whereWin': 'Boo waits in the system tray. Quit SideQuest from there.',

  // journal
  'journal.empty': 'Your answers, doodles and moods land here after each quest.',
  'journal.insight': 'You usually feel best after {cat} quests {emoji}',
  'journal.insightSoon': 'Share how you feel after quests and Boo will spot what lifts you up.',
  'journal.private': 'Only you can see this. It stays on this device.',
  'journal.count': 'Count: {n}',
  'journal.today': 'Today',
  'journal.yesterday': 'Yesterday',

  // my quests
  'mine.new': '+ Make a quest',
  'mine.empty': 'Got a fun idea? Turn it into a quest and Boo will offer it to you.',
  'mine.edit': 'Edit “{title}”',
  'mine.play': 'Play now',
  'mine.delete': 'Delete',
  'mine.deleteConfirm': 'Delete?',
  'mine.deleteHint': 'Tap again to delete',
  'mine.deleteFailed': 'Could not delete that quest.',
  'mine.playFailed': 'Could not start that quest.',

  // editor
  'editor.new': 'Make a quest',
  'editor.edit': 'Edit quest',
  'editor.back': 'Back',
  'editor.what': "What's the quest?",
  'editor.whatPh': 'e.g. Draw my cat as a knight',
  'editor.details': 'Any details?',
  'editor.optional': 'optional',
  'editor.detailsPh': 'A tip, a rule, a silly twist…',
  'editor.kind': 'Kind',
  'editor.long': 'How long?',
  'editor.hard': 'How hard?',
  'editor.reward': 'Reward',
  'editor.save': 'Save quest',
  'editor.saving': 'Saving…',
  'editor.saved': 'Quest saved!',
  'editor.savedHint': 'Boo will mix it into your quests.',
  'editor.tryNow': 'Try it now',
  'editor.backToPack': 'Back to backpack',
  'editor.failed': 'Could not save that quest. Try again?',
  'custom.defaultDesc': 'A quest you made yourself. Go for it!',
  'custom.noTitle': 'Give your quest a name first.',
  'custom.titleLong': 'That name is a bit long. Keep it under 60 characters.',
  'custom.descLong': 'The details are a bit long. Keep them under 220 characters.',
  'custom.badCategory': 'Pick a kind of quest.',
  'custom.badMinutes': 'Pick how long it takes (1 to 15 minutes).',
  'custom.badDifficulty': 'Pick how hard it is.',
  'custom.tooMany': 'You already have 100 quests. Delete one to make room.',
  'custom.notFound': 'That quest no longer exists.',

  // onboarding
  'intro.hi': "Hi, I'm Boo!",
  'intro.hiText': "I'll pop by now and then with a tiny, fun side quest. Which language do you like?",
  'intro.how': 'How it works',
  'intro.howGo': 'Like it? Tap "Let\'s go!"',
  'intro.howReroll': 'Want another? Reroll',
  'intro.howSkip': 'Not your thing? Skip it',
  'intro.play': 'Play right here',
  'intro.playText': 'Many quests happen inside the app: write a short answer, tick a list, follow a timer, tap a counter or doodle. Everything lands in your Journal.',
  'intro.grow': 'Grow with Boo',
  'intro.growText': 'Quests give XP. Levelling up unlocks outfits for Boo and bolder quests. No streaks and no pressure: play whenever you feel like it.',
  'intro.next': 'Next',
  'intro.start': "Let's start!",
  'intro.skip': 'Skip intro',

  // categories, situations, levels, badges, looks
  'cat.creative': 'Creative',
  'cat.chill': 'Chill',
  'cat.social': 'Social',
  'cat.explore': 'Explore',
  'cat.play': 'Play',
  'cat.wildcard': 'Wildcard',
  'need.people': 'Flying solo',
  'need.sound': 'Keep it quiet',
  'need.move': 'Staying seated',
  'need.paper': 'No paper nearby',
  'need.phone': 'No phone',
  'need.outside': 'No window view',
  'level.1': 'Curious Blob',
  'level.2': 'Wanderer',
  'level.3': 'Side Kick',
  'level.4': 'Trailblazer',
  'level.5': 'Quest Goblin',
  'level.6': 'Daydream Ranger',
  'level.7': 'Chaos Wizard',
  'level.8': 'Legend of the Desk',
  'badge.first-step': 'First Step|Finish your first quest',
  'badge.doodler': 'Doodle Wizard|3 creative quests',
  'badge.zen-bean': 'Zen Bean|3 chill quests',
  'badge.butterfly': 'Social Butterfly|3 social quests',
  'badge.pathfinder': 'Pathfinder|3 explore quests',
  'badge.player-one': 'Player One|3 play quests',
  'badge.gremlin': 'Chaos Gremlin|3 wildcard quests',
  'badge.twisty': 'Twist Enjoyer|5 quests with a twist',
  'badge.spicy': 'Brave Heart|A bold quest',
  'badge.all-rounder': 'All-Rounder|One of every kind',
  'badge.main-character': 'Main Character|25 quests total',
  'look.hat-violet': 'Classic hat',
  'look.hat-mint': 'Mint hat',
  'look.glasses': 'Round glasses',
  'look.hat-coral': 'Coral hat',
  'look.scarf': 'Cozy scarf',
  'look.hat-gold': 'Golden hat',
  'look.flower': 'Little flower',
  'look.crown': 'Crown',
};

const id = {
  'capsule.hello': 'Halo!',
  'capsule.ready': 'Ada quest!',
  'capsule.questing': 'Lagi quest',
  'capsule.nudge': 'Quest yuk?',
  'capsule.open': 'Buka SideQuest',
  'capsule.hint': 'Klik buat buka · geser buat pindah',

  'panel.backpack': 'Tas',
  'panel.shrink': 'Kecilin (Esc)',
  'panel.loading': 'Boo lagi nyariin quest…',
  'panel.error': 'Hmm, gagal nih. Coba lagi ya?',
  'action.go': 'Yuk, mulai!',
  'action.reroll': 'Ganti',
  'action.rerollHint': 'Kasih quest yang lain',
  'action.skip': 'Lewati',
  'action.skipHint': 'Lagi nggak mood. Sembunyiin dulu.',
  'action.done': 'Udah! 🎉',
  'action.later': 'Nanti aja',
  'action.notYet.answer': 'Isi minimal satu jawaban dulu ya',
  'action.notYet.checklist': 'Centang semuanya dulu ya',
  'action.notYet.timer': 'Tunggu timernya selesai dulu ya',
  'action.notYet.counter': 'Tap minimal sekali dulu ya',
  'action.notYet.doodle': 'Gambar sesuatu dulu ya',

  'card.onQuest': 'Lagi jalan',
  'card.minutes': '{n} menit',
  'card.twistAdd': 'Mau tambah twist?',
  'card.twistOn': 'Twist aktif!',
  'card.mine': 'Bikinan kamu',
  'kind.answer': 'Tulis di sini',
  'kind.checklist': 'Checklist',
  'kind.timer': 'Pakai timer',
  'kind.counter': 'Tap-tap',
  'kind.doodle': 'Coret-coret',
  'difficulty.1': 'Santai',
  'difficulty.2': 'Sedang',
  'difficulty.3': 'Berani',

  'play.answerHint': 'Tulis di sini…',
  'play.timerStart': 'Mulai timer',
  'play.timerAgain': 'Ulang',
  'play.timerDone': 'Beres! Mantap.',
  'play.breathIn': 'Tarik napas',
  'play.breathHold': 'Tahan',
  'play.breathOut': 'Buang napas',
  'play.tap': 'Tap!',
  'play.undo': 'Batal',
  'play.clear': 'Hapus',
  'play.drawHere': 'Gambar di sini',

  'cheer.1': 'Keren!',
  'cheer.2': 'Mantap jiwa!',
  'cheer.3': 'Gokil, lancar banget!',
  'cheer.4': 'Quest beres!',
  'cheer.5': 'Boo bangga sama kamu!',
  'cheer.6': 'Menang kecil-kecilan!',
  'party.levelUp': 'Naik level! Halo, {title}!',
  'party.newBadge': 'Badge baru',
  'party.newBadges': '{n} badge baru',
  'party.unlocked': 'Kebuka',
  'party.newDifficulty': 'Quest {name} sekarang kebuka!',
  'party.next': 'Lagi dong!',
  'party.enough': 'Udah dulu deh',
  'party.moodAsk': 'Gimana rasanya sekarang?',
  'party.moodThanks': 'Oke, dicatat. Makasih udah cerita!',
  'party.rateAsk': 'Mau quest kayak gini lagi?',
  'party.rateUp': 'Mau lagi',
  'party.rateDown': 'Jarang aja',
  'party.rateThanks': 'Siap, Boo inget.',
  'mood.low': 'Capek',
  'mood.meh': 'Biasa',
  'mood.good': 'Oke',
  'mood.great': 'Seneng',

  'pack.title': 'Tas',
  'pack.close': 'Tutup tas',
  'pack.tab.boo': 'Boo',
  'pack.tab.journal': 'Jurnal',
  'pack.tab.mine': 'Quest-ku',
  'pack.tab.settings': 'Atur',
  'pack.stats': '{done} quest · {xp} XP · {badges}/{total} badge',
  'pack.xpExplain': 'Selesaiin quest buat dapet XP. Tiap naik level, ada outfit baru buat Boo dan quest yang lebih seru. Nggak ada streak, nggak ada tekanan.',
  'pack.outfit': 'Outfit Boo',
  'pack.hats': 'Topi',
  'pack.extras': 'Aksesori',
  'pack.none': 'Nggak pakai',
  'pack.lockedAt': 'Level {n}',
  'pack.difficulty': 'Level quest',
  'pack.difficultyOpen': 'kebuka',
  'pack.badges': 'Badge',
  'pack.badgeLocked': 'Belum kebuka · {hint}',
  'pack.vibes': 'Jenis quest',
  'pack.vibesHint': 'Tap buat nyalain atau matiin jenis quest.',
  'pack.rightNow': 'Sekarang aku lagi…',
  'pack.language': 'Bahasa',
  'pack.tutorial': 'Lihat perkenalan lagi',
  'pack.hide': 'Sembunyiin SideQuest dulu',
  'pack.whereMac': 'Boo nunggu di menu bar. Keluar dari SideQuest lewat situ.',
  'pack.whereWin': 'Boo nunggu di system tray. Keluar dari SideQuest lewat situ.',

  'journal.empty': 'Jawaban, gambar, dan mood kamu bakal kesimpen di sini tiap habis quest.',
  'journal.insight': 'Kamu biasanya paling happy habis quest {cat} {emoji}',
  'journal.insightSoon': 'Ceritain mood kamu tiap habis quest, nanti Boo kasih tahu quest mana yang paling bikin kamu happy.',
  'journal.private': 'Cuma kamu yang bisa lihat. Datanya nggak keluar dari perangkat ini.',
  'journal.count': 'Hitungan: {n}',
  'journal.today': 'Hari ini',
  'journal.yesterday': 'Kemarin',

  'mine.new': '+ Bikin quest',
  'mine.empty': 'Punya ide seru? Jadiin quest, nanti Boo yang nawarin ke kamu.',
  'mine.edit': 'Edit “{title}”',
  'mine.play': 'Main sekarang',
  'mine.delete': 'Hapus',
  'mine.deleteConfirm': 'Hapus?',
  'mine.deleteHint': 'Tap sekali lagi buat hapus',
  'mine.deleteFailed': 'Quest-nya gagal dihapus.',
  'mine.playFailed': 'Quest-nya gagal dimulai.',

  'editor.new': 'Bikin quest',
  'editor.edit': 'Edit quest',
  'editor.back': 'Balik',
  'editor.what': 'Quest-nya apa?',
  'editor.whatPh': 'misal: Gambar kucingku jadi ksatria',
  'editor.details': 'Ada detailnya?',
  'editor.optional': 'opsional',
  'editor.detailsPh': 'Tips, aturan, atau twist kocak…',
  'editor.kind': 'Jenis',
  'editor.long': 'Berapa lama?',
  'editor.hard': 'Seberapa susah?',
  'editor.reward': 'Hadiah',
  'editor.save': 'Simpan quest',
  'editor.saving': 'Nyimpen…',
  'editor.saved': 'Quest kesimpen!',
  'editor.savedHint': 'Boo bakal nyampurin ke quest kamu.',
  'editor.tryNow': 'Coba sekarang',
  'editor.backToPack': 'Balik ke tas',
  'editor.failed': 'Quest-nya gagal disimpan. Coba lagi ya?',
  'custom.defaultDesc': 'Quest bikinan kamu sendiri. Gas!',
  'custom.noTitle': 'Kasih nama quest-nya dulu ya.',
  'custom.titleLong': 'Namanya kepanjangan nih. Maksimal 60 karakter ya.',
  'custom.descLong': 'Detailnya kepanjangan nih. Maksimal 220 karakter ya.',
  'custom.badCategory': 'Pilih jenis quest-nya dulu.',
  'custom.badMinutes': 'Pilih berapa lama (1 sampai 15 menit).',
  'custom.badDifficulty': 'Pilih seberapa susah.',
  'custom.tooMany': 'Quest kamu udah 100. Hapus satu dulu biar ada tempat.',
  'custom.notFound': 'Quest-nya udah nggak ada.',

  'intro.hi': 'Hai, aku Boo!',
  'intro.hiText': 'Sesekali aku bakal mampir bawa quest kecil yang seru. Kamu lebih suka bahasa apa?',
  'intro.how': 'Cara mainnya',
  'intro.howGo': 'Suka? Tap "Yuk, mulai!"',
  'intro.howReroll': 'Mau yang lain? Ganti',
  'intro.howSkip': 'Lagi nggak mood? Lewati',
  'intro.play': 'Mainnya di sini',
  'intro.playText': 'Banyak quest yang bisa langsung kamu kerjain di sini: nulis jawaban singkat, centang daftar, ikutin timer, tap penghitung, atau coret-coret. Semuanya kesimpen di Jurnal.',
  'intro.grow': 'Tumbuh bareng Boo',
  'intro.growText': 'Tiap quest ngasih XP. Naik level bikin outfit Boo dan quest yang lebih seru kebuka. Nggak ada streak, nggak ada tekanan, main aja pas lagi pengen.',
  'intro.next': 'Lanjut',
  'intro.start': 'Yuk, mulai!',
  'intro.skip': 'Lewati perkenalan',

  'cat.creative': 'Kreatif',
  'cat.chill': 'Santai',
  'cat.social': 'Sosial',
  'cat.explore': 'Jelajah',
  'cat.play': 'Main',
  'cat.wildcard': 'Random',
  'need.people': 'Lagi sendirian',
  'need.sound': 'Harus tenang',
  'need.move': 'Nggak bisa gerak',
  'need.paper': 'Nggak ada kertas',
  'need.phone': 'Nggak pegang HP',
  'need.outside': 'Nggak ada jendela',
  'level.1': 'Blob Penasaran',
  'level.2': 'Pengembara',
  'level.3': 'Sidekick',
  'level.4': 'Pembuka Jalan',
  'level.5': 'Goblin Quest',
  'level.6': 'Ranger Lamunan',
  'level.7': 'Penyihir Chaos',
  'level.8': 'Legenda Meja Belajar',
  'badge.first-step': 'Langkah Pertama|Selesaiin quest pertama',
  'badge.doodler': 'Penyihir Coretan|3 quest kreatif',
  'badge.zen-bean': 'Si Paling Kalem|3 quest santai',
  'badge.butterfly': 'Kupu-kupu Sosial|3 quest sosial',
  'badge.pathfinder': 'Penjelajah|3 quest jelajah',
  'badge.player-one': 'Player One|3 quest main',
  'badge.gremlin': 'Gremlin Chaos|3 quest random',
  'badge.twisty': 'Pecinta Twist|5 quest pakai twist',
  'badge.spicy': 'Si Pemberani|Satu quest berani',
  'badge.all-rounder': 'Serba Bisa|Satu dari tiap jenis',
  'badge.main-character': 'Main Character|Total 25 quest',
  'look.hat-violet': 'Topi klasik',
  'look.hat-mint': 'Topi mint',
  'look.glasses': 'Kacamata bulat',
  'look.hat-coral': 'Topi koral',
  'look.scarf': 'Syal hangat',
  'look.hat-gold': 'Topi emas',
  'look.flower': 'Bunga kecil',
  'look.crown': 'Mahkota',
};

const dictionaries = { en, id };
export const LANGUAGES = [
  { code: 'id', label: 'Bahasa Indonesia', flag: '🇮🇩' },
  { code: 'en', label: 'English', flag: '🇬🇧' },
];

// Before the player picks, guess from the system language.
const guess = (typeof navigator !== 'undefined' && /^id\b|^in\b/i.test(navigator.language || '')) ? 'id' : 'en';
export const lang = ref(guess);

export function setLang(code) {
  if (dictionaries[code]) lang.value = code;
}

// t('key', { n: 3 }) → translated text with {n} filled in.
export function t(key, params) {
  const raw = dictionaries[lang.value]?.[key] ?? en[key] ?? key;
  if (!params) return raw;
  return raw.replace(/\{(\w+)\}/g, (_, k) => (params[k] ?? `{${k}}`));
}

export const has = (key) => key in en;

// Badges store "Name|hint" so both stay together in the dictionaries.
export function badgeText(id, fallback) {
  const raw = t(`badge.${id}`);
  if (raw === `badge.${id}`) return { name: fallback?.name ?? id, hint: fallback?.hint ?? '' };
  const [name, hint] = raw.split('|');
  return { name, hint };
}

export function levelTitle(n) {
  return t(`level.${Math.min(Math.max(n, 1), 8)}`);
}

// Translated error from the backend, which sends keys like "custom.noTitle".
export function errorText(err, fallbackKey) {
  const key = typeof err === 'string' ? err : err?.message;
  return key && has(key) ? t(key) : t(fallbackKey);
}

// --- quest text ---

const tr = (q) => q?.i18n?.[lang.value];

export function questTitle(q) {
  return tr(q)?.title || q?.title || '';
}

export function questDesc(q) {
  if (q?.source === 'custom' && q.description === en['custom.defaultDesc']) return t('custom.defaultDesc');
  return tr(q)?.description || q?.description || '';
}

export function questPrompts(q) {
  const base = q?.interaction?.prompts ?? [];
  const local = tr(q)?.prompts;
  return base.map((p, i) => local?.[i] || p);
}

export function twistText(q, twist) {
  return (twist && tr(q)?.twists?.[twist.id]) || twist?.text || '';
}

// Journal entries keep titles per language.
export function entryTitle(e) {
  return e?.titles?.[lang.value] || e?.titles?.en || '';
}
