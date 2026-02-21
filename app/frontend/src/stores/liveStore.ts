import { defineStore } from "pinia";

export interface Participant {
  identity: string;
  tracks: Map<string, Track>;
}

export interface Track {
  id: string;
  kind: string;
  source: string;
  subscribeURL: string;
  show: boolean;
}

export interface Channel {
  participants: Participant[];
}

export const useLiveStore = defineStore("liveStore", () => {
  const channels = ref(new Map<string, Channel>()); // participants in channels which we are not in
  const participants = ref(new Map<string, Participant>()); // particippants in active channels
  const connected = ref("");
  const micEnabled = ref(false);
  const camEnabled = ref(false);
  const screenShareDialog = ref(false);

  function addParticipant(identity: string) {
    participants.value.set(identity, {
      identity,
      tracks: new Map<string, Track>(),
    });
  }

  function addTrack(participantID: string, track: Track) {
    const p = participants.value.get(participantID);
    if (!p) return;
    p.tracks.set(track.id, track);
  }

  function dropParticipant(id: string) {
    participants.value.delete(id);
  }

  function setConnected(v: string) {
    connected.value = v;
  }

  function showTrack(participantID: string, trackID: string) {
    const p = participants.value.get(participantID);
    if (!p) return;
    const t = p.tracks.get(trackID);
    if (!t) return;
    t.show = true;
  }

  function reset() {
    setConnected("");
    participants.value.clear();
    channels.value.clear();
    micEnabled.value = false;
    camEnabled.value = false;
  }

  return {
    connected,
    participants,
    channels,
    camEnabled,
    micEnabled,
    screenShareDialog,
    setConnected,
    reset,
    addParticipant,
    addTrack,
    dropParticipant,
    showTrack,
  };
});
