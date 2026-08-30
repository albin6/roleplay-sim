"use client";

import React, { useEffect, useState } from "react";
import { SpinResultData } from "../../lib/types";

interface DualWheelProps {
  isSpinning: boolean;
  spinResult: SpinResultData | null;
}

export function DualWheel({ isSpinning, spinResult }: DualWheelProps) {
  const [rotation1, setRotation1] = useState(0);
  const [rotation2, setRotation2] = useState(0);

  useEffect(() => {
    if (isSpinning) {
      // Rotate several times during the 3.5s animation
      const target1 = 1440 + Math.floor(Math.random() * 360);
      const target2 = 1800 + Math.floor(Math.random() * 360);
      setRotation1(target1);
      setRotation2(target2);
    }
  }, [isSpinning]);

  return (
    <div className="flex flex-col items-center justify-center p-8 bg-slate-900/80 rounded-2xl border border-slate-700 shadow-2xl backdrop-blur-md">
      <h2 className="text-2xl font-bold text-indigo-400 mb-2">Dual Wheel Spin</h2>
      <p className="text-sm text-slate-400 mb-8">
        Synchronized context & role assignment in progress...
      </p>

      <div className="flex flex-col md:flex-row items-center gap-12">
        {/* Wheel 1: Context */}
        <div className="flex flex-col items-center">
          <div className="relative w-48 h-48 rounded-full border-4 border-indigo-500/50 flex items-center justify-center shadow-lg shadow-indigo-500/20 overflow-hidden">
            <div
              className="w-full h-full rounded-full transition-transform duration-[3500ms] cubic-bezier(0.15, 0.9, 0.2, 1) flex items-center justify-center bg-gradient-to-tr from-indigo-900 to-slate-800"
              style={{ transform: `rotate(${rotation1}deg)` }}
            >
              <div className="text-center font-bold text-slate-200">
                {spinResult ? spinResult.context.name : "Context"}
              </div>
            </div>
            {/* Indicator Needle */}
            <div className="absolute -top-2 w-4 h-6 bg-indigo-400 [clip-path:polygon(50%_100%,0%_0%,100%_0%)] z-10" />
          </div>
          <span className="mt-3 text-xs uppercase tracking-wider font-semibold text-slate-400">
            Workplace Context
          </span>
        </div>

        {/* Wheel 2: Role */}
        <div className="flex flex-col items-center">
          <div className="relative w-48 h-48 rounded-full border-4 border-purple-500/50 flex items-center justify-center shadow-lg shadow-purple-500/20 overflow-hidden">
            <div
              className="w-full h-full rounded-full transition-transform duration-[3500ms] cubic-bezier(0.15, 0.9, 0.2, 1) flex items-center justify-center bg-gradient-to-tr from-purple-900 to-slate-800"
              style={{ transform: `rotate(${rotation2}deg)` }}
            >
              <div className="text-center font-bold text-slate-200 px-2">
                {spinResult ? spinResult.your_role.name : "Role"}
              </div>
            </div>
            {/* Indicator Needle */}
            <div className="absolute -top-2 w-4 h-6 bg-purple-400 [clip-path:polygon(50%_100%,0%_0%,100%_0%)] z-10" />
          </div>
          <span className="mt-3 text-xs uppercase tracking-wider font-semibold text-slate-400">
            Your Assigned Role
          </span>
        </div>
      </div>

      {spinResult && (
        <div className="mt-8 p-4 bg-slate-800/90 rounded-xl border border-indigo-500/30 text-center animate-fade-in">
          <span className="text-xs text-indigo-400 uppercase font-semibold">Matched Roles</span>
          <p className="text-lg font-bold text-white mt-1">
            You: <span className="text-indigo-300">{spinResult.your_role.name}</span> vs. Opponent:{" "}
            <span className="text-purple-300">{spinResult.peer_role.name}</span>
          </p>
        </div>
      )}
    </div>
  );
}