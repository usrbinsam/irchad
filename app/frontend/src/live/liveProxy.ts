import { Events } from "@wailsio/runtime";
import { useLiveStore } from "@/stores/liveStore";
import {
  Connect,
  Disconnect,
  PublishMicrophone,
  PublishWebcam,
  UnpublishWebcam,
  GetWindows,
  PublishScreenShare,
  UnpublishScreenShare,
} from "@/bindings/IrChad/internal/live/livechat";

const playConnectionChime = () => {
  // Initialize the browser's native audio engine
  const ctx = new (window.AudioContext || window.webkitAudioContext)();
  const osc = ctx.createOscillator();
  const gain = ctx.createGain();

  osc.connect(gain);
  gain.connect(ctx.destination);

  // Configure a pleasant, soft double-chime (C5 jumping to E5)
  osc.type = "sine";
  osc.frequency.setValueAtTime(523.25, ctx.currentTime); // Note C5
  osc.frequency.setValueAtTime(659.25, ctx.currentTime + 0.15); // Note E5

  // Volume envelope: quick fade-in, smooth fade-out (prevents clicking)
  gain.gain.setValueAtTime(0, ctx.currentTime);
  gain.gain.linearRampToValueAtTime(0.2, ctx.currentTime + 0.05);
  gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.5);

  // Play and self-destruct
  osc.start(ctx.currentTime);
  osc.stop(ctx.currentTime + 0.5);
};

export function setupEvents() {
  Events.On("live:participant-connected", (event) => {
    console.log("participant connected: ", event.data);
    const store = useLiveStore();
    store.addParticipant(event.data.Identity);
  });

  Events.On("live:participant-track-published", (event) => {
    const store = useLiveStore();
    const { data } = event;
    store.addTrack(data.Identity, {
      id: data.TrackID,
      kind: data.Kind,
      source: data.Source,
      subscribeURL: data.SubscribeURL,
      show: data.Show,
    });
    console.log("new track " + data.Kind + " : ", data);
  });

  Events.On("live:participant-dissconnected", (event) => {
    const store = useLiveStore();
    const { data } = event;
    store.dropParticipant(data.Identity);
  });

  Events.On("live:screen-share-closed", () => {
    useLiveStore().screenShareEnabled = false;
  });

  console.log("live events registered");
}

export async function connect(channel: string) {
  await Connect(channel);
  playConnectionChime();
  useLiveStore().setConnected(channel);
}

export async function disconnect() {
  const store = useLiveStore();
  await Disconnect();

  store.reset();
}

export async function publishMic() {
  const store = useLiveStore();
  const rv = await PublishMicrophone();
  console.log(rv);
  store.micEnabled = true;
}

export async function publishCamera() {
  console.log("publishCamera");
  const store = useLiveStore();
  const rv = await PublishWebcam();
  console.log(rv);
  store.camEnabled = true;
}

export async function unpublishCamera() {
  const store = useLiveStore();
  const rv = await UnpublishWebcam();
  console.log("cam disabled: ", rv);
  store.camEnabled = false;
}

export async function getWindows() {
  return await GetWindows();
}

export async function shareWindow(id: number) {
  await PublishScreenShare(id);
  useLiveStore().screenShareEnabled = true;
}

export async function unpublishScreenShare() {
  await UnpublishScreenShare();
  useLiveStore().screenShareEnabled = false;
}
