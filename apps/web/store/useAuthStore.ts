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
  setAuth: (token: string, user?: User) => void;
  logout: () => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,

  setAuth: (token: string, user?: User) => {
    // Save to cookie (expires in 7 days)
    Cookies.set("auth_token", token, { expires: 7, path: "/" });
    set({ user: user || null, isAuthenticated: true });
  },

  logout: () => {
    Cookies.remove("auth_token", { path: "/" });
    set({ user: null, isAuthenticated: false });
  },

  initialize: () => {
    const token = Cookies.get("auth_token");
    if (token) {
      set({ isAuthenticated: true });
    }
  },
}));
