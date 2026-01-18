<script setup>
import { ref } from "vue";
import { useIRCStore } from "@/stores/irc";
import { useBufferStore } from "@/stores/bufferStore";
const emit = defineEmits(["send"]);

const store = useIRCStore();
const bufferStore = useBufferStore();
const text = ref();
const menu = ref({
  open: false,
  activator: null,
});

const menuList = ref({
  density: "compact",
  slim: true,
  items: [],
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
  emit("send", text.value);
  menu.value.open = false;
  text.value = "";
}

function filterUsers(s) {
  if (store.activeBuffer) {
    return store.activeBuffer.users.filter((u) => u.nick.startsWith(s));
  }
}

function trigger(ev) {
  const input = ev.target;
  const cursor = input.selectionStart;
  cursorPos.value = cursor;
  const text = input.value;

  const textBefore = text.slice(0, cursor);
  const mentionMatch = textBefore.match(/@(\w*)$/);
  if (mentionMatch) {
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
  <v-text-field
    v-model="text"
    autofocus
    hide-details
    :placeholder="`Message ${bufferStore.activeBufferName}`"
    @input="trigger"
    @keydown.enter.prevent="send"
    @keydown.tab.prevent="tabComplete"
    variant="outlined"
    class="ma-1"
    role="irchad"
  ></v-text-field>
</template>
