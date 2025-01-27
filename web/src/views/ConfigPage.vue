<template>
  <div class="config-view">
    <div>配置文件</div>
    <MdPreview v-if="isDark" v-model:model-value="previewData" theme="dark" />
    <MdPreview v-else v-model:model-value="previewData" theme="light" />
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref, computed } from "vue";
import { getMeta } from "@/apis";
import { MetaData } from "@/types";
import { MdPreview } from "md-editor-v3";
import "md-editor-v3/lib/preview.css";
import { useDark } from "@vueuse/core";

const isDark = useDark();
const data = ref<MetaData>();
const previewData = computed(() => {
  if (data && data.value) {
    return "```json\n" + JSON.stringify(data.value.Cfg, null, "  ") + "\n```";
  } else {
    return "";
  }
});

onMounted(async () => {
  data.value = await getMeta();
});
</script>

<style>
.config-view .md-editor-preview-wrapper {
  padding: 0;
}
</style>
