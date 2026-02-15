<script lang="ts" setup>
import { ref, useTemplateRef } from "vue";
import { RemoteTrack, Room, RoomEvent, Track } from "livekit-client";
const room = ref(null as Room | null);
const rootElement = useTemplateRef<HTMLDivElement>("root");

defineExpose({ connect });

function getToken() {
  return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NzExOTg2MjAsImlkZW50aXR5IjoiY2hhZCIsImlzcyI6ImRldmtleSIsIm5hbWUiOiJjaGFkIiwibmJmIjoxNzcxMTEyMjIwLCJzdWIiOiJjaGFkIiwidmlkZW8iOnsicm9vbSI6IiNjaGFkIiwicm9vbUpvaW4iOnRydWV9fQ.kKrUUVUVVe3OCHw_CGL24LHEzbQyNXR42R_JvDyWXzw";
}

function getServerUrl() {
  return "ws://10.10.10.106:7880";
}

async function connect() {
  const r = new Room({
    adaptiveStream: true,
    dynacast: true,
  });

  const token = getToken();
  const url = getServerUrl();
  r.prepareConnection(url, token);
  await navigator.mediaDevices.getUserMedia({
    audio: true,
    video: true,
  });

  r.on(RoomEvent.TrackSubscribed, handleTrackSubscribed).on(
    RoomEvent.TrackUnsubscribed,
    handleTrackUnsubscribed,
  );

  console.log("connecting to livekit", url);
  await r.connect(url, token, {
    autoSubscribe: true,
    rtcConfig: {
      iceTransportPolicy: "all",
      iceServers: [],
      rtcpMuxPolicy: "require",
      bundlePolicy: "max-bundle",
    },
  });
  console.log("connected");
  await r.localParticipant.enableCameraAndMicrophone();
  room.value = r;
}

function handleTrackSubscribed(track: RemoteTrack) {
  if (track.kind === Track.Kind.Video || track.kind == Track.Kind.Audio) {
    console.log("track subscribed, adding element");
    const el = track.attach();
    rootElement.value!.appendChild(el);
  }
}

function handleTrackUnsubscribed(track: RemoteTrack) {
  track.detach();
}
</script>
<template>
  <div ref="root"></div>
</template>
