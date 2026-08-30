import { create } from "zustand";
import {
  RoomState,
  RoomReadyData,
  SpinResultData,
  ScenarioAssignData,
  EvaluationData,
  WSEnvelope,
} from "../lib/types";

interface RoomStoreState {
  roomId: string | null;
  state: RoomState;
  seat: "A" | "B" | null;
  peer: {
    displayName: string;
    avatarUrl: string;
    eloRating: number;
  } | null;
  spinSeed: number | null;
  spinResult: SpinResultData | null;
  scenario: ScenarioAssignData | null;
  prepSecondsRemaining: number;
  sessionSecondsRemaining: number;
  isPeerReady: boolean;
  isMeReady: boolean;
  evaluation: EvaluationData | null;

  setRoomId: (id: string) => void;
  setRoomState: (state: RoomState) => void;
  setMeReady: () => void;
  handleWSEvent: (env: WSEnvelope) => void;
  reset: () => void;
}

export const useRoomStore = create<RoomStoreState>((set, get) => ({
  roomId: null,
  state: "idle",
  seat: null,
  peer: null,
  spinSeed: null,
  spinResult: null,
  scenario: null,
  prepSecondsRemaining: 180,
  sessionSecondsRemaining: 360,
  isPeerReady: false,
  isMeReady: false,
  evaluation: null,

  setRoomId: (id) => set({ roomId: id }),
  setRoomState: (state) => set({ state }),
  setMeReady: () => set({ isMeReady: true }),

  handleWSEvent: (env: WSEnvelope) => {
    switch (env.event) {
      case "ROOM_READY": {
        const p: RoomReadyData = env.payload;
        set({
          roomId: p.room_id,
          state: "ready",
          seat: p.seat,
          peer: {
            displayName: p.peer_display_name,
            avatarUrl: p.peer_avatar_url,
            eloRating: p.peer_elo_rating,
          },
        });
        break;
      }

      case "SPIN_START": {
        set({
          state: "spinning",
          spinSeed: env.payload.spin_seed,
        });
        break;
      }

      case "SPIN_RESULT": {
        const p: SpinResultData = env.payload;
        set({
          state: "scenario",
          spinResult: p,
        });
        break;
      }

      case "SCENARIO_ASSIGN": {
        const p: ScenarioAssignData = env.payload;
        set({
          scenario: p,
          prepSecondsRemaining: p.prep_duration_seconds,
          sessionSecondsRemaining: p.session_duration_seconds,
          state: "prep",
        });
        break;
      }

      case "PREP_TIMER_TICK": {
        set({
          prepSecondsRemaining: env.payload.seconds_remaining,
          isPeerReady: env.payload.peer_ready,
        });
        break;
      }

      case "PREP_END": {
        set({ state: "signaling" });
        break;
      }

      case "SESSION_TIMER_TICK": {
        set({
          state: "live",
          sessionSecondsRemaining: env.payload.seconds_remaining,
        });
        break;
      }

      case "SESSION_COMPLETE": {
        set({ state: "evaluating" });
        break;
      }

      case "EVALUATION_READY": {
        set({
          state: "complete",
          evaluation: env.payload,
        });
        break;
      }

      case "ROOM_CLOSED": {
        set({ state: "closed" });
        break;
      }
    }
  },

  reset: () =>
    set({
      roomId: null,
      state: "idle",
      seat: null,
      peer: null,
      spinSeed: null,
      spinResult: null,
      scenario: null,
      prepSecondsRemaining: 180,
      sessionSecondsRemaining: 360,
      isPeerReady: false,
      isMeReady: false,
      evaluation: null,
    }),
}));