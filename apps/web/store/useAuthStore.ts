import { create } from "zustand";
import Cookies from "js-cookie";

interface User {
  id: string;
  email: string;
  [key: string]: any;
}

interface AuthState {
  user: User | null;
  setAuth: (user?: User) => void;
  initialize: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,

  setAuth: (user?: User) => {
    set({ user: user || null, isAuthenticated: true });
  },

  logout: () => {
    set({ user: null, isAuthenticated: false });
  },

  initialize: () => {
    const token = Cookies.get("auth_token");
    if (token) {
      set({ isAuthenticated: true });
    }
  },
}));
