<script setup>
// Boo, the SideQuest ghost. Pure SVG so it stays crisp at any size and can
// react with tiny expressions.
const props = defineProps({
  size: { type: Number, default: 40 },
  // idle | happy | wink | wow
  mood: { type: String, default: 'idle' },
  // Outfit unlocked by levelling up: a hat id and an optional extra.
  hat: { type: String, default: 'hat-violet' },
  extra: { type: String, default: '' },
});

// Hat colours as gradient stops; the crown has its own shape.
const HATS = {
  'hat-violet': ['#8f74ff', '#5a3fd6'],
  'hat-mint': ['#86f2d1', '#2fb98f'],
  'hat-coral': ['#ffb3a6', '#e8604e'],
  'hat-gold': ['#ffe08a', '#d99a14'],
  crown: ['#ffe08a', '#d99a14'],
};
const hatColors = () => HATS[props.hat] ?? HATS['hat-violet'];
</script>

<template>
  <svg
    class="boo"
    :class="`boo--${mood}`"
    :width="size"
    :height="size"
    viewBox="0 0 64 64"
    aria-hidden="true"
  >
    <defs>
      <linearGradient id="boo-body" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0" stop-color="#ffffff" />
        <stop offset="1" stop-color="#e3d9ff" />
      </linearGradient>
      <linearGradient :id="`boo-hat-${hat}`" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0" :stop-color="hatColors()[0]" />
        <stop offset="1" :stop-color="hatColors()[1]" />
      </linearGradient>
    </defs>

    <g class="boo__float">
      <path
        class="boo__body"
        d="M13 36c0-13 8.5-21 19-21s19 8 19 21v15c0 2-2 3-3.6 1.6-1.6-1.4-3.2-1.4-4.8 0-1.6 1.4-3.2 1.4-4.8 0-1.6-1.4-3.2-1.4-4.8 0-1.6 1.4-3.2 1.4-4.8 0-1.6-1.4-3.2-1.4-4.8 0C15 54 13 53 13 51z"
        fill="url(#boo-body)"
      />
      <template v-if="hat === 'crown'">
        <path
          class="boo__hat"
          d="M18 21.5 19.5 9l6.5 6.5L32 5l6 10.5L44.5 9 46 21.5c-9-2.4-19-2.4-28 0z"
          :fill="`url(#boo-hat-${hat})`"
        />
        <circle cx="32" cy="17" r="2" fill="#ff8fb8" />
      </template>
      <template v-else>
        <path
          class="boo__hat"
          d="M10.5 25.5C14 15 23 8 34 8c7 0 12 3 15.5 7.5 2.5 3.2 6 4.4 9.5 4-2.6 3.6-7 4.6-10.8 3.2C42 20.5 35 19.6 28 20.6c-6.2.9-12.2 2.8-17.5 4.9z"
          :fill="`url(#boo-hat-${hat})`"
        />
        <circle cx="59.5" cy="19.5" r="2.4" fill="#ffcb57" />
      </template>
      <g v-if="extra === 'flower'" class="boo__flower">
        <circle v-for="a in [0, 72, 144, 216, 288]" :key="a" :cx="19 + 3 * Math.cos((a * Math.PI) / 180)" :cy="16 + 3 * Math.sin((a * Math.PI) / 180)" r="2.3" fill="#ff8fb8" />
        <circle cx="19" cy="16" r="1.8" fill="#ffcb57" />
      </g>
      <g v-if="extra === 'scarf'">
        <path d="M13.5 44.5Q32 50.5 50.5 44.5V49Q32 55 13.5 49z" fill="#ff8a7a" />
        <path d="M38.5 50.2l2.6 8.3 5-1.4-2.3-8.1z" fill="#e8604e" />
      </g>

      <g class="boo__face">
        <template v-if="mood === 'happy'">
          <path d="M23.5 35.5q3-3.6 6 0" class="boo__line" />
          <path d="M35 35.5q3-3.6 6 0" class="boo__line" />
        </template>
        <template v-else-if="mood === 'wink'">
          <ellipse class="boo__eye" cx="26.5" cy="34.5" rx="2.6" ry="3.4" />
          <path d="M35 35q3-2.8 6 0" class="boo__line" />
        </template>
        <template v-else>
          <g class="boo__eyes">
            <ellipse class="boo__eye" cx="26.5" cy="34.5" rx="2.6" :ry="mood === 'wow' ? 4 : 3.4" />
            <ellipse class="boo__eye" cx="38" cy="34.5" rx="2.6" :ry="mood === 'wow' ? 4 : 3.4" />
          </g>
        </template>
        <ellipse cx="21.5" cy="41" rx="3.2" ry="2" fill="#ff9db0" opacity=".7" />
        <ellipse cx="43" cy="41" rx="3.2" ry="2" fill="#ff9db0" opacity=".7" />
        <ellipse v-if="mood === 'wow'" cx="32.2" cy="42.5" rx="2.2" ry="2.6" fill="#3b2466" />
        <path v-else d="M29 41.2q3.2 3.6 6.4 0" class="boo__line boo__mouth" />
      </g>
      <g v-if="extra === 'glasses'" class="boo__glasses">
        <circle cx="26.5" cy="34.5" r="5" />
        <circle cx="38" cy="34.5" r="5" />
        <path d="M31.5 34q.75-1 1.5 0" />
      </g>
    </g>
  </svg>
</template>

<style scoped>
.boo {
  display: block;
  overflow: visible;
}
.boo__float {
  transform-origin: 32px 40px;
  animation: boo-float 3.6s ease-in-out infinite;
}
.boo__body {
  filter: drop-shadow(0 2px 3px rgb(40 20 90 / 0.35));
}
.boo__glasses {
  fill: rgb(255 255 255 / 0.18);
  stroke: #2a1a4d;
  stroke-width: 1.5;
}
.boo__eye {
  fill: #2a1a4d;
}
.boo__line {
  fill: none;
  stroke: #2a1a4d;
  stroke-width: 2.2;
  stroke-linecap: round;
}
.boo__mouth {
  stroke: #e0587a;
}
.boo__eyes {
  transform-origin: 32px 34.5px;
  animation: boo-blink 5.2s infinite;
}
.boo--happy .boo__float {
  animation: boo-hop 0.6s ease-out 2;
}

@keyframes boo-float {
  0%, 100% { transform: translateY(0) rotate(0); }
  50% { transform: translateY(-1.5px) rotate(-2deg); }
}
@keyframes boo-blink {
  0%, 92%, 100% { transform: scaleY(1); }
  95% { transform: scaleY(0.1); }
}
@keyframes boo-hop {
  0%, 100% { transform: translateY(0) scale(1); }
  40% { transform: translateY(-5px) scale(1.04, 0.97); }
}
@media (prefers-reduced-motion: reduce) {
  .boo__float, .boo__eyes { animation: none; }
}
</style>
