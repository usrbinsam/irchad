import { registerPlugins } from "@/plugins";
import { setupEvents } from "@/live/liveProxy";

// Components
import App from "./App.vue";

// Composables
import { createApp } from "vue";

// Styles
// import "unfonts.css";

const app = createApp(App);

registerPlugins(app);

setupEvents();
app.mount("#app");
