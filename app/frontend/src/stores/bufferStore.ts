import { Buffer, type BufferOptions } from "@/lib/buffer.ts";
import { defineStore } from "pinia";
import { computed, ref } from "vue";

export const useBufferStore = defineStore("bufferStore", () => {
  const buffers = ref({} as Record<string, Buffer>);

  const activeBufferName = ref(null as string | null);

  function setActiveBuffer(bufferName: string) {
    const buffer = getBuffer(bufferName);
    if (!buffer) return;
    activeBufferName.value = bufferName;
    buffer.resetLastSeen();
  }

  function addBuffer(bufferName: string, options: BufferOptions) {
    buffers.value[bufferName] = new Buffer(options);
    return buffers.value[bufferName];
  }

  function getBuffer(bufferName: string) {
    return buffers.value[bufferName];
  }

  function delBuffer(bufferName: string) {
    if (buffers.value[bufferName]) {
      delete buffers.value[bufferName];
    }
  }

  const activeBuffer = computed(() => {
    if (activeBufferName.value) {
      return buffers.value[activeBufferName.value];
    }
  });

  return {
    buffers,
    activeBufferName,
    activeBuffer,
    addBuffer,
    getBuffer,
    delBuffer,
    setActiveBuffer,
  };
});
