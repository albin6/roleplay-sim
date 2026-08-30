"use client";

import React, { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { api } from "../../lib/api";
import { Trophy, ArrowLeft, Shield, User } from "lucide-react";

interface LeaderboardUser {
  rank: number;
  user_id: string;
  display_name: string;
  avatar_url?: string | null;
  elo_rating: number;
  total_sessions: number;
  wins: number;
}

export default function LeaderboardPage() {
  const router = useRouter();
  const [users, setUsers] = useState<LeaderboardUser[]>([]);
  const [myRank, setMyRank] = useState<number | null>(null);
  const [myElo, setMyElo] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api
      .getLeaderboard(1, 50)
      .then((res) => {
        setUsers(res.data || []);
        setMyRank(res.my_rank);
        setMyElo(res.my_elo);
      })
      .catch((err) => console.error("Failed to load leaderboard:", err))
      .finally(() => setLoading(false));
  }, []);

  return (
    <main className="min-h-screen bg-slate-950 p-6 flex flex-col items-center">
      <div className="w-full max-w-4xl">
        {/* Navigation */}
        <button
          onClick={() => router.push("/")}
          className="flex items-center gap-2 text-xs font-semibold text-slate-400 hover:text-white mb-6 transition"
        >
          <ArrowLeft className="w-4 h-4" />
          <span>Back to Dashboard</span>
        </button>

        {/* Title */}
        <div className="flex items-center justify-between mb-8 pb-4 border-b border-slate-800">
          <div>
            <span className="text-xs uppercase font-bold tracking-wider text-indigo-400">
              Global Standings
            </span>
            <h1 className="text-3xl font-black text-white mt-1">Leaderboard</h1>
          </div>

          {myRank && (
            <div className="flex items-center gap-4 bg-slate-900 px-4 py-2 rounded-2xl border border-slate-800">
              <div>
                <div className="text-[10px] text-slate-400 uppercase font-semibold">Your Rank</div>
                <div className="text-lg font-black text-indigo-400">#{myRank}</div>
              </div>
              <div className="border-l border-slate-800 pl-4">
                <div className="text-[10px] text-slate-400 uppercase font-semibold">Your Rating</div>
                <div className="text-lg font-mono font-bold text-white">{Math.round(myElo || 1200)}</div>
              </div>
            </div>
          )}
        </div>

        {/* Leaderboard Table */}
        {loading ? (
          <div className="flex justify-center p-12 text-slate-500 text-sm">
            Loading leaderboard data...
          </div>
        ) : users.length === 0 ? (
          <div className="p-12 text-center text-slate-500 bg-slate-900/40 rounded-2xl border border-slate-800">
            No simulation rankings recorded yet. Be the first to play!
          </div>
        ) : (
          <div className="bg-slate-900/60 rounded-3xl border border-slate-800 overflow-hidden shadow-2xl">
            <table className="w-full text-left text-sm text-slate-300">
              <thead className="bg-slate-900/90 text-[11px] uppercase tracking-wider text-slate-400 border-b border-slate-800">
                <tr>
                  <th className="px-6 py-4">Rank</th>
                  <th className="px-6 py-4">Player</th>
                  <th className="px-6 py-4 text-center">Sessions</th>
                  <th className="px-6 py-4 text-center">Wins</th>
                  <th className="px-6 py-4 text-right">Elo Rating</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60">
                {users.map((u) => (
                  <tr key={u.user_id} className="hover:bg-slate-800/40 transition">
                    <td className="px-6 py-4 font-mono font-bold">
                      {u.rank === 1 ? (
                        <span className="text-amber-400 flex items-center gap-1">
                          <Trophy className="w-4 h-4" /> #1
                        </span>
                      ) : u.rank === 2 ? (
                        <span className="text-slate-300">#2</span>
                      ) : u.rank === 3 ? (
                        <span className="text-amber-700">#3</span>
                      ) : (
                        `#${u.rank}`
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-full bg-slate-800 flex items-center justify-center text-slate-400 border border-slate-700">
                          <User className="w-4 h-4" />
                        </div>
                        <span className="font-bold text-white">{u.display_name}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 text-center font-mono">{u.total_sessions}</td>
                    <td className="px-6 py-4 text-center font-mono text-emerald-400">{u.wins}</td>
                    <td className="px-6 py-4 text-right font-mono font-bold text-indigo-400">
                      {Math.round(u.elo_rating)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </main>
  );
}