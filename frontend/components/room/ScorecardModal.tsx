"use client";

import React from "react";
import Link from "next/link";
import { EvaluationData } from "@/lib/types";
import {
  Trophy,
  CheckCircle2,
  XCircle,
  TrendingUp,
  TrendingDown,
  Sparkles,
  ArrowRight,
  Shield,
  Layers,
  Award,
} from "lucide-react";
import {
  Radar,
  RadarChart,
  PolarGrid,
  PolarAngleAxis,
  PolarRadiusAxis,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  Cell,
} from "recharts";

interface ScorecardModalProps {
  data: EvaluationData;
  scenarioTitle?: string;
  onClose: () => void;
}

const DIMENSION_LABELS: Record<string, string> = {
  communication_clarity: "Communication",
  active_listening: "Listening",
  negotiation_strategy: "Negotiation",
  emotional_regulation: "Regulation",
  empathy: "Empathy",
  objective_alignment: "Goal Alignment",
};

export default function ScorecardModal({
  data,
  scenarioTitle,
  onClose,
}: ScorecardModalProps) {
  const { your_score, peer_score } = data;

  const chartData = (your_score.rubric_scores || []).map((r) => ({
    dimension: DIMENSION_LABELS[r.dimension] || r.dimension,
    score: r.score,
    fullMark: 10,
  }));

  const isEloPositive = your_score.elo_delta >= 0;

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto bg-slate-950/85 backdrop-blur-md flex items-center justify-center p-4 sm:p-6">
      <div className="bg-slate-900 border border-slate-700/80 rounded-2xl max-w-4xl w-full max-h-[92vh] overflow-y-auto shadow-2xl text-slate-100 flex flex-col">
        {/* Header */}
        <div className="p-6 border-b border-slate-800 flex items-center justify-between bg-slate-900/60 sticky top-0 z-10 backdrop-blur-md">
          <div className="flex items-center space-x-3">
            <div className="p-2.5 bg-indigo-500/10 border border-indigo-500/30 rounded-xl text-indigo-400">
              <Award className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-xl font-bold text-white flex items-center gap-2">
                Simulation Scorecard
                <span className="text-xs font-semibold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                  AI Evaluated
                </span>
              </h2>
              <p className="text-xs text-slate-400">
                {scenarioTitle || "Workplace Negotiation"} • Session ID:{" "}
                <span className="font-mono text-slate-300">
                  {data.session_id.slice(0, 8)}...
                </span>
              </p>
            </div>
          </div>
          <Link
            href="/leaderboard"
            className="text-xs font-semibold px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white transition-colors flex items-center gap-1.5"
          >
            <span>Leaderboard</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>

        {/* Content Body */}
        <div className="p-6 space-y-6">
          {/* Top Metric Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
            {/* Overall Score */}
            <div className="p-4 rounded-xl bg-slate-800/60 border border-slate-700 flex flex-col justify-between">
              <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                Overall Score
              </span>
              <div className="mt-2 flex items-baseline gap-2">
                <span className="text-3xl font-extrabold text-white">
                  {your_score.overall_score.toFixed(1)}
                </span>
                <span className="text-sm font-semibold text-slate-400">/ 100</span>
              </div>
              <div className="mt-3 w-full bg-slate-700/50 rounded-full h-2 overflow-hidden">
                <div
                  className={`h-full rounded-full transition-all duration-500 ${
                    your_score.overall_score >= 75
                      ? "bg-emerald-500"
                      : your_score.overall_score >= 50
                      ? "bg-amber-500"
                      : "bg-rose-500"
                  }`}
                  style={{ width: `${Math.min(100, Math.max(0, your_score.overall_score))}%` }}
                />
              </div>
            </div>

            {/* Objective Achieved */}
            <div className="p-4 rounded-xl bg-slate-800/60 border border-slate-700 flex flex-col justify-between">
              <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                Objective Status
              </span>
              <div className="mt-2 flex items-center gap-2">
                {your_score.objective_achieved ? (
                  <CheckCircle2 className="w-6 h-6 text-emerald-400 shrink-0" />
                ) : (
                  <XCircle className="w-6 h-6 text-rose-400 shrink-0" />
                )}
                <span className="text-base font-bold text-white">
                  {your_score.objective_achieved
                    ? "Objective Achieved"
                    : "Objective Compromised"}
                </span>
              </div>
              <p className="mt-2 text-xs text-slate-400">
                {your_score.objective_achieved
                  ? "Met core scenario conditions and secured negotiation agreement."
                  : "Target outcome was partially conceded during negotiation."}
              </p>
            </div>

            {/* Elo Rating Delta */}
            <div className="p-4 rounded-xl bg-slate-800/60 border border-slate-700 flex flex-col justify-between">
              <span className="text-xs font-medium text-slate-400 uppercase tracking-wider">
                Elo Rating Update
              </span>
              <div className="mt-2 flex items-baseline gap-2">
                <div
                  className={`flex items-center gap-1 text-2xl font-extrabold ${
                    isEloPositive ? "text-emerald-400" : "text-rose-400"
                  }`}
                >
                  {isEloPositive ? (
                    <TrendingUp className="w-5 h-5" />
                  ) : (
                    <TrendingDown className="w-5 h-5" />
                  )}
                  <span>
                    {isEloPositive ? `+${your_score.elo_delta.toFixed(1)}` : your_score.elo_delta.toFixed(1)}
                  </span>
                </div>
              </div>
              <p className="mt-2 text-xs text-slate-400">
                New Rating:{" "}
                <span className="font-semibold text-white">
                  {your_score.elo_new ? your_score.elo_new.toFixed(0) : "1200"}
                </span>
              </p>
            </div>
          </div>

          {/* Radar Chart & Dimension Scores */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 bg-slate-800/40 p-5 rounded-2xl border border-slate-700/80">
            {/* Visual Radar */}
            <div className="flex flex-col items-center justify-center">
              <h3 className="text-sm font-semibold text-slate-300 self-start mb-2 flex items-center gap-2">
                <Sparkles className="w-4 h-4 text-indigo-400" />
                Rubric Radar
              </h3>
              <div className="w-full h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <RadarChart cx="50%" cy="50%" outerRadius="75%" data={chartData}>
                    <PolarGrid stroke="#334155" />
                    <PolarAngleAxis dataKey="dimension" stroke="#94a3b8" tick={{ fontSize: 11 }} />
                    <PolarRadiusAxis angle={30} domain={[0, 10]} stroke="#475569" />
                    <Radar
                      name="Score"
                      dataKey="score"
                      stroke="#818cf8"
                      fill="#6366f1"
                      fillOpacity={0.4}
                    />
                  </RadarChart>
                </ResponsiveContainer>
              </div>
            </div>

            {/* Dimension Breakdown list */}
            <div className="space-y-3 flex flex-col justify-center">
              <h3 className="text-sm font-semibold text-slate-300 flex items-center gap-2">
                <Layers className="w-4 h-4 text-indigo-400" />
                Dimensional Breakdown (0 - 10)
              </h3>
              <div className="space-y-2.5">
                {(your_score.rubric_scores || []).map((r) => (
                  <div key={r.dimension} className="space-y-1">
                    <div className="flex justify-between text-xs font-medium">
                      <span className="text-slate-300">
                        {DIMENSION_LABELS[r.dimension] || r.dimension}
                      </span>
                      <span className="font-bold text-white">{r.score.toFixed(1)} / 10</span>
                    </div>
                    <div className="w-full bg-slate-700/50 rounded-full h-1.5 overflow-hidden">
                      <div
                        className="bg-indigo-500 h-full rounded-full"
                        style={{ width: `${(r.score / 10) * 100}%` }}
                      />
                    </div>
                    {r.justification && (
                      <p className="text-[11px] text-slate-400 line-clamp-1 italic">
                        &ldquo;{r.justification}&rdquo;
                      </p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Coaching Narrative Feedback */}
          {your_score.summary_feedback && (
            <div className="p-4 rounded-xl bg-slate-800/40 border border-slate-700/80 space-y-2">
              <h3 className="text-sm font-semibold text-indigo-300 flex items-center gap-2">
                <Trophy className="w-4 h-4 text-amber-400" />
                Coach Summary & Executive Feedback
              </h3>
              <p className="text-xs leading-relaxed text-slate-300 whitespace-pre-line">
                {your_score.summary_feedback}
              </p>
            </div>
          )}

          {/* Strengths & Areas for Improvement */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/* Strengths */}
            <div className="p-4 rounded-xl bg-emerald-950/20 border border-emerald-800/40 space-y-2">
              <h4 className="text-xs font-bold text-emerald-400 uppercase tracking-wider flex items-center gap-1.5">
                <CheckCircle2 className="w-4 h-4" />
                Key Strengths
              </h4>
              <ul className="space-y-1.5">
                {(your_score.strengths || []).map((s, idx) => (
                  <li key={idx} className="text-xs text-slate-300 flex items-start gap-2">
                    <span className="text-emerald-400 font-bold">•</span>
                    <span>{s}</span>
                  </li>
                ))}
              </ul>
            </div>

            {/* Areas for Growth */}
            <div className="p-4 rounded-xl bg-amber-950/20 border border-amber-800/40 space-y-2">
              <h4 className="text-xs font-bold text-amber-400 uppercase tracking-wider flex items-center gap-1.5">
                <TrendingUp className="w-4 h-4" />
                Areas for Growth
              </h4>
              <ul className="space-y-1.5">
                {(your_score.areas_for_improvement || []).map((a, idx) => (
                  <li key={idx} className="text-xs text-slate-300 flex items-start gap-2">
                    <span className="text-amber-400 font-bold">•</span>
                    <span>{a}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>

        {/* Footer Actions */}
        <div className="p-5 border-t border-slate-800 bg-slate-900 flex items-center justify-between sticky bottom-0">
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-xl text-xs font-semibold text-slate-300 hover:text-white bg-slate-800 hover:bg-slate-700 transition-colors"
          >
            Back to Dashboard
          </button>
          <Link
            href="/leaderboard"
            className="px-5 py-2 rounded-xl text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white transition-all shadow-md flex items-center gap-2"
          >
            <span>View Standings</span>
            <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>
      </div>
    </div>
  );
}