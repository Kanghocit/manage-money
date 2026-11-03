import { create } from "zustand";

type UIState = {
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  dashboardRange: "7d" | "30d" | "60d";
  setDashboardRange: (r: "7d" | "30d" | "60d") => void;
};

export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: false,
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  dashboardRange: "30d",
  setDashboardRange: (dashboardRange) => set({ dashboardRange }),
}));
