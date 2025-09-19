<template>
  <div v-loading="loading" class="h-full flex flex-col space-y-4">
    <div>
      <el-button :icon="Plus" @click="showAddDialog()">添加</el-button>
    </div>
    <div class="space-y-2">
      <div class="flex justify-between">
        <div class="flex items-center space-x-4">
          <span>区域</span>
          <el-checkbox-group v-model="regionFilter">
            <el-checkbox v-for="region in data?.regions" :label="region" />
          </el-checkbox-group>
          <div class="flex space-x-0">
            <el-button link @click="regionFilter = data?.regions || []">全选</el-button>
            <el-button link @click="regionFilter = []">取消</el-button>
          </div>
        </div>

        <div>
          <el-input v-model="ipFilter" placeholder="按IP搜索" :clearable="true" :suffix-icon="Search" />
        </div>
      </div>
      <el-table :border="true" :data="tableData">
        <el-table-column type="index" width="50" />
        <el-table-column label="IP" prop="ip" :sortable="true" />
        <el-table-column label="区域" prop="region" :sortable="true" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="danger" size="small" @click="handleDelete(row.ip, row.region)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>

  <!-- 添加IP -->
  <el-dialog :model-value="showDialog" :close-on-click-modal="false" title="添加IP" width="600px">
    <el-form ref="formRef" :model="form" :rules="rules" status-icon label-width="auto">
      <el-form-item label="区域" prop="region">
        <el-select v-model:model-value="form.region">
          <el-option v-for="item in data?.regions" :key="item" :label="item" :value="item" />
        </el-select>
      </el-form-item>
      <el-form-item label="IP" prop="ip">
        <el-input v-model="form.ip" placeholder="请输入IP" />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="showDialog = false">取消</el-button>
        <el-button :loading="saving" type="primary" @click="handleAdd(formRef)">提交</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ElButton, ElCheckbox, ElCheckboxGroup, ElMessageBox, ElTable, ElTableColumn, ElMessage, ElDialog, FormInstance, FormRules, ElForm, ElFormItem, ElSelect, ElOption, ElInput, vLoading } from "element-plus";
import { Plus, Search } from "@element-plus/icons-vue";
import { onMounted, ref, reactive, computed } from "vue";
import { MetaData, RegionData } from "@/types";
import { getMeta, deleteIp, addIp, listIp } from "@/apis";

interface RegionIp {
  ip: string;
  region: string;
}

const ipFilter = ref<string>();
const regionFilter = ref<string[]>([]);
const data = ref<MetaData>();
const regionData = ref<RegionData>();
const tableData = computed(() => {
  if (regionData.value) {
    let result = [] as RegionIp[];
    Object.keys(regionData.value).forEach((region) => {
      // 按区域筛选
      if (regionFilter.value.some((r) => r === region)) {
        const ips = regionData.value![region as keyof RegionData];
        ips.forEach((ip) => {
          // 按IP搜索
          if (ipFilter.value && ipFilter.value.length > 0) {
            if (ip.toLocaleLowerCase().indexOf(ipFilter.value) > -1) {
              result.push({ ip: ip, region: region });
            }
          } else {
            result.push({ ip: ip, region: region });
          }
        });
      }
    });
    return result;
  }

  return [] as RegionIp[];
});

const saving = ref(false);
const showDialog = ref(false);
const formRef = ref<FormInstance>();
const form = reactive({
  ip: "",
  region: "",
});
const rules = reactive<FormRules>({
  ip: { required: true, message: "请输入IP" },
  region: { required: true, message: "请选择区域" },
});

const showAddDialog = () => {
  form.ip = "";
  form.region = "";
  showDialog.value = true;
  formRef.value?.resetFields();
};

const handleAdd = (formEl: FormInstance | undefined) => {
  if (!formEl) return;

  formEl.validate((valid) => {
    if (valid) {
      saving.value = true;
      addIp(form.ip, form.region)
        .then(() => {
          ElMessage.success("添加成功");
          showDialog.value = false;
          loadRegion();
        })
        .catch((err) => {
          ElMessage.error("添加失败:" + err);
        })
        .finally(() => {
          saving.value = false;
        });
    }
  });
};

const handleDelete = (ip: string, region: string) => {
  if (!ip) {
    return;
  }

  ElMessageBox.confirm("确定删除" + ip + "?", "警告", {
    confirmButtonText: "确定",
    cancelButtonText: "取消",
    type: "warning",
  })
    .then(() => {
      deleteIp(ip, region)
        .then(() => {
          ElMessage({
            type: "success",
            message: "删除成功",
          });
        })
        .catch((err) => {
          ElMessage({
            type: "error",
            message: "删除失败:" + err,
          });
        })
        .finally(() => {
          loadRegion();
        });
    })
    .catch(() => {});
};

const loading = ref(false);
const loadRegion = async () => {
  regionData.value = await listIp();
};
onMounted(async () => {
  try {
    loading.value = true;
    data.value = await getMeta();
    regionData.value = await listIp();

    regionFilter.value = data.value.regions;
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped></style>
