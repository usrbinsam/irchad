<script lang="ts" setup>
import { ref, useTemplateRef } from "vue";
import {
  LocalParticipant,
  LocalTrackPublication,
  Participant,
  RemoteTrack,
  RemoteTrackPublication,
  Room,
  RoomEvent,
  Track,
} from "livekit-client";
const room = ref(null as Room | null);
const rootElement = useTemplateRef<HTMLDivElement>("root");

defineExpose({ connect });

function getToken() {
  return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzExOTg2MjAsImlkZW50aXR5IjoiY2hhZCIsImlzcyI6ImRldmtleSIsIm5hbWUiOiJjaGFkIiwibmJmIjoxNzcxMTEyMjIwLCJzdWIiOiJjaGFkIiwidmlkZW8iOnsicm9vbSI6IiNjaGFkIiwicm9vbUpvaW4iOnRydWV9fQ.kKrUUVUVVe3OCHw_CGL24LHEzbQyNXR42R_JvDyWXzw";
}

function getServerUrl() {
  return "ws://localhost:7880";
}

async function connect() {
  const r = new Room({
    adaptiveStream: true,
    dynacast: true,
  });

  const token = getToken();
  const url = getServerUrl();
  r.prepareConnection(url, token);

  r.on(RoomEvent.TrackSubscribed, handleTrackSubscribed)
    .on(RoomEvent.TrackUnsubscribed, handleTrackUnsubscribed)
    .on(RoomEvent.ActiveSpeakersChanged, handleActiveSpeakerChange)
    .on(RoomEvent.LocalTrackUnpublished, handleLocalTrackUnpublished)
    .on(RoomEvent.Disconnected, handleDisconnect)
    .on(RoomEvent.TrackPublished, handleTrackPublished);

  console.log("connecting to livekit", url);
  await r.connect(url, token, {
    autoSubscribe: false,
    rtcConfig: {
      // iceTransportPolicy: "all",
      // iceServers: [],
      // rtcpMuxPolicy: "require",
      bundlePolicy: "max-bundle",
    },
  });
  console.log("connected to ", r.name);
  await r.localParticipant.enableCameraAndMicrophone();
  room.value = r;
}

function handleTrackPublished(
  publication: RemoteTrackPublication,
  participant: Participant,
) {
  console.log("new track from - ", participant);
  publication.setSubscribed(true);
}

function handleTrackSubscribed(track: RemoteTrack) {
  console.log("handleTrackSubscribed()");
  if (track.kind === Track.Kind.Video || track.kind == Track.Kind.Audio) {
    console.log("track subscribed, adding element");
    const el = track.attach();
    rootElement.value!.appendChild(el);
  }
}

function handleTrackUnsubscribed(track: RemoteTrack) {
  track.detach();

  if (track.kind === "video") {
    track.mediaStreamTrack.stop();
  }
}

function handleLocalTrackUnpublished(publication: LocalTrackPublication) {
  // when local tracks are ended, update UI to remove them from rendering
  publication.track.detach();
}
function handleActiveSpeakerChange(speakers: Participant[]) {
  console.log(speakers);
}
function handleDisconnect() {
  console.log("disconnected from ", room.value!.name);
}
</script>
<template>
  <div ref="root"></div>
</template>
