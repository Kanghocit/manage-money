"use client";

import { Box, SvgIconProps } from "@mui/material";
import { radiantGradient } from "@/theme/money-manager.theme";

type NavIconProps = {
  icon: React.ComponentType<SvgIconProps>;
  active?: boolean;
  size?: number;
};

export function NavIcon({ icon: Icon, active = false, size = 40 }: NavIconProps) {
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
        background: active ? radiantGradient : (t) => t.palette.action.hover,
        boxShadow: active ? "0 4px 12px rgba(124, 58, 237, 0.45)" : "0 1px 3px rgba(0,0,0,0.08)",
        color: active ? "#fff" : "text.secondary",
        transition: "transform 0.15s ease, box-shadow 0.15s ease",
        ...(active && { transform: "translateY(-1px)" }),
      }}
    >
      <Icon sx={{ fontSize: size * 0.48 }} />
    </Box>
  );
}
