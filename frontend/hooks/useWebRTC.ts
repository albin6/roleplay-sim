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
  const remoteStreamRef = useRef<MediaStream>(new MediaStream());
  const pendingIceCandidatesRef = useRef<RTCIceCandidateInit[]>([]);

  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStream, setRemoteStream] = useState<MediaStream | null>(null);
  const [connectionState, setConnectionState] = useState<RTCPeerConnectionState>("new");
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoOff, setIsVideoOff] = useState(false);

  const seat = useRoomStore((s) => s.seat);
  const roomState = useRoomStore((s) => s.state);

  // 1. Initialize local media (handles single-webcam concurrency gracefully)
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
      })
      .catch((err) => {
        console.warn("Could not acquire video+audio (camera may be busy), falling back to audio:", err);
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

  // 2. Initialize PeerConnection
  const initPeerConnection = useCallback(() => {
    if (pcRef.current) return pcRef.current;

    const pc = new RTCPeerConnection(ICE_SERVERS);
    pcRef.current = pc;

    // Ensure transceivers exist for both audio and video so we receive remote tracks
    // even if local device only has audio
    try {
      pc.addTransceiver("audio", { direction: "sendrecv" });
      pc.addTransceiver("video", { direction: "sendrecv" });
    } catch (e) {
      console.warn("Transceiver initialization notice:", e);
    }

    // Attach any existing local stream tracks
    if (localStream) {
      localStream.getTracks().forEach((track) => {
        pc.addTrack(track, localStream);
      });
    }

    // Handle incoming remote media tracks
    pc.ontrack = (event) => {
      console.log("WebRTC ontrack received:", event.track.kind, event.track.id);
      remoteStreamRef.current.addTrack(event.track);
      // Create a new MediaStream instance to guarantee React state re-render
      setRemoteStream(new MediaStream(remoteStreamRef.current.getTracks()));
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

  // 3. Dynamically sync local tracks to PeerConnection when localStream arrives
  useEffect(() => {
    if (!pcRef.current || !localStream) return;
    const pc = pcRef.current;
    const senders = pc.getSenders();

    localStream.getTracks().forEach((track) => {
      const sender = senders.find((s) => s.track?.kind === track.kind);
      if (sender) {
        sender.replaceTrack(track);
      } else {
        pc.addTrack(track, localStream);
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

          // Drain queued ICE candidates
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

  // 5. Seat A initiates WebRTC Offer when entering signaling phase
  useEffect(() => {
    if (roomState === "signaling" && seat === "A") {
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
        .catch((err) => console.error("Failed to generate offer:", err));
    }
  }, [roomState, seat, initPeerConnection, sendSignal]);

  const toggleMute = () => {
    if (!localStream) return;
    const audioTrack = localStream.getAudioTracks()[0];
    if (audioTrack) {
      audioTrack.enabled = !audioTrack.enabled;
      setIsMuted(!audioTrack.enabled);
    }
  };

  const toggleVideo = () => {
    if (!localStream) return;
    const videoTrack = localStream.getVideoTracks()[0];
    if (videoTrack) {
      videoTrack.enabled = !videoTrack.enabled;
      setIsVideoOff(!videoTrack.enabled);
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