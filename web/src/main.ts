import { createApp } from "vue";
import "./styles/index.css";
import "element-plus/dist/index.css";
import "element-plus/theme-chalk/dark/css-vars.css";
import "./styles/dark/css-vars.css";
import App from "./App.vue";
import router from "./router";

createApp(App).use(router).mount("#app");
