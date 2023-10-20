import { createRouter, createWebHashHistory, RouteRecordRaw } from "vue-router";

// 路由
const routes = [
  {
    path: "/",
    name: "ROOT",
    redirect: "/home",
    component: () => import("@/layout/MainLayout.vue"),
    children: [
      {
        path: "home",
        name: "HomePage",
        component: () => import("@/views/HomePage.vue"),
      },
      {
        path: "config",
        name: "ConfigPage",
        component: () => import("@/views/ConfigPage.vue"),
      },
    ],
  },
] as RouteRecordRaw[];

// 创建路由对象
const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

// 导出路由对象
export default router;
