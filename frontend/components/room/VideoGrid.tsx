"use client";

import React, { useEffect, useRef, useState } from "react";
import { User, Shield } from "lucide-react";

interface VideoGridProps {
  localStream: MediaStream | null;
  remoteStream: MediaStream | null;
  myDisplayName: string;
  peerDisplayName: string;
  peerEloRating: number;
  connectionState: RTCPeerConnectionState;
}

export function VideoGrid({
  localStream,
  remoteStream,
  myDisplayName,
  peerDisplayName,
  peerEloRating,
  connectionState,
}: VideoGridProps) {
  const localVideoRef = useRef<HTMLVideoElement | null>(null);
  const remoteVideoRef = useRef<HTMLVideoElement | null>(null);
  const [, setRerender] = useState(0);

  // Sync local stream to local video element
  useEffect(() => {
    if (localVideoRef.current && localStream) {
      localVideoRef.current.srcObject = localStream;
    }
  }, [localStream]);

  // Sync remote stream to remote video element
  useEffect(() => {
    if (remoteVideoRef.current && remoteStream) {
      remoteVideoRef.current.srcObject = remoteStream;
      remoteVideoRef.current.play().catch((err) => {
        console.warn("Remote video play pending:", err);
      });
    }
  }, [remoteStream]);

  // Track event listeners for track enable/mute changes
  useEffect(() => {
    if (!remoteStream) return;
    const forceUpdate = () => setRerender((n) => n + 1);

    remoteStream.getVideoTracks().forEach((track) => {
      track.addEventListener("mute", forceUpdate);
      track.addEventListener("unmute", forceUpdate);
      track.addEventListener("ended", forceUpdate);
    });

    return () => {
      remoteStream.getVideoTracks().forEach((track) => {
        track.removeEventListener("mute", forceUpdate);
        track.removeEventListener("unmute", forceUpdate);
        track.removeEventListener("ended", forceUpdate);
      });
    };
  }, [remoteStream]);

  const hasRemoteVideo =
    remoteStream &&
    remoteStream.getVideoTracks().some((t) => t.enabled && t.readyState === "live");

  const hasLocalVideo =
    localStream &&
    localStream.getVideoTracks().some((t) => t.enabled && t.readyState === "live");

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4 w-full h-[450px]">
      {/* Remote Peer View (Left / Main) */}
      <div className="relative rounded-2xl bg-slate-900 border border-slate-700 overflow-hidden flex items-center justify-center shadow-lg">
        {hasRemoteVideo ? (
          <video
            ref={remoteVideoRef}
            autoPlay
            playsInline
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="flex flex-col items-center gap-3 text-slate-500">
            <div className="w-20 h-20 rounded-full bg-slate-800 flex items-center justify-center border border-slate-700">
              <User className="w-10 h-10 text-slate-400" />
            </div>
            <span className="text-sm font-medium">
              {connectionState === "connected"
                ? "Camera Off"
                : `Connecting... (${connectionState})`}
            </span>
          </div>
        )}

        {/* Peer Info Tag */}
        <div className="absolute top-4 left-4 flex items-center gap-2 bg-slate-900/80 backdrop-blur-md px-3 py-1.5 rounded-lg border border-slate-700 text-xs font-semibold text-white">
          <span>{peerDisplayName || "Opponent"}</span>
          <span className="flex items-center gap-1 text-indigo-400 bg-indigo-950/60 px-1.5 py-0.5 rounded">
            <Shield className="w-3 h-3" />
            {Math.round(peerEloRating || 1200)}
          </span>
        </div>

        {/* Connection State indicator */}
        <div className="absolute top-4 right-4 flex items-center gap-1.5 bg-slate-900/80 backdrop-blur-md px-2.5 py-1 rounded-full border border-slate-700 text-[10px] text-slate-300">
          <span
            className={`w-2 h-2 rounded-full ${
              connectionState === "connected"
                ? "bg-emerald-500"
                : connectionState === "connecting"
                ? "bg-amber-500 animate-pulse"
                : "bg-rose-500"
            }`}
          />
          <span className="capitalize">{connectionState}</span>
        </div>
      </div>

      {/* Local User View (Right) */}
      <div className="relative rounded-2xl bg-slate-900 border border-slate-700 overflow-hidden flex items-center justify-center shadow-lg">
        {hasLocalVideo ? (
          <video
            ref={localVideoRef}
            autoPlay
            muted
            playsInline
            className="w-full h-full object-cover -scale-x-100"
          />
        ) : (
          <div className="flex flex-col items-center gap-3 text-slate-500">
            <div className="w-20 h-20 rounded-full bg-slate-800 flex items-center justify-center border border-slate-700">
              <User className="w-10 h-10 text-slate-400" />
            </div>
            <span className="text-sm font-medium">Your Camera is Off</span>
          </div>
        )}

        {/* Local Info Tag */}
        <div className="absolute top-4 left-4 bg-slate-900/80 backdrop-blur-md px-3 py-1.5 rounded-lg border border-slate-700 text-xs font-semibold text-slate-300">
          {myDisplayName || "You"} (Self)
        </div>
      </div>
    </div>
  );
}