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
  const [localStream, setLocalStream] = useState<MediaStream | null>(null);
  const [remoteStream, setRemoteStream] = useState<MediaStream | null>(null);
  const [connectionState, setConnectionState] = useState<RTCPeerConnectionState>("new");
  const [isMuted, setIsMuted] = useState(false);
  const [isVideoOff, setIsVideoOff] = useState(false);

  const seat = useRoomStore((s) => s.seat);
  const roomState = useRoomStore((s) => s.state);

  // Initialize local media
  useEffect(() => {
    let stream: MediaStream;

    navigator.mediaDevices
      .getUserMedia({ audio: true, video: true })
      .then((s) => {
        stream = s;
        setLocalStream(s);
      })
      .catch((err) => {
        console.warn("Could not get camera/mic, falling back to audio only:", err);
        navigator.mediaDevices
          .getUserMedia({ audio: true, video: false })
          .then((s) => {
            stream = s;
            setLocalStream(s);
          })
          .catch((e) => console.error("Could not capture any media:", e));
      });

    return () => {
      if (stream) {
        stream.getTracks().forEach((track) => track.stop());
      }
    };
  }, []);

  // Initialize PeerConnection
  const initPeerConnection = useCallback(() => {
    if (pcRef.current) return pcRef.current;

    const pc = new RTCPeerConnection(ICE_SERVERS);
    pcRef.current = pc;

    // Attach local stream tracks
    if (localStream) {
      localStream.getTracks().forEach((track) => pc.addTrack(track, localStream));
    }

    pc.ontrack = (event) => {
      console.log("Received remote stream track:", event.streams[0]);
      if (event.streams && event.streams[0]) {
        setRemoteStream(event.streams[0]);
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
      console.log("WebRTC state changed:", pc.connectionState);
      setConnectionState(pc.connectionState);
    };

    return pc;
  }, [localStream, sendSignal]);

  // Handle incoming signals from the other peer
  useEffect(() => {
    const unsubscribe = onSignalReceived(async (raw) => {
      const pc = initPeerConnection();
      const signal = raw.signal || raw;

      try {
        if (signal.type === "offer") {
          console.log("Handling WebRTC offer");
          await pc.setRemoteDescription(new RTCSessionDescription(signal));
          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);
          sendSignal({
            type: "answer",
            sdp: answer.sdp,
          });
        } else if (signal.type === "answer") {
          console.log("Handling WebRTC answer");
          await pc.setRemoteDescription(new RTCSessionDescription(signal));
        } else if (signal.type === "ice" && signal.candidate) {
          console.log("Handling ICE candidate");
          await pc.addIceCandidate(new RTCIceCandidate(signal.candidate));
        }
      } catch (err) {
        console.error("Failed to process WebRTC signal:", err);
      }
    });

    return unsubscribe;
  }, [onSignalReceived, initPeerConnection, sendSignal]);

  // Trigger Offer if we are Seat A when transitioning to signaling phase
  useEffect(() => {
    if (roomState === "signaling" && seat === "A") {
      const pc = initPeerConnection();
      pc.createOffer()
        .then((offer) => pc.setLocalDescription(offer).then(() => offer))
        .then((offer) => {
          console.log("Seat A sending WebRTC offer");
          sendSignal({
            type: "offer",
            sdp: offer.sdp,
          });
        })
        .catch((err) => console.error("Failed to create offer:", err));
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