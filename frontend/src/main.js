import { createApp } from 'vue';
import App from './App.vue';
import './style.css';
import * as runtime from '../wailsjs/runtime';

const app = createApp(App);
app.mount('#app');

window.addEventListener('DOMContentLoaded', () => {
  runtime.WindowSetAlwaysOnTop(true);
  runtime.WindowSetDarkTheme();
  runtime.WindowSetSize(220, 100);
  runtime.WindowSetMinSize(200, 80);
  runtime.WindowSetPosition(window.innerWidth - 260, 40);
});
