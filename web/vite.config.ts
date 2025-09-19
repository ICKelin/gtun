import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import { resolve } from "path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  // 配置别名
  resolve: {
    alias: [
      {
        find: "@",
        replacement: resolve(__dirname, "./src"),
      },
    ],
  },
  // 配置本地代理
  server: {
    proxy: {
      "/meta": {
        target: "http://wjw7d13q1zdi.glana.link",
        changeOrigin: true,
      },
      "/ip": {
        target: "http://wjw7d13q1zdi.glana.link",
        changeOrigin: true,
      },
    },
  },
});
