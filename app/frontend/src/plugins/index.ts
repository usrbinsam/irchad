import vuetify from "./vuetify";
import pinia from "@/stores";
import router from "./router.ts";

import type { App } from "vue";

export function registerPlugins(app: App) {
  app.use(vuetify).use(router).use(pinia);
}
