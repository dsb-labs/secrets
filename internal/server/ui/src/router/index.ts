import { createRouter, createWebHistory } from "vue-router";
import { AccountClient } from "@/lib/account";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      name: "login",
      path: "/auth/login",
      component: () => import("../views/Login.vue"),
    },
  ],
});

const accounts = new AccountClient();

router.beforeEach(async (to, from) => {
  if (to.name === "login") {
    return true;
  }

  try {
    await accounts.find();
  } catch (e) {
    return { name: "login" };
  }

  return to;
});

export default router;
