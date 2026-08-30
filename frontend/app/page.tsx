"use client";

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "../stores/authStore";
import { useRoomStore } from "../stores/roomStore";
import { useWebSocket } from "../hooks/useWebSocket";
import { api } from "../lib/api";
import {
  Shield,
  Trophy,
  Users,
  Swords,
  LogOut,
  Sparkles,
  ArrowRight,
  UserCheck,
} from "lucide-react";

export default function DashboardPage() {
  const router = useRouter();
  const { user, token, loadUser, logout } = useAuthStore();
  const { roomId, state: roomState } = useRoomStore();
  const { send } = useWebSocket();

  const [isQueuing, setIsQueuing] = useState(false);
  const [selectedDifficulty, setSelectedDifficulty] = useState("medium");

  useEffect(() => {
    if (token && !user) {
      loadUser();
    }
  }, [token, user, loadUser]);

  // Navigate to room as soon as room is ready
  useEffect(() => {
    if (roomId && (roomState === "ready" || roomState === "spinning" || roomState === "scenario" || roomState === "prep" || roomState === "live")) {
      router.push(`/room/${roomId}`);
    }
  }, [roomId, roomState, router]);

  const handleJoinQueue = async () => {
    if (!token) {
      router.push("/login");
      return;
    }

    setIsQueuing(true);
    try {
      await api.enqueueMatchmaking({ preferred_difficulty: selectedDifficulty });
      // Also send over WS to register interest
      send("JOIN_QUEUE", { difficulty: selectedDifficulty });
    } catch (err: any) {
      alert("Failed to join queue: " + err.message);
      setIsQueuing(false);
    }
  };

  const handleLeaveQueue = async () => {
    try {
      await api.dequeueMatchmaking();
      send("LEAVE_QUEUE");
    } finally {
      setIsQueuing(false);
    }
  };

  return (
    <main className="min-h-screen bg-slate-950 flex flex-col items-center justify-between p-6">
      {/* Top Navbar */}
      <nav className="w-full max-w-6xl flex items-center justify-between py-4 border-b border-slate-800">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-indigo-600 flex items-center justify-center text-white shadow-lg shadow-indigo-600/30">
            <Swords className="w-5 h-5" />
          </div>
          <div>
            <h1 className="font-black text-lg text-white">Roleplay Simulator</h1>
            <p className="text-xs text-slate-400">P2P Workplace Simulation Platform</p>
          </div>
        </div>

        <div className="flex items-center gap-4">
          <button
            onClick={() => router.push("/leaderboard")}
            className="flex items-center gap-1.5 text-xs font-semibold text-slate-300 hover:text-white px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 transition"
          >
            <Trophy className="w-4 h-4 text-amber-400" />
            <span>Leaderboard</span>
          </button>

          {user ? (
            <div className="flex items-center gap-3 pl-4 border-l border-slate-800">
              <div className="text-right">
                <div className="text-xs font-bold text-white">{user.display_name}</div>
                <div className="text-[11px] text-indigo-400 font-mono">
                  Rating: {Math.round(user.elo_rating)}
                </div>
              </div>
              <button
                onClick={logout}
                className="p-2 text-slate-400 hover:text-rose-400 rounded-lg bg-slate-900 border border-slate-800 transition"
                title="Log Out"
              >
                <LogOut className="w-4 h-4" />
              </button>
            </div>
          ) : (
            <div className="flex items-center gap-2">
              <button
                onClick={() => router.push("/login")}
                className="text-xs font-bold text-slate-300 hover:text-white px-4 py-2"
              >
                Sign In
              </button>
              <button
                onClick={() => router.push("/register")}
                className="text-xs font-bold text-white bg-indigo-600 hover:bg-indigo-500 px-4 py-2 rounded-xl transition shadow-lg shadow-indigo-600/20"
              >
                Sign Up
              </button>
            </div>
          )}
        </div>
      </nav>

      {/* Hero / Matchmaking Card */}
      <section className="w-full max-w-4xl flex flex-col items-center text-center my-12">
        <span className="text-xs font-bold uppercase tracking-widest text-indigo-400 bg-indigo-950/60 border border-indigo-500/30 px-3 py-1 rounded-full mb-4">
          Real-Time P2P Workplace Scenarios
        </span>
        <h2 className="text-4xl md:text-5xl font-black text-white tracking-tight mb-4 max-w-2xl">
          Master negotiation and leadership through live peer roleplay.
        </h2>
        <p className="text-base text-slate-400 max-w-xl mb-10">
          Match with peers in real-time, spin for dynamic roles, and engage in high-stakes workplace
          dilemmas evaluated by AI rubric analysis.
        </p>

        {/* Matchmaking Queue Box */}
        <div className="w-full max-w-md p-6 bg-slate-900/80 rounded-3xl border border-slate-800 shadow-2xl backdrop-blur-md">
          <div className="flex items-center justify-between mb-4 text-xs font-semibold text-slate-400">
            <span>Select Scenario Difficulty:</span>
            <span className="uppercase text-indigo-400 font-bold">{selectedDifficulty}</span>
          </div>

          {/* Difficulty Tier Selector */}
          <div className="grid grid-cols-3 gap-2 mb-6">
            {["easy", "medium", "hard"].map((diff) => (
              <button
                key={diff}
                onClick={() => setSelectedDifficulty(diff)}
                disabled={isQueuing}
                className={`py-2.5 rounded-xl text-xs font-bold capitalize transition-all border ${
                  selectedDifficulty === diff
                    ? "bg-indigo-600 text-white border-indigo-500 shadow-md shadow-indigo-600/30"
                    : "bg-slate-950 text-slate-400 border-slate-800 hover:bg-slate-800"
                }`}
              >
                {diff}
              </button>
            ))}
          </div>

          {/* Action Button */}
          {isQueuing ? (
            <div className="flex flex-col items-center gap-3">
              <div className="flex items-center gap-2 text-indigo-400 text-sm font-semibold">
                <div className="w-4 h-4 rounded-full border-2 border-indigo-400 border-t-transparent animate-spin" />
                <span>Searching for matched opponent...</span>
              </div>
              <button
                onClick={handleLeaveQueue}
                className="text-xs text-rose-400 hover:underline mt-1"
              >
                Cancel Search
              </button>
            </div>
          ) : (
            <button
              onClick={handleJoinQueue}
              className="w-full py-4 bg-indigo-600 hover:bg-indigo-500 text-white rounded-2xl font-bold flex items-center justify-center gap-2 transition-all shadow-xl shadow-indigo-600/30 group"
            >
              <span>Find Simulation Match</span>
              <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </button>
          )}
        </div>
      </section>

      {/* Feature Highlights Footer */}
      <footer className="w-full max-w-5xl grid grid-cols-1 md:grid-cols-3 gap-4 border-t border-slate-800/80 pt-8 pb-4 text-left">
        <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800/60">
          <div className="w-8 h-8 rounded-lg bg-indigo-600/20 text-indigo-400 flex items-center justify-center mb-3">
            <Users className="w-4 h-4" />
          </div>
          <h3 className="text-sm font-bold text-white mb-1">Synchronized Dual Spin</h3>
          <p className="text-xs text-slate-400">
            Both peers spin identical context wheels for deterministic, balanced role assignment.
          </p>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800/60">
          <div className="w-8 h-8 rounded-lg bg-purple-600/20 text-purple-400 flex items-center justify-center mb-3">
            <Shield className="w-4 h-4" />
          </div>
          <h3 className="text-sm font-bold text-white mb-1">Conflicting Objectives</h3>
          <p className="text-xs text-slate-400">
            Private instructions and constraints challenge negotiation and emotional regulation skills.
          </p>
        </div>

        <div className="p-4 rounded-2xl bg-slate-900/40 border border-slate-800/60">
          <div className="w-8 h-8 rounded-lg bg-emerald-600/20 text-emerald-400 flex items-center justify-center mb-3">
            <Sparkles className="w-4 h-4" />
          </div>
          <h3 className="text-sm font-bold text-white mb-1">AI Rubric & Elo Updates</h3>
          <p className="text-xs text-slate-400">
            Dual-channel Deepgram STT and Gemini 1.5 evaluate communication clarity and update ratings.
          </p>
        </div>
      </footer>
    </main>
  );
}