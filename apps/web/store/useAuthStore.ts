import { create } from "zustand";
import Cookies from "js-cookie";

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
      const token = Cookies.get("auth_token");
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split(".")[1]));
          set({ user: { id: payload.sub || "", email: payload.email || "" }, isAuthenticated: true });
          return;
        } catch (e) {
          // ignore
        }
      }
      set({ user: null, isAuthenticated: true });
    }
  },

  logout: () => {
    set({ user: null, isAuthenticated: false });
  },

  initialize: () => {
    const token = Cookies.get("auth_token");
    if (token) {
      try {
        const payload = JSON.parse(atob(token.split(".")[1]));
        set({ isAuthenticated: true, user: { id: payload.sub || "", email: payload.email || "" } });
      } catch (e) {
        set({ isAuthenticated: true });
      }
    }
  },
}));
