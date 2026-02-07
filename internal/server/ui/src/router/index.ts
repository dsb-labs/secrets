import { createRouter, createWebHistory } from "vue-router";
import { AccountClient } from "@/lib/account";

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      name: "login",
      path: "/auth/login",
      component: () => import("../views/Login.vue"),
      meta: {
        hideSidebar: true,
      },
    },
    {
      name: "login-list",
      path: "/secrets/logins",
      component: () => import("../views/logins/List.vue"),
      children: [
        {
          name: "login-empty",
          path: "/secrets/logins",
          component: () => import("../views/logins/Empty.vue"),
        },
        {
          name: "login-view",
          props: true,
          path: "/secrets/logins/:id/view",
          component: () => import("../views/logins/View.vue"),
        },
      ],
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

  return true;
});

export default router;
