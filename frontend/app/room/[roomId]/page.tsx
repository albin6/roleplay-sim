"use client";

import React, { useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import { useRoomStore } from "../../../stores/roomStore";
import { useAuthStore } from "../../../stores/authStore";
import { useWebSocket } from "../../../hooks/useWebSocket";
import { useWebRTC } from "../../../hooks/useWebRTC";
import { DualWheel } from "../../../components/wheel/DualWheel";
import { VideoGrid } from "../../../components/room/VideoGrid";
import { CallControls } from "../../../components/room/CallControls";
import { useAudioRecorder } from "../../../hooks/useAudioRecorder";
import ScorecardModal from "../../../components/room/ScorecardModal";
import { Shield, Sparkles, CheckCircle2, AlertTriangle, ArrowRight, Home } from "lucide-react";

export default function RoleplayRoomPage() {
  const params = useParams();
  const router = useRouter();
  const roomId = params?.roomId as string;

  const { user } = useAuthStore();
  const {
    state,
    peer,
    seat,
    spinResult,
    scenario,
    prepSecondsRemaining,
    sessionSecondsRemaining,
    isMeReady,
    evaluation,
    setMeReady,
    setRoomId,
  } = useRoomStore();

  const { isConnected, send, sendBinary, on } = useWebSocket();

  const {
    localStream,
    remoteStream,
    connectionState,
    isMuted,
    isVideoOff,
    toggleMute,
    toggleVideo,
  } = useWebRTC({
    roomId,
    sendSignal: (signal) => send("SIGNAL", { room_id: roomId, signal }),
    onSignalReceived: (handler) => on("SIGNAL", handler),
  });

  const { isRecording, chunksSent } = useAudioRecorder({
    localStream,
    isLive: state === "live",
    sendBinary,
  });

  useEffect(() => {
    if (roomId) {
      setRoomId(roomId);
    }
  }, [roomId, setRoomId]);

  const handlePrepReady = () => {
    setMeReady();
    send("PREP_READY", { room_id: roomId });
  };

  const handleEndCall = () => {
    send("SESSION_END", { room_id: roomId, reason: "user_ended" });
  };

  return (
    <main className="min-h-screen bg-slate-950 text-slate-100 flex flex-col items-center justify-center p-4">
      {/* Header bar */}
      <header className="w-full max-w-6xl flex items-center justify-between py-4 mb-4 border-b border-slate-800">
        <div className="flex items-center gap-3">
          <span className="font-bold text-lg text-indigo-400">Roleplay Simulator</span>
          <span className="text-xs bg-slate-800 text-slate-400 px-2 py-0.5 rounded-md border border-slate-700">
            Room: {roomId}
          </span>
        </div>

        <div className="flex items-center gap-4 text-sm">
          <div className="flex items-center gap-2">
            <span
              className={`w-2 h-2 rounded-full ${
                isConnected ? "bg-emerald-500" : "bg-rose-500 animate-ping"
              }`}
            />
            <span className="text-slate-400 text-xs">{isConnected ? "Connected" : "Reconnecting..."}</span>
          </div>

          <button
            onClick={() => router.push("/")}
            className="flex items-center gap-1.5 text-xs text-slate-400 hover:text-white transition"
          >
            <Home className="w-4 h-4" />
            <span>Dashboard</span>
          </button>
        </div>
      </header>

      {/* Main Stage View depending on Room State */}
      <div className="w-full max-w-6xl flex flex-col items-center justify-center flex-1">
        {/* Phase 1: Waiting for Peer */}
        {state === "waiting" && (
          <div className="flex flex-col items-center text-center p-12 bg-slate-900/60 rounded-3xl border border-slate-800 backdrop-blur-md max-w-md">
            <div className="relative w-24 h-24 mb-6 flex items-center justify-center">
              <div className="absolute inset-0 rounded-full border-2 border-indigo-500/30 animate-ping" />
              <div className="w-16 h-16 rounded-full bg-indigo-600/20 border border-indigo-500 flex items-center justify-center text-indigo-400">
                <Shield className="w-8 h-8" />
              </div>
            </div>
            <h2 className="text-xl font-bold text-white mb-2">Waiting for Peer</h2>
            <p className="text-sm text-slate-400">
              Matched! Waiting for opponent to establish real-time connection...
            </p>
          </div>
        )}

        {/* Phase 2: Dual Wheel Spin */}
        {(state === "ready" || state === "spinning") && (
          <DualWheel isSpinning={state === "spinning"} spinResult={spinResult} />
        )}

        {/* Phase 3: Private Briefing & Prep Countdown */}
        {(state === "scenario" || state === "prep") && scenario && (
          <div className="w-full max-w-2xl bg-slate-900 rounded-3xl border border-slate-800 p-8 shadow-2xl">
            <div className="flex items-center justify-between mb-6 pb-4 border-b border-slate-800">
              <div>
                <span className="text-xs uppercase font-bold tracking-wider text-indigo-400">
                  Private Scenario Briefing
                </span>
                <h1 className="text-2xl font-black text-white mt-1">{scenario.title}</h1>
              </div>
              <div className="text-right">
                <span className="text-xs text-slate-400">Prep Countdown</span>
                <div className="text-2xl font-mono font-bold text-amber-400">
                  {prepSecondsRemaining}s
                </div>
              </div>
            </div>

            <div className="space-y-6">
              <div>
                <h3 className="text-xs uppercase font-semibold text-slate-400 mb-2">Context</h3>
                <p className="text-sm text-slate-300 leading-relaxed bg-slate-950/60 p-4 rounded-xl border border-slate-800">
                  {scenario.background_context}
                </p>
              </div>

              <div className="p-4 rounded-xl bg-indigo-950/40 border border-indigo-500/30">
                <div className="flex items-center gap-2 text-indigo-300 font-bold text-sm mb-1">
                  <Sparkles className="w-4 h-4" />
                  <span>Your Secret Objective (Keep Hidden)</span>
                </div>
                <p className="text-sm text-indigo-100 font-medium">{scenario.your_objective}</p>
              </div>

              {scenario.your_constraints?.length > 0 && (
                <div>
                  <h3 className="text-xs uppercase font-semibold text-slate-400 mb-2">
                    Behavioral Constraints
                  </h3>
                  <ul className="space-y-2">
                    {scenario.your_constraints.map((c, i) => (
                      <li
                        key={i}
                        className="flex items-start gap-2 text-xs text-slate-300 bg-slate-950/40 p-2.5 rounded-lg border border-slate-800/80"
                      >
                        <AlertTriangle className="w-4 h-4 text-amber-400 shrink-0 mt-0.5" />
                        <span>{c}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}

              <div className="pt-4 flex justify-end">
                <button
                  onClick={handlePrepReady}
                  disabled={isMeReady}
                  className={`flex items-center gap-2 px-6 py-3 rounded-xl font-bold transition-all shadow-lg ${
                    isMeReady
                      ? "bg-slate-800 text-emerald-400 border border-emerald-500/30 cursor-default"
                      : "bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-600/30"
                  }`}
                >
                  <CheckCircle2 className="w-5 h-5" />
                  <span>{isMeReady ? "Ready! Waiting for peer..." : "I'm Ready to Start"}</span>
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Phase 4: Live P2P Roleplay Video Call */}
        {(state === "signaling" || state === "live") && (
          <div className="w-full flex flex-col gap-4">
            <VideoGrid
              localStream={localStream}
              remoteStream={remoteStream}
              myDisplayName={user?.display_name || "You"}
              peerDisplayName={peer?.displayName || "Peer"}
              peerEloRating={peer?.eloRating || 1200}
              connectionState={connectionState}
            />

            {/* Audio Recording Status Indicator */}
            {isRecording && (
              <div className="self-center flex items-center gap-2 px-4 py-1.5 rounded-full bg-rose-500/10 border border-rose-500/30 text-rose-400 text-xs font-semibold animate-pulse">
                <span className="w-2.5 h-2.5 rounded-full bg-rose-500" />
                <span>AI Dual-Channel Audio Active ({chunksSent} chunks streamed)</span>
              </div>
            )}

            <CallControls
              isMuted={isMuted}
              isVideoOff={isVideoOff}
              onToggleMute={toggleMute}
              onToggleVideo={toggleVideo}
              onEndCall={handleEndCall}
              secondsRemaining={sessionSecondsRemaining}
            />

            {/* In-Call Objectives Cheat Sheet */}
            {scenario && (
              <div className="p-4 bg-slate-900/60 rounded-xl border border-slate-800 text-xs flex items-center justify-between">
                <div>
                  <span className="text-slate-400 font-semibold uppercase">Your Goal: </span>
                  <span className="text-indigo-300">{scenario.your_objective}</span>
                </div>
                <span className="text-slate-500 font-mono">Seat {seat}</span>
              </div>
            )}
          </div>
        )}

        {/* Phase 5: Evaluating */}
        {state === "evaluating" && (
          <div className="flex flex-col items-center text-center p-12 bg-slate-900/60 rounded-3xl border border-slate-800 backdrop-blur-md max-w-md">
            <div className="w-20 h-20 rounded-full bg-indigo-600/20 border border-indigo-500/50 flex items-center justify-center text-indigo-400 mb-6 animate-pulse">
              <Sparkles className="w-10 h-10" />
            </div>
            <h2 className="text-xl font-bold text-white mb-2">Evaluating Roleplay</h2>
            <p className="text-sm text-slate-400 leading-relaxed">
              Synthesizing multichannel audio transcripts via Deepgram Nova-2 and running Gemini 1.5
              rubric assessment...
            </p>
          </div>
        )}

        {/* Phase 6: Complete & Rich Scorecard Modal */}
        {state === "complete" && evaluation && (
          <ScorecardModal
            data={evaluation}
            scenarioTitle={scenario?.title}
            onClose={() => router.push("/")}
          />
        )}
      </div>
    </main>
  );
}