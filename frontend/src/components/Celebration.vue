<script setup>
import { computed, onMounted, ref } from 'vue';
import Mascot from './Mascot.vue';
import { MOODS, difficultyName } from '../lib/categories';
import { badgeText, levelTitle, t } from '../lib/i18n';

const props = defineProps({
  result: { type: Object, required: true },
});
const emit = defineEmits(['next', 'close', 'mood', 'rate']);

const level = computed(() => props.result.view.level);
const look = computed(() => props.result.view.look ?? {});
const cheer = props.result.leveledUp
  ? t('party.levelUp', { title: levelTitle(props.result.view.level.number) })
  : t(`cheer.${1 + Math.floor(Math.random() * 6)}`);

// Everything new in one row: badges, Boo outfits and quest levels.
const rewards = computed(() => {
  const r = props.result;
  const out = (r.newBadges ?? []).map((b) => ({ key: `b-${b.id}`, emoji: b.emoji, text: badgeText(b.id, b).name }));
  for (const l of r.newLooks ?? []) out.push({ key: `l-${l.id}`, emoji: '👕', text: t(`look.${l.id}`) });
  if (r.newDifficulty) out.push({ key: 'd', emoji: '🔓', text: t('party.newDifficulty', { name: difficultyName(r.newDifficulty) }) });
  return out;
});
const rewardsLabel = computed(() => {
  const r = props.result;
  if ((r.newLooks?.length ?? 0) + (r.newDifficulty ? 1 : 0) > 0) return t('party.unlocked');
  return rewards.value.length === 1 ? t('party.newBadge') : t('party.newBadges', { n: rewards.value.length });
});

// Thumbs up or down tunes how often quests like this come up. Tapping the
// chosen one again clears it.
const rating = ref(props.result.rating ?? '');
const rated = ref(false);
function rate(r) {
  rating.value = rating.value === r ? '' : r;
  rated.value = true;
  emit('rate', rating.value);
}

// Optional mood check-in; it only ever feeds the player's own journal.
const mood = ref('');
function pickMood(id) {
  mood.value = id;
  emit('mood', id);
}

const shownXP = ref(0);
const barPct = ref(0);

// A handful of confetti bits with randomised flight paths.
const COLORS = ['var(--gold)', 'var(--violet)', 'var(--c-creative)', 'var(--c-chill)', 'var(--c-explore)'];
const confetti = Array.from({ length: 22 }, (_, i) => {
  const angle = (i / 22) * Math.PI * 2 + Math.random() * 0.4;
  const dist = 90 + Math.random() * 70;
  return {
    '--x': `${Math.cos(angle) * dist}px`,
    '--y': `${Math.sin(angle) * dist - 30}px`,
    '--r': `${Math.random() * 540 - 270}deg`,
    '--d': `${Math.random() * 0.15}s`,
    background: COLORS[i % COLORS.length],
    borderRadius: i % 3 === 0 ? '50%' : '2px',
  };
});

onMounted(() => {
  const target = props.result.gained;
  const start = performance.now();
  const tick = (now) => {
    const p = Math.min(1, (now - start) / 700);
    shownXP.value = Math.round(target * (1 - Math.pow(1 - p, 3)));
    if (p < 1) requestAnimationFrame(tick);
  };
  requestAnimationFrame(tick);
  // Let the bar animate from empty on a level-up, or from its old spot otherwise.
  const l = level.value;
  const before = props.result.leveledUp ? 0 : Math.max(0, l.xpInLevel - target);
  barPct.value = (before / l.xpToNext) * 100;
  setTimeout(() => (barPct.value = (l.xpInLevel / l.xpToNext) * 100), 250);
});
</script>

<template>
  <section class="party" aria-live="polite">
    <div class="party__burst" aria-hidden="true">
      <i v-for="(c, i) in confetti" :key="i" :style="c" />
    </div>

    <div class="party__scroll">
      <Mascot :size="44" mood="happy" :hat="look.hat" :extra="look.extra" class="party__boo" />
      <h2 class="party__cheer">{{ cheer }}</h2>
      <p class="party__xp">+{{ shownXP }} XP</p>

      <div class="party__level" :class="{ 'party__level--up': result.leveledUp }">
        <div class="party__lvrow">
          <span>Lv {{ level.number }} · {{ levelTitle(level.number) }}</span>
          <span class="party__lvnum">{{ level.xpInLevel }}/{{ level.xpToNext }}</span>
        </div>
        <div class="bar"><span :style="{ width: barPct + '%' }" /></div>
      </div>

      <div v-if="rewards.length" class="party__many">
        <small>{{ rewardsLabel }}</small>
        <ul>
          <li v-for="(r, i) in rewards" :key="r.key" :style="{ animationDelay: 0.5 + i * 0.12 + 's' }">
            {{ r.emoji }} {{ r.text }}
          </li>
        </ul>
      </div>

      <div class="mood">
        <p class="mood__ask">{{ mood ? t('party.moodThanks') : t('party.moodAsk') }}</p>
        <div class="mood__row">
          <button
            v-for="m in MOODS"
            :key="m.id"
            class="mood__btn"
            :class="{ on: mood === m.id, dim: mood && mood !== m.id }"
            :aria-pressed="mood === m.id"
            :title="t(`mood.${m.id}`)"
            @click="pickMood(m.id)"
          >
            <span aria-hidden="true">{{ m.emoji }}</span>
            <small>{{ t(`mood.${m.id}`) }}</small>
          </button>
        </div>
      </div>

      <div class="rate">
        <span class="rate__ask">{{ rated && rating ? t('party.rateThanks') : t('party.rateAsk') }}</span>
        <button
          class="rate__btn"
          :class="{ on: rating === 'up' }"
          :aria-pressed="rating === 'up'"
          :title="t('party.rateUp')"
          @click="rate('up')"
        >
          👍 <span>{{ t('party.rateUp') }}</span>
        </button>
        <button
          class="rate__btn"
          :class="{ on: rating === 'down' }"
          :aria-pressed="rating === 'down'"
          :title="t('party.rateDown')"
          @click="rate('down')"
        >
          👎 <span>{{ t('party.rateDown') }}</span>
        </button>
      </div>
    </div>

    <div class="party__actions">
      <button class="btn btn--primary" @click="$emit('next')">{{ t('party.next') }}</button>
      <button class="btn btn--ghost" @click="$emit('close')">{{ t('party.enough') }}</button>
    </div>
  </section>
</template>

<style scoped>
.party {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
  min-height: 0;
  padding: 6px 4px 0;
  text-align: center;
}
.party__scroll {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 100%;
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  scrollbar-width: none;
}
.party__burst {
  position: absolute;
  top: 56px;
  left: 50%;
  pointer-events: none;
}
.party__burst i {
  position: absolute;
  width: 8px;
  height: 8px;
  opacity: 0;
  animation: fly 1.1s var(--ease) var(--d) forwards;
}
.party__boo {
  animation: pop 0.5s var(--spring) both;
}
.party__cheer {
  margin: 0;
  font-size: 22px;
  text-wrap: balance;
  font-weight: 900;
}
.party__xp {
  margin: 0;
  font-size: 27px;
  line-height: 1.15;
  font-weight: 1000;
  color: var(--gold);
  text-shadow: 0 0 24px rgb(255 203 87 / 0.45);
  font-variant-numeric: tabular-nums;
}
.party__level {
  width: 100%;
  margin-top: 6px;
  padding: 8px 14px;
  border-radius: 16px;
  background: var(--night-2);
  border: 1px solid var(--line);
}
.party__lvrow {
  display: flex;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 800;
  margin-bottom: 6px;
}
.party__lvnum {
  color: var(--faint);
  font-variant-numeric: tabular-nums;
}
.party__level--up {
  border-color: rgb(255 203 87 / 0.5);
  animation: glow 1.2s ease-out 0.4s both;
}
.party__many {
  width: 100%;
  margin-top: 8px;
}
.party__many small {
  font-size: 10px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--gold);
}
.party__many ul {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 5px;
  margin: 6px 0 0;
  padding: 0;
  list-style: none;
}
.party__many li {
  padding: 4px 10px;
  border-radius: 99px;
  font-size: 12px;
  font-weight: 800;
  background: rgb(255 203 87 / 0.1);
  border: 1px solid rgb(255 203 87 / 0.35);
  animation: pop 0.5s var(--spring) both;
}
.mood {
  width: 100%;
  margin-top: 8px;
}
.mood__ask {
  margin: 0 0 6px;
  font-size: 12.5px;
  font-weight: 800;
  color: var(--muted);
}
.mood__row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;
  padding: 2px 4px;
}
.mood__btn {
  display: grid;
  justify-items: center;
  gap: 1px;
  padding: 4px 0 3px;
  border-radius: 12px;
  background: var(--night-2);
  border: 1px solid var(--line);
  transition: transform 0.2s var(--spring), opacity 0.2s, border-color 0.2s;
}
.mood__btn span {
  font-size: 18px;
  line-height: 1.1;
}
.mood__btn small {
  font-size: 10.5px;
  font-weight: 800;
  color: var(--faint);
}
.mood__btn:hover {
  transform: translateY(-2px);
}
.mood__btn.on {
  border-color: var(--gold);
  background: rgb(255 203 87 / 0.1);
  transform: scale(1.06);
}
.mood__btn.dim {
  opacity: 0.45;
}
.rate {
  display: flex;
  align-items: center;
  gap: 6px;
  width: 100%;
  margin-top: 6px;
  padding: 0 4px;
}
.rate__ask {
  flex: 1;
  min-width: 0;
  text-align: left;
  font-size: 12px;
  font-weight: 800;
  color: var(--muted);
}
.rate__btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  flex: none;
  padding: 4px 9px;
  border-radius: 99px;
  font-size: 13px;
  background: var(--night-2);
  border: 1px solid var(--line);
  transition: border-color 0.2s, background 0.2s, transform 0.2s var(--spring);
}
.rate__btn span {
  font-size: 11px;
  font-weight: 800;
  color: var(--muted);
}
.rate__btn:active {
  transform: scale(0.94);
}
.rate__btn.on {
  border-color: var(--gold);
  background: rgb(255 203 87 / 0.1);
}
.rate__btn.on span {
  color: var(--ghost);
}
.party__actions {
  display: grid;
  gap: 6px;
  width: 100%;
  padding-top: 4px;
}

@keyframes fly {
  0% { transform: translate(0, 0) rotate(0); opacity: 1; }
  80% { opacity: 1; }
  100% { transform: translate(var(--x), var(--y)) rotate(var(--r)); opacity: 0; }
}
@keyframes glow {
  from { box-shadow: 0 0 0 0 rgb(255 203 87 / 0.6); }
  to { box-shadow: 0 0 0 10px rgb(255 203 87 / 0); }
}
@keyframes pop {
  from { transform: scale(0.4); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}
</style>
