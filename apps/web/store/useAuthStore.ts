import { create } from "zustand";
import { DefaultBrowserStorage } from "@ubi/sdk";

const storage = new DefaultBrowserStorage();

interface User {
  id: string;
  email: string;
  [key: string]: any;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  setAuth: (user?: User) => void;
  logout: () => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,

  setAuth: (user?: User) => {
    if (user) {
      set({ user, isAuthenticated: true });
    } else {
      const token = storage.get("auth_token");
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split(".")[1]));
          const id = payload.sub || payload.user_id || "";
          const email = payload.email || "";
          set({ user: { id, email }, isAuthenticated: true });
          return;
        } catch (e) {
          // ignore
        }
      }
      set({ user: null, isAuthenticated: false });
    }
  },

  logout: () => {
    storage.remove("auth_token");
    set({ user: null, isAuthenticated: false });
  },

  initialize: () => {
    const token = storage.get("auth_token");
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split(".")[1]));
        const id = payload.sub || payload.user_id || "";
        const email = payload.email || "";
        set({ isAuthenticated: true, user: { id, email } });
      } catch (e) {
        set({ isAuthenticated: false, user: null });
      }
    } else {
      set({ isAuthenticated: false, user: null });
    }
  },
}));
