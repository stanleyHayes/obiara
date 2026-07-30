export interface LivenessCapture {
  voiceMediaType: string;
  voiceBase64: string;
  faceMediaType: "image/jpeg";
  faceBase64: string;
}

function blobBase64(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () =>
      reject(new Error("The secure capture could not be read."));
    reader.onload = () => {
      const value = typeof reader.result === "string" ? reader.result : "";
      const comma = value.indexOf(",");
      if (comma < 0)
        reject(new Error("The secure capture could not be encoded."));
      else resolve(value.slice(comma + 1));
    };
    reader.readAsDataURL(blob);
  });
}

function recordAudio(stream: MediaStream): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const audio = new MediaStream(stream.getAudioTracks());
    const preferred = [
      "audio/webm;codecs=opus",
      "audio/webm",
      "audio/mp4",
    ].find((type) => MediaRecorder.isTypeSupported(type));
    const recorder = new MediaRecorder(
      audio,
      preferred
        ? { mimeType: preferred, audioBitsPerSecond: 64_000 }
        : undefined,
    );
    const chunks: Blob[] = [];
    recorder.ondataavailable = (event) => {
      if (event.data.size > 0) chunks.push(event.data);
    };
    recorder.onerror = () =>
      reject(new Error("The microphone capture failed."));
    recorder.onstop = () =>
      resolve(new Blob(chunks, { type: recorder.mimeType || "audio/webm" }));
    recorder.start();
    window.setTimeout(() => recorder.stop(), 2500);
  });
}

async function captureFace(stream: MediaStream): Promise<Blob> {
  const video = document.createElement("video");
  video.muted = true;
  video.playsInline = true;
  video.srcObject = stream;
  await video.play();
  if (video.videoWidth === 0) {
    await new Promise<void>((resolve) => {
      video.onloadedmetadata = () => resolve();
    });
  }
  const maxWidth = 720;
  const scale = Math.min(1, maxWidth / video.videoWidth);
  const canvas = document.createElement("canvas");
  canvas.width = Math.max(1, Math.round(video.videoWidth * scale));
  canvas.height = Math.max(1, Math.round(video.videoHeight * scale));
  canvas.getContext("2d")?.drawImage(video, 0, 0, canvas.width, canvas.height);
  video.pause();
  video.srcObject = null;
  return new Promise((resolve, reject) =>
    canvas.toBlob(
      (blob) =>
        blob ? resolve(blob) : reject(new Error("The face capture failed.")),
      "image/jpeg",
      0.82,
    ),
  );
}

export async function captureLiveness(): Promise<LivenessCapture> {
  if (
    !navigator.mediaDevices?.getUserMedia ||
    typeof MediaRecorder === "undefined"
  ) {
    throw new Error("This browser cannot perform the secure camera check.");
  }
  const stream = await navigator.mediaDevices.getUserMedia({
    audio: { echoCancellation: true, noiseSuppression: true },
    video: {
      facingMode: "user",
      width: { ideal: 720 },
      height: { ideal: 720 },
    },
  });
  try {
    const [voice, face] = await Promise.all([
      recordAudio(stream),
      captureFace(stream),
    ]);
    if (
      voice.size === 0 ||
      voice.size > 2 * 1024 * 1024 ||
      face.size > 1024 * 1024
    ) {
      throw new Error(
        "The capture was empty or exceeded the secure size limit.",
      );
    }
    return {
      voiceMediaType: voice.type || "audio/webm",
      voiceBase64: await blobBase64(voice),
      faceMediaType: "image/jpeg",
      faceBase64: await blobBase64(face),
    };
  } catch (error) {
    if (error instanceof DOMException && error.name === "NotAllowedError") {
      throw new Error(
        "Camera and microphone permission are required for this check.",
      );
    }
    throw error;
  } finally {
    stream.getTracks().forEach((track) => track.stop());
  }
}
