"use client";

import { Box, SvgIconProps } from "@mui/material";
import { radiantGradient } from "@/theme/money-manager.theme";

type NavIconProps = {
  icon: React.ComponentType<SvgIconProps>;
  active?: boolean;
  size?: number;
  /** Sidebar: gradient when active. Dock: lighter active state for dark mode. */
  variant?: "sidebar" | "dock";
};

export function NavIcon({ icon: Icon, active = false, size = 40, variant = "sidebar" }: NavIconProps) {
  const dockActive = variant === "dock" && active;
  const sidebarActive = variant === "sidebar" && active;

  return (
    <Box
      sx={{
        width: size,
        height: size,
        borderRadius: "50%",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        flexShrink: 0,
        background: sidebarActive
          ? radiantGradient
          : dockActive
            ? (t) => t.palette.primary.main
            : (t) =>
                t.palette.mode === "dark"
                  ? "rgba(124, 58, 237, 0.15)"
                  : t.palette.action.hover,
        border: dockActive || sidebarActive ? "none" : (t) => `1px solid ${t.palette.divider}`,
        boxShadow: sidebarActive
          ? "0 4px 12px rgba(124, 58, 237, 0.45)"
          : dockActive
            ? "0 2px 8px rgba(124, 58, 237, 0.35)"
            : "none",
        color: sidebarActive || dockActive ? "#fff" : "text.primary",
        opacity: active ? 1 : 0.88,
        transition: "transform 0.15s ease, box-shadow 0.15s ease",
        ...(active && { transform: "translateY(-1px)" }),
      }}
    >
      <Icon sx={{ fontSize: size * 0.48 }} />
    </Box>
  );
}
