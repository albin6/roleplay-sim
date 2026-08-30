export type UserRole = "player" | "admin";

export interface User {
  id: string;
  username: string;
  email: string;
  display_name: string;
  avatar_url?: string | null;
  elo_rating: number;
  total_sessions: number;
  wins: number;
  losses: number;
  role: UserRole;
  created_at: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  expires_in: number;
}

export type RoomState =
  | "idle"
  | "waiting"
  | "ready"
  | "spinning"
  | "scenario"
  | "prep"
  | "signaling"
  | "live"
  | "evaluating"
  | "complete"
  | "closed";

export interface RoleInfo {
  id: string;
  name: string;
  hierarchy_level: number;
  description: string;
}

export interface ContextInfo {
  id: string;
  name: string;
  slug: string;
}

export interface ScenarioAssignData {
  room_id: string;
  scenario_id: string;
  title: string;
  difficulty: string;
  background_context: string;
  your_objective: string;
  your_constraints: string[];
  prep_duration_seconds: number;
  session_duration_seconds: number;
}

export interface SpinResultData {
  room_id: string;
  context: ContextInfo;
  your_role: RoleInfo;
  peer_role: RoleInfo;
  difficulty: string;
}

export interface RoomReadyData {
  room_id: string;
  peer_display_name: string;
  peer_avatar_url: string;
  peer_elo_rating: number;
  seat: "A" | "B";
}

export interface RubricScore {
  dimension: string;
  score: number;
  weight: number;
  justification: string;
}

export interface EvaluationData {
  room_id: string;
  session_id: string;
  your_score: {
    overall_score: number;
    objective_achieved: boolean;
    elo_delta: number;
    elo_new: number;
    summary_feedback: string;
    strengths: string[];
    areas_for_improvement: string[];
    rubric_scores: RubricScore[];
  };
  peer_score: {
    overall_score: number;
    objective_achieved: boolean;
    elo_delta: number;
  };
}

export interface WSEnvelope<T = any> {
  event: string;
  payload: T;
  timestamp: string;
  seq: number;
}