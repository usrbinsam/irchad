<script setup lang="ts">
import { ref } from "vue";
import { useIRCStore } from "@/stores/irc";
import { useBufferStore } from "@/stores/bufferStore";
import { useAccountStore } from "@/stores/accountStore";
import { useRouter } from "vue-router";

const emit = defineEmits(["send", "raw"]);

const store = useIRCStore();
const accountStore = useAccountStore();
const { showRegistration } = storeToRefs(accountStore);
const router = useRouter();
const bufferStore = useBufferStore();
const text = ref();
const menu = ref({
  open: false,
  activator: null,
});

const menuList = ref({
  density: "compact",
  slim: true,
  items: [] as any[],
  itemTitle: "title",
  itemValue: "value",
  selected: [],
  selectable: true,
  mandatory: true,
  returnObject: true,
});

const cursorPos = ref(0);
const completionPos = ref(0);

function clickItem() {}
function send() {
  if (!text.value) return;

  if (text.value.slice(0, 2) === "//") {
    emit("raw", text.value.substring(2));
  } else if (text.value[0] === "/") {
    router.push({ path: "/register" });
  } else {
    emit("send", text.value);
  }

  menu.value.open = false;
  text.value = "";
}

function filterUsers(s: string) {
  if (store.activeBuffer) {
    return store.activeBuffer.users.filter((u) => u.nick.startsWith(s));
  }
}

function trigger(ev: KeyboardEvent) {
  const input = ev.target;
  const cursor = input.selectionStart;
  cursorPos.value = cursor;
  const text = input.value;

  const textBefore = text.slice(0, cursor);
  const mentionMatch = textBefore.match(/@(\w*)$/);
  if (text[0] === "/") {
    menu.value.open = true;
    menuList.value.items = [
      {
        title: "register",
        value: "register",
        cmd: () => {
          showRegistration.value = true;
        },
      },
    ];
  } else if (mentionMatch) {
    menu.value.open = true;
    menuList.value.items = filterUsers(mentionMatch[1]);
    menuList.value.itemTitle = "nick";
    menuList.value.itemValue = "nick";
  } else {
    menu.value.open = false;
  }
}

function tabComplete() {
  if (!menu.value.open) {
    return;
  }

  let nextSelection = 0;
  if (menuList.value.selected.length) {
    const currentIdx = menuList.value.items.indexOf(menuList.value.selected[0]);
    const nextIdx = currentIdx + 1;
    if (menuList.value.items[nextIdx]) nextSelection = nextIdx;
  }

  menuList.value.selected = [menuList.value.items[nextSelection]];

  // hello @

  const beforeCursor = text.value.splice(0, cursor);
  const afterCursor = text.value.slice(cursor);
}
const rows = ref(1);
</script>
<template>
  <v-menu
    v-model="menu.open"
    location="top"
    activator="parent"
    :open-on-click="false"
    :open-on-focus="false"
  >
    <v-list v-bind="menuList" @click:select="clickItem" />
  </v-menu>
  <v-textarea
    :rows="rows"
    auto-grow
    v-model="text"
    autofocus
    hide-details
    :placeholder="`Message ${bufferStore.activeBufferName}`"
    @input="trigger"
    @keydown.enter.exact.prevent="send"
    @keydown.tab.prevent="tabComplete"
    variant="outlined"
    class="ma-1"
    role="irchad"
  />
</template>
