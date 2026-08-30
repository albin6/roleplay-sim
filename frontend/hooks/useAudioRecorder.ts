import { useEffect, useRef, useState } from "react";

interface UseAudioRecorderOptions {
  localStream: MediaStream | null;
  isLive: boolean;
  sendBinary: (data: Blob) => void;
}

export function useAudioRecorder({
  localStream,
  isLive,
  sendBinary,
}: UseAudioRecorderOptions) {
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const [isRecording, setIsRecording] = useState(false);
  const [chunksSent, setChunksSent] = useState(0);

  useEffect(() => {
    if (!isLive || !localStream) {
      if (mediaRecorderRef.current && mediaRecorderRef.current.state !== "inactive") {
        mediaRecorderRef.current.stop();
        setIsRecording(false);
      }
      return;
    }

    const audioTracks = localStream.getAudioTracks();
    if (audioTracks.length === 0) {
      console.warn("useAudioRecorder: No audio tracks found on localStream");
      return;
    }

    // Isolate audio track to prevent recording unnecessary video stream
    const audioStream = new MediaStream(audioTracks);

    let mimeType = "audio/webm;codecs=opus";
    if (!MediaRecorder.isTypeSupported(mimeType)) {
      if (MediaRecorder.isTypeSupported("audio/webm")) {
        mimeType = "audio/webm";
      } else if (MediaRecorder.isTypeSupported("audio/ogg;codecs=opus")) {
        mimeType = "audio/ogg;codecs=opus";
      } else {
        mimeType = "";
      }
    }

    try {
      const options: MediaRecorderOptions = mimeType ? { mimeType } : {};
      const recorder = new MediaRecorder(audioStream, options);
      mediaRecorderRef.current = recorder;

      recorder.ondataavailable = (event) => {
        if (event.data && event.data.size > 0) {
          sendBinary(event.data);
          setChunksSent((prev) => prev + 1);
        }
      };

      recorder.onstart = () => {
        setIsRecording(true);
        console.log("Audio recording started for AI evaluation");
      };

      recorder.onstop = () => {
        setIsRecording(false);
        console.log("Audio recording stopped");
      };

      // Request slices every 5 seconds (5000ms)
      recorder.start(5000);
    } catch (err) {
      console.error("Failed to initialize MediaRecorder:", err);
    }

    return () => {
      if (mediaRecorderRef.current && mediaRecorderRef.current.state !== "inactive") {
        mediaRecorderRef.current.stop();
      }
    };
  }, [isLive, localStream, sendBinary]);

  return { isRecording, chunksSent };
}