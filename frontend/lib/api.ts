import { AuthResponse, User } from "./types";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/v1";

class ApiClient {
  private getToken(): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem("access_token");
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = this.getToken();
    const headers: HeadersInit = {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    };

    if (token) {
      (headers as Record<string, string>)["Authorization"] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}${endpoint}`, {
      ...options,
      headers,
    });

    if (!res.ok) {
      let errData: any;
      try {
        errData = await res.json();
      } catch {
        errData = { message: res.statusText };
      }
      throw new Error(errData?.error?.message || errData?.message || `HTTP ${res.status}`);
    }

    if (res.status === 204) {
      return {} as T;
    }

    return res.json();
  }

  async register(data: { username: string; email: string; password: string; display_name: string }): Promise<AuthResponse> {
    return this.request<AuthResponse>("/auth/register", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async login(data: { email: string; password: string }): Promise<AuthResponse> {
    return this.request<AuthResponse>("/auth/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async getMe(): Promise<User> {
    return this.request<User>("/users/me");
  }

  async updateMe(data: { display_name?: string; avatar_url?: string }): Promise<User> {
    return this.request<User>("/users/me", {
      method: "PATCH",
      body: JSON.stringify(data),
    });
  }

  async getLeaderboard(page = 1, limit = 50) {
    return this.request<{
      data: any[];
      my_rank: number;
      my_elo: number;
      pagination: { page: number; limit: number; total: number; total_pages: number };
    }>(`/leaderboard?page=${page}&limit=${limit}`);
  }

  async enqueueMatchmaking(data: { preferred_difficulty?: string; preferred_context?: string }) {
    return this.request<{ status: string; position: number; estimated_wait_seconds: number }>("/matchmaking/enqueue", {
      method: "POST",
      body: JSON.stringify(data),
    });
  }

  async dequeueMatchmaking() {
    return this.request<void>("/matchmaking/dequeue", {
      method: "DELETE",
    });
  }
}

export const api = new ApiClient();