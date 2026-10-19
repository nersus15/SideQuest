<script setup>
// Boo, the SideQuest ghost. Pure SVG so it stays crisp at any size and can
// react with tiny expressions.
defineProps({
  size: { type: Number, default: 40 },
  // idle | happy | wink | wow
  mood: { type: String, default: 'idle' },
});
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
      <linearGradient id="boo-hat" x1="0" y1="0" x2="1" y2="1">
        <stop offset="0" stop-color="#8f74ff" />
        <stop offset="1" stop-color="#5a3fd6" />
      </linearGradient>
    </defs>

    <g class="boo__float">
      <path
        class="boo__body"
        d="M13 36c0-13 8.5-21 19-21s19 8 19 21v15c0 2-2 3-3.6 1.6-1.6-1.4-3.2-1.4-4.8 0-1.6 1.4-3.2 1.4-4.8 0-1.6-1.4-3.2-1.4-4.8 0-1.6 1.4-3.2 1.4-4.8 0-1.6-1.4-3.2-1.4-4.8 0C15 54 13 53 13 51z"
        fill="url(#boo-body)"
      />
      <path
        class="boo__hat"
        d="M10.5 25.5C14 15 23 8 34 8c7 0 12 3 15.5 7.5 2.5 3.2 6 4.4 9.5 4-2.6 3.6-7 4.6-10.8 3.2C42 20.5 35 19.6 28 20.6c-6.2.9-12.2 2.8-17.5 4.9z"
        fill="url(#boo-hat)"
      />
      <circle cx="59.5" cy="19.5" r="2.4" fill="#ffcb57" />

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
