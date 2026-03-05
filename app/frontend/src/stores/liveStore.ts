import { defineStore } from "pinia";

export interface Participant {
  identity: string;
  tracks: Map<string, Track>;
  speaking: boolean;
}

export interface Track {
  id: string;
  kind: string;
  source: string;
  subscribeURL: string;
  show: boolean;
  trackName: string;
}

export interface Channel {
  participants: Participant[];
}

export const useLiveStore = defineStore("liveStore", () => {
  const channels = ref(new Map<string, Channel>()); // participants in channels which we are not in
  const participants = ref(new Map<string, Participant>()); // particippants in active channels
  const connected = ref(null as string | null);
  const micEnabled = ref(false);
  const camEnabled = ref(false);
  const screenShareEnabled = ref(false);
  const screenShareDialog = ref(false);
  const screenSharePreview = ref(null as string | null);
  const videoDialog = ref(false);

  function addParticipant(identity: string) {
    participants.value.set(identity, {
      identity,
      speaking: false,
      tracks: new Map<string, Track>(),
    });
  }

  function addTrack(participantID: string, track: Track) {
    const p = participants.value.get(participantID);
    if (!p) return;
    p.tracks.set(track.id, track);
  }

  function delTrack(participantID: string, trackID: string) {
    const p = participants.value.get(participantID);
    if (!p) return;
    p.tracks.delete(trackID);
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

  function setSpeaking(participantID: string, speaking: boolean) {
    const p = participants.value.get(participantID);
    if (!p) return;
    p.speaking = speaking;
  }

  function reset() {
    setConnected("");
    participants.value.clear();
    channels.value.clear();
    micEnabled.value = false;
    camEnabled.value = false;
    screenShareEnabled.value = false;
    videoDialog.value = false;
  }

  return {
    connected,
    participants,
    channels,
    camEnabled,
    micEnabled,
    screenShareEnabled,
    screenShareDialog,
    screenSharePreview,
    videoDialog,
    setSpeaking,
    setConnected,
    reset,
    addParticipant,
    addTrack,
    delTrack,
    dropParticipant,
    showTrack,
  };
});
