"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Box, Fab, Typography } from "@mui/material";
import HomeIcon from "@mui/icons-material/Home";
import ReceiptLongIcon from "@mui/icons-material/ReceiptLong";
import PhotoCameraIcon from "@mui/icons-material/PhotoCamera";
import CalendarMonthIcon from "@mui/icons-material/CalendarMonth";
import SettingsIcon from "@mui/icons-material/Settings";
import { NavIcon } from "@/components/layout/nav-icon";
import { radiantGradient } from "@/theme/money-manager.theme";

const leftTabs = [
  { href: "/overview", label: "Home", icon: HomeIcon },
  { href: "/transactions", label: "GD", icon: ReceiptLongIcon },
] as const;

const rightTabs = [
  { href: "/expenses/calendar", label: "Lịch", icon: CalendarMonthIcon },
  { href: "/settings", label: "Cài đặt", icon: SettingsIcon },
] as const;

function DockTab({ href, label, icon: Icon, active }: { href: string; label: string; icon: typeof HomeIcon; active: boolean }) {
  return (
    <Box
      component={Link}
      href={href}
      sx={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        gap: 0.5,
        minHeight: 56,
        minWidth: 64,
        py: 1,
        px: 0.5,
        textDecoration: "none",
        color: active ? "primary.main" : "text.primary",
        WebkitTapHighlightColor: "transparent",
      }}
    >
      <NavIcon icon={Icon} active={active} size={44} variant="dock" />
      <Typography variant="caption" sx={{ fontSize: "0.7rem", lineHeight: 1.2, fontWeight: active ? 600 : 400 }}>
        {label}
      </Typography>
    </Box>
  );
}

export function IphoneBottomDock() {
  const pathname = usePathname();
  const isActive = (href: string) => pathname === href || pathname.startsWith(`${href}/`);
  const fabActive = isActive("/expenses/daily");

  return (
    <Box
      sx={{
        position: "fixed",
        bottom: 0,
        left: 0,
        right: 0,
        zIndex: 1200,
        pointerEvents: "none",
      }}
    >
      <Box
        component={Link}
        href="/expenses/daily"
        sx={{
          pointerEvents: "auto",
          position: "absolute",
          left: "50%",
          transform: "translateX(-50%)",
          bottom: "calc(62px + env(safe-area-inset-bottom, 0px))",
          zIndex: 1202,
          textDecoration: "none",
          p: 1,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        <Fab
          aria-label="Chụp chi tiêu"
          sx={{
            width: 56,
            height: 56,
            background: radiantGradient,
            color: "#fff",
            outline: (t) => `3px solid ${t.palette.background.paper}`,
            boxShadow: fabActive
              ? "0 8px 28px rgba(124, 58, 237, 0.55)"
              : (t) =>
                  t.palette.mode === "dark"
                    ? "0 6px 24px rgba(0, 0, 0, 0.5), 0 4px 16px rgba(124, 58, 237, 0.4)"
                    : "0 6px 20px rgba(124, 58, 237, 0.4)",
            "&:hover": {
              background: radiantGradient,
              filter: "brightness(1.05)",
            },
          }}
        >
          <PhotoCameraIcon />
        </Fab>
      </Box>
      <Box
        sx={{
          pointerEvents: "auto",
          display: "flex",
          alignItems: "flex-end",
          minHeight: 80,
          pt: 1.5,
          pb: "calc(10px + env(safe-area-inset-bottom, 0px))",
          px: 0.5,
          borderTopLeftRadius: 24,
          borderTopRightRadius: 24,
          bgcolor: (t) =>
            t.palette.mode === "dark" ? "rgba(24, 24, 27, 0.96)" : "rgba(255, 255, 255, 0.94)",
          backdropFilter: "blur(16px)",
          borderTop: 1,
          borderColor: "divider",
          boxShadow: (t) =>
            t.palette.mode === "dark" ? "0 -8px 32px rgba(0, 0, 0, 0.45)" : "0 -4px 24px rgba(0, 0, 0, 0.08)",
        }}
      >
        <Box sx={{ flex: 1, display: "flex", justifyContent: "space-around", pr: 3 }}>
          {leftTabs.map((t) => (
            <DockTab key={t.href} {...t} active={isActive(t.href)} />
          ))}
        </Box>
        <Box sx={{ width: 76, flexShrink: 0 }} />
        <Box sx={{ flex: 1, display: "flex", justifyContent: "space-around", pl: 3 }}>
          {rightTabs.map((t) => (
            <DockTab key={t.href} {...t} active={isActive(t.href)} />
          ))}
        </Box>
      </Box>
    </Box>
  );
}

/** Reserve space above fixed dock + FAB */
export const MOBILE_DOCK_PADDING = "calc(120px + env(safe-area-inset-bottom, 0px))";
