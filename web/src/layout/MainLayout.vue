<template>
  <div class="h-full flex flex-col w-full dark:bg-black">
    <header class="h-14 flex-shrink-0 border dark:border-gray-700 bg-[var(--el-fill-color-light)]">
      <div class="h-full flex justify-between items-center container mx-auto">
        <div class="flex space-x-4">
          <div class="header-brand">GTUN</div>
          <div class="header-menus flex space-x-4">
            <div><router-link to="/">加速管理</router-link></div>
            <div><router-link to="/config">系统配置</router-link></div>
          </div>
        </div>

        <div class="flex space-x-4">
          <div>
            <el-button :circle="true" :icon="isDark ? Moon : Sunny" @click="toggleDark()"></el-button>
          </div>
          <div title="Github">
            <a href="https://GitHub.com/ICKelin/gtun" target="_blank">
              <el-icon size="28">
                <svg preserveAspectRatio="xMidYMid meet" viewBox="0 0 24 24" width="1.2em" height="1.2em" data-v-6c8d2bba="">
                  <path
                    fill="currentColor"
                    d="M12 2C6.475 2 2 6.475 2 12a9.994 9.994 0 0 0 6.838 9.488c.5.087.687-.213.687-.476c0-.237-.013-1.024-.013-1.862c-2.512.463-3.162-.612-3.362-1.175c-.113-.288-.6-1.175-1.025-1.413c-.35-.187-.85-.65-.013-.662c.788-.013 1.35.725 1.538 1.025c.9 1.512 2.338 1.087 2.912.825c.088-.65.35-1.087.638-1.337c-2.225-.25-4.55-1.113-4.55-4.938c0-1.088.387-1.987 1.025-2.688c-.1-.25-.45-1.275.1-2.65c0 0 .837-.262 2.75 1.026a9.28 9.28 0 0 1 2.5-.338c.85 0 1.7.112 2.5.337c1.912-1.3 2.75-1.024 2.75-1.024c.55 1.375.2 2.4.1 2.65c.637.7 1.025 1.587 1.025 2.687c0 3.838-2.337 4.688-4.562 4.938c.362.312.675.912.675 1.85c0 1.337-.013 2.412-.013 2.75c0 .262.188.574.688.474A10.016 10.016 0 0 0 22 12c0-5.525-4.475-10-10-10z"
                  ></path>
                </svg>
              </el-icon>
            </a>
          </div>
        </div>
      </div>
    </header>
    <main class="flex-grow container mx-auto py-4">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ElButton, ElIcon } from "element-plus";
import { useDark, useToggle } from "@vueuse/core";
import { Moon, Sunny } from "@element-plus/icons-vue";

const isDark = useDark();
const toggleDark = useToggle(isDark);
</script>
