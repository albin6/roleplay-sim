import { create } from "zustand";
import { User } from "../lib/types";
import { api } from "../lib/api";

interface AuthState {
  user: User | null;
  token: string | null;
  isLoading: boolean;
  error: string | null;
  setAuth: (user: User, token: string) => void;
  logout: () => void;
  loadUser: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: typeof window !== "undefined" ? localStorage.getItem("access_token") : null,
  isLoading: false,
  error: null,

  setAuth: (user, token) => {
    if (typeof window !== "undefined") {
      localStorage.setItem("access_token", token);
    }
    set({ user, token, error: null });
  },

  logout: () => {
    if (typeof window !== "undefined") {
      localStorage.removeItem("access_token");
    }
    set({ user: null, token: null, error: null });
  },

  loadUser: async () => {
    set({ isLoading: true });
    try {
      const user = await api.getMe();
      set({ user, isLoading: false, error: null });
    } catch (err: any) {
      if (typeof window !== "undefined") {
        localStorage.removeItem("access_token");
      }
      set({ user: null, token: null, isLoading: false, error: err.message });
    }
  },
}));