"use client";

import React from "react";
import { Mic, MicOff, Video, VideoOff, PhoneOff, Clock } from "lucide-react";

interface CallControlsProps {
  isMuted: boolean;
  isVideoOff: boolean;
  onToggleMute: () => void;
  onToggleVideo: () => void;
  onEndCall: () => void;
  secondsRemaining: number;
}

export function CallControls({
  isMuted,
  isVideoOff,
  onToggleMute,
  onToggleVideo,
  onEndCall,
  secondsRemaining,
}: CallControlsProps) {
  const formatTime = (secs: number) => {
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m}:${s < 10 ? "0" : ""}${s}`;
  };

  return (
    <div className="flex items-center justify-between px-6 py-4 bg-slate-900/90 backdrop-blur-md rounded-2xl border border-slate-700 shadow-xl w-full">
      {/* Session Timer */}
      <div className="flex items-center gap-2 text-indigo-400 font-mono text-lg font-bold bg-slate-800/80 px-3 py-1.5 rounded-lg border border-slate-700">
        <Clock className="w-5 h-5 text-indigo-400 animate-pulse" />
        <span>{formatTime(secondsRemaining)}</span>
      </div>

      {/* Main Buttons */}
      <div className="flex items-center gap-3">
        <button
          onClick={onToggleMute}
          className={`p-3.5 rounded-xl border transition-all ${
            isMuted
              ? "bg-rose-500/20 border-rose-500/50 text-rose-400 hover:bg-rose-500/30"
              : "bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700"
          }`}
          title={isMuted ? "Unmute Microphone" : "Mute Microphone"}
        >
          {isMuted ? <MicOff className="w-5 h-5" /> : <Mic className="w-5 h-5" />}
        </button>

        <button
          onClick={onToggleVideo}
          className={`p-3.5 rounded-xl border transition-all ${
            isVideoOff
              ? "bg-rose-500/20 border-rose-500/50 text-rose-400 hover:bg-rose-500/30"
              : "bg-slate-800 border-slate-700 text-slate-300 hover:bg-slate-700"
          }`}
          title={isVideoOff ? "Turn Video On" : "Turn Video Off"}
        >
          {isVideoOff ? <VideoOff className="w-5 h-5" /> : <Video className="w-5 h-5" />}
        </button>

        <button
          onClick={onEndCall}
          className="flex items-center gap-2 px-5 py-3.5 rounded-xl bg-rose-600 hover:bg-rose-500 text-white font-semibold transition-colors shadow-lg shadow-rose-600/30 ml-4"
        >
          <PhoneOff className="w-5 h-5" />
          <span>End Roleplay</span>
        </button>
      </div>

      {/* Status note */}
      <div className="hidden md:flex text-xs text-slate-400">
        Audio recording active for AI feedback
      </div>
    </div>
  );
}