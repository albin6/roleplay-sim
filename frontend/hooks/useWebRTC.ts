import { useEffect, useRef, useState, useCallback } from "react";
import { useRoomStore } from "../stores/roomStore";

const ICE_SERVERS: RTCConfiguration = {
  iceServers: [
    { urls: "stun:stun.l.google.com:19302" },
    { urls: "stun:stun1.l.google.com:19302" },
  ],
};

interface UseWebRTCOptions {
  roomId: string;
  sendSignal: (signal: any) => void;
  onSignalReceived: (handler: (signal: any) => void) => () => void;
}

export function useWebRTC({ roomId, sendSignal, onSignalReceived }: UseWebRTCOptions) {
  const pcRef = useRef<RTCPeerConnection | null>(null);
  const remoteStreamRef = useRef<MediaStream | null>(null);
  const pendingIceCandidatesRef = useRef<RTCIceCandidateInit[]>([]);
  const hasOfferedRef = useRef(false);

  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStream, setRemoteStream] = useState<MediaStream | null>(null);
  const [connectionState, setConnectionState] = useState<RTCPeerConnectionState>("new");
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoOff, setIsVideoOff] = useState(false);

  const seat = useRoomStore((s) => s.seat);
  const roomState = useRoomStore((s) => s.state);

  // 1. Acquire local microphone and camera (handles single camera hardware lock)
  useEffect(() => {
    let activeStream: MediaStream | null = null;
    let isCancelled = false;

    navigator.mediaDevices
      .getUserMedia({ audio: true, video: true })
      .then((s) => {
        if (isCancelled) {
          s.getTracks().forEach((t) => t.stop());
          return;
        }
        activeStream = s;
        setLocalStream(s);
        setIsVideoOff(false);
      })
      .catch((err) => {
        console.warn("Could not acquire video+audio (camera may be in use by another tab), falling back to audio:", err);
        navigator.mediaDevices
          .getUserMedia({ audio: true, video: false })
          .then((s) => {
            if (isCancelled) {
              s.getTracks().forEach((t) => t.stop());
              return;
            }
            activeStream = s;
            setLocalStream(s);
            setIsVideoOff(true);
          })
          .catch((e) => console.error("Could not capture any local media:", e));
      });

    return () => {
      isCancelled = true;
      if (activeStream) {
        activeStream.getTracks().forEach((track) => track.stop());
      }
    };
  }, []);

  // 2. Initialize PeerConnection with exact 1-audio, 1-video transceivers
  const initPeerConnection = useCallback(() => {
    if (pcRef.current) return pcRef.current;

    const pc = new RTCPeerConnection(ICE_SERVERS);
    pcRef.current = pc;

    const audioTrack = localStream?.getAudioTracks()[0];
    const videoTrack = localStream?.getVideoTracks()[0];

    // Configure exactly one audio transceiver
    if (audioTrack && localStream) {
      pc.addTransceiver(audioTrack, { direction: "sendrecv", streams: [localStream] });
    } else {
      pc.addTransceiver("audio", { direction: "sendrecv" });
    }

    // Configure exactly one video transceiver
    if (videoTrack && localStream) {
      pc.addTransceiver(videoTrack, { direction: "sendrecv", streams: [localStream] });
    } else {
      pc.addTransceiver("video", { direction: "sendrecv" });
    }

    // Listen for remote tracks
    pc.ontrack = (event) => {
      console.log("WebRTC ontrack received:", event.track.kind, event.track.id);
      if (!remoteStreamRef.current && typeof MediaStream !== "undefined") {
        remoteStreamRef.current = new MediaStream();
      }
      if (remoteStreamRef.current) {
        remoteStreamRef.current.addTrack(event.track);
        setRemoteStream(new MediaStream(remoteStreamRef.current.getTracks()));
      }
    };

    pc.onicecandidate = (event) => {
      if (event.candidate) {
        sendSignal({
          type: "ice",
          candidate: event.candidate,
        });
      }
    };

    pc.onconnectionstatechange = () => {
      console.log("WebRTC connection state:", pc.connectionState);
      setConnectionState(pc.connectionState);
    };

    return pc;
  }, [localStream, sendSignal]);

  // 3. Dynamically replace tracks on transceivers when localStream changes
  useEffect(() => {
    if (!pcRef.current || !localStream) return;
    const pc = pcRef.current;

    pc.getTransceivers().forEach((transceiver) => {
      const kind = transceiver.receiver.track.kind;
      const track =
        kind === "video"
          ? localStream.getVideoTracks()[0]
          : localStream.getAudioTracks()[0];

      if (track) {
        transceiver.sender.replaceTrack(track);
        transceiver.direction = "sendrecv";
      }
    });
  }, [localStream]);

  // 4. Handle incoming signaling messages
  useEffect(() => {
    const unsubscribe = onSignalReceived(async (raw) => {
      const pc = initPeerConnection();
      const signal = raw.signal || raw;

      try {
        if (signal.type === "offer") {
          console.log("Processing WebRTC offer");
          await pc.setRemoteDescription(new RTCSessionDescription(signal));

          // Drain any queued ICE candidates
          while (pendingIceCandidatesRef.current.length > 0) {
            const cand = pendingIceCandidatesRef.current.shift();
            if (cand) await pc.addIceCandidate(new RTCIceCandidate(cand));
          }

          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);
          sendSignal({
            type: "answer",
            sdp: answer.sdp,
          });
        } else if (signal.type === "answer") {
          console.log("Processing WebRTC answer");
          await pc.setRemoteDescription(new RTCSessionDescription(signal));

          // Drain queued ICE candidates
          while (pendingIceCandidatesRef.current.length > 0) {
            const cand = pendingIceCandidatesRef.current.shift();
            if (cand) await pc.addIceCandidate(new RTCIceCandidate(cand));
          }
        } else if (signal.type === "ice" && signal.candidate) {
          if (!pc.remoteDescription) {
            console.log("Queueing ICE candidate until remote description is set");
            pendingIceCandidatesRef.current.push(signal.candidate);
          } else {
            await pc.addIceCandidate(new RTCIceCandidate(signal.candidate));
          }
        }
      } catch (err) {
        console.error("WebRTC signal processing error:", err);
      }
    });

    return unsubscribe;
  }, [onSignalReceived, initPeerConnection, sendSignal]);

  // 5. Seat A sends WebRTC Offer upon entering signaling or live phase
  useEffect(() => {
    if ((roomState === "signaling" || roomState === "live") && seat === "A" && !hasOfferedRef.current) {
      hasOfferedRef.current = true;
      const pc = initPeerConnection();
      pc.createOffer()
        .then((offer) => pc.setLocalDescription(offer).then(() => offer))
        .then((offer) => {
          console.log("Seat A dispatched WebRTC offer");
          sendSignal({
            type: "offer",
            sdp: offer.sdp,
          });
        })
        .catch((err) => {
          hasOfferedRef.current = false;
          console.error("Failed to generate offer:", err);
        });
    }
  }, [roomState, seat, initPeerConnection, sendSignal]);

  // 6. Toggle Mute: enables/disables microphone
  const toggleMute = () => {
    if (!localStream) return;
    const audioTrack = localStream.getAudioTracks()[0];
    if (audioTrack) {
      audioTrack.enabled = !audioTrack.enabled;
      setIsMuted(!audioTrack.enabled);
    }
  };

  // 7. Toggle Video: physically stops hardware camera & turns off LED when off
  const toggleVideo = async () => {
    if (!localStream) return;

    if (!isVideoOff) {
      // Turn camera OFF: stop all video tracks to turn off hardware sensor and LED
      localStream.getVideoTracks().forEach((track) => {
        track.stop();
        localStream.removeTrack(track);
      });

      if (pcRef.current) {
        pcRef.current.getTransceivers().forEach((t) => {
          if (t.receiver.track.kind === "video") {
            t.sender.replaceTrack(null);
          }
        });
      }

      setIsVideoOff(true);
      setLocalStream(new MediaStream(localStream.getTracks()));
    } else {
      // Turn camera ON: request camera device from browser
      try {
        const newStream = await navigator.mediaDevices.getUserMedia({ video: true });
        const newTrack = newStream.getVideoTracks()[0];
        if (newTrack) {
          localStream.addTrack(newTrack);

          if (pcRef.current) {
            pcRef.current.getTransceivers().forEach((t) => {
              if (t.receiver.track.kind === "video") {
                t.sender.replaceTrack(newTrack);
              }
            });
          }

          setIsVideoOff(false);
          setLocalStream(new MediaStream(localStream.getTracks()));
        }
      } catch (err) {
        console.warn("Could not re-activate camera:", err);
      }
    }
  };

  return {
    localStream,
    remoteStream,
    connectionState,
    isMuted,
    isVideoOff,
    toggleMute,
    toggleVideo,
  };
}