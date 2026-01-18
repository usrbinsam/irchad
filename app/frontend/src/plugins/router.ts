import { createMemoryHistory, createRouter } from "vue-router";
import Chat from "@/components/Chat.vue";
import { useIRCStore } from "@/stores/irc";
const routes = [
  {
    path: "/",
    name: "Chat",
    component: Chat,
    children: [
      {
        path: "register",
        name: "Register",
        component: () => import("@/components/Register.vue"),
      },
    ],
  },
  {
    path: "/login",
    name: "Login",
    component: () => import("@/components/Login.vue"),
  },
];

const router = createRouter({
  history: createMemoryHistory(),
  routes,
});

router.beforeEach(async (to, from) => {
  if (!useIRCStore().connected && to.name !== "Login") {
    return { name: "Login" };
  }

  return true;
});

export default router;
