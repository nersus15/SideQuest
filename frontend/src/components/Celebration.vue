<script setup>
import { computed, onMounted, ref } from 'vue';
import Mascot from './Mascot.vue';

const props = defineProps({
  result: { type: Object, required: true },
});
defineEmits(['next', 'close']);

const CHEERS = ['Nailed it!', 'Legendary!', 'Look at you go!', 'Quest complete!', 'Boo is proud!', 'Tiny victory!'];
const cheer = props.result.leveledUp
  ? `Level up! Hi, ${props.result.view.level.title}!`
  : CHEERS[Math.floor(Math.random() * CHEERS.length)];

const level = computed(() => props.result.view.level);
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
  const tick = (t) => {
    const p = Math.min(1, (t - start) / 700);
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
      <Mascot :size="56" mood="happy" class="party__boo" />
      <h2 class="party__cheer">{{ cheer }}</h2>
      <p class="party__xp">+{{ shownXP }} XP</p>

      <div class="party__level" :class="{ 'party__level--up': result.leveledUp }">
        <div class="party__lvrow">
          <span>Lv {{ level.number }} · {{ level.title }}</span>
          <span class="party__lvnum">{{ level.xpInLevel }}/{{ level.xpToNext }}</span>
        </div>
        <div class="bar"><span :style="{ width: barPct + '%' }" /></div>
      </div>

      <ul v-if="result.newBadges?.length === 1" class="party__badges">
        <li v-for="b in result.newBadges" :key="b.id">
          <span class="party__badge-emoji">{{ b.emoji }}</span>
          <span>
            <small>New badge</small>
            <strong>{{ b.name }}</strong>
          </span>
        </li>
      </ul>
      <div v-else-if="result.newBadges?.length" class="party__many">
        <small>{{ result.newBadges.length }} new badges</small>
        <ul>
          <li v-for="(b, i) in result.newBadges" :key="b.id" :style="{ animationDelay: 0.5 + i * 0.12 + 's' }">
            {{ b.emoji }} {{ b.name }}
          </li>
        </ul>
      </div>
    </div>

    <div class="party__actions">
      <button class="btn btn--primary" @click="$emit('next')">Another one!</button>
      <button class="btn btn--ghost" @click="$emit('close')">I'm good for now</button>
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
  font-size: 30px;
  line-height: 1.2;
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
.party__badges {
  display: grid;
  gap: 6px;
  width: 100%;
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
}
.party__badges li {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 14px;
  text-align: left;
  background: rgb(255 203 87 / 0.1);
  border: 1px solid rgb(255 203 87 / 0.35);
  animation: pop 0.5s var(--spring) 0.5s both;
}
.party__badges small {
  display: block;
  font-size: 10px;
  font-weight: 900;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: var(--gold);
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
.party__badge-emoji {
  font-size: 24px;
}
.party__actions {
  display: grid;
  gap: 6px;
  width: 100%;
  padding-top: 10px;
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
