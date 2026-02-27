<script lang="ts" setup>
import { useAccountStore } from "@/stores/accountStore";
import { useLiveStore, type Participant } from "@/stores/liveStore";
import { useIRCStore } from "@/stores/irc";
const liveStore = useLiveStore();
const accountStore = useAccountStore();
const ircStore = useIRCStore();

function hasVideo(p: Participant) {
  for (const t of p.tracks.values()) {
    if (t.kind.toLowerCase() === "video") return true;
  }
  return false;
}

interface ParticipantItem {
  identity: string;
  avatar?: string;
  speaking: boolean;
  webcam: boolean;
  streaming: boolean;
  muted: boolean;
}

const sortedList = computed(() => {
  const out: ParticipantItem[] = [];
  for (const [ident, p] of liveStore.participants.entries()) {
    out.push({
      identity: ident,
      speaking: false,
      avatar: ircStore.getMetadata(ident, "avatar"),
      webcam: hasVideo(p),
      streaming: hasVideo(p),
      muted: false,
    });
  }

  out.push({
    identity: accountStore.account.nick,
    speaking: false,
    webcam: liveStore.camEnabled,
    streaming: liveStore.screenShareEnabled,
    muted: !liveStore.micEnabled,
  });

  out.sort((a, b) => a.identity.localeCompare(b.identity));
  return out;
});
</script>

<template>
  <v-sheet color="blue-grey-darken-3" class="py-5 px-2 d-flex flex-column">
    <ParticipantListItem
      v-for="p in sortedList"
      :key="p.identity"
      v-bind="p"
      class="mb-1"
    />
  </v-sheet>
</template>
