"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  AppBar,
  Avatar,
  Box,
  Drawer,
  IconButton,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import HomeIcon from "@mui/icons-material/Home";
import ReceiptLongIcon from "@mui/icons-material/ReceiptLong";
import PhotoCameraIcon from "@mui/icons-material/PhotoCamera";
import CalendarMonthIcon from "@mui/icons-material/CalendarMonth";
import SavingsIcon from "@mui/icons-material/Savings";
import FlagIcon from "@mui/icons-material/Flag";
import BarChartIcon from "@mui/icons-material/BarChart";
import SettingsIcon from "@mui/icons-material/Settings";
import MenuIcon from "@mui/icons-material/Menu";
import LogoutIcon from "@mui/icons-material/Logout";
import Brightness4Icon from "@mui/icons-material/Brightness4";
import { useColorScheme } from "@mui/material/styles";
import { useState } from "react";
import { useMe, useLogout } from "@/hooks/use-auth";
import { NavIcon } from "@/components/layout/nav-icon";
import { IphoneBottomDock, MOBILE_DOCK_PADDING } from "@/components/layout/iphone-bottom-dock";

const nav = [
  { href: "/overview", label: "Tổng quan", icon: HomeIcon },
  { href: "/transactions", label: "Giao dịch", icon: ReceiptLongIcon },
  { href: "/expenses/daily", label: "Chụp chi tiêu", icon: PhotoCameraIcon },
  { href: "/expenses/calendar", label: "Lịch chi tiêu", icon: CalendarMonthIcon },
  { href: "/budget", label: "Ngân sách", icon: SavingsIcon },
  { href: "/goals", label: "Mục tiêu", icon: FlagIcon },
  { href: "/statistics", label: "Thống kê", icon: BarChartIcon },
  { href: "/settings", label: "Cài đặt", icon: SettingsIcon },
];

const drawerWidth = 260;

export function MuiAppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const theme = useTheme();
  const desktop = useMediaQuery(theme.breakpoints.up("md"));
  const [mobileOpen, setMobileOpen] = useState(false);
  const { data: user } = useMe();
  const logout = useLogout();
  const { mode, setMode } = useColorScheme();

  const drawer = (
    <Box sx={{ py: 2 }}>
      <Typography variant="subtitle2" sx={{ px: 2, color: "text.secondary" }}>
        Money Manager
      </Typography>
      <Typography variant="h6" sx={{ px: 2, mb: 2 }}>
        {user?.displayName}
      </Typography>
      <List>
        {nav.map(({ href, label, icon: Icon }) => {
          const selected = pathname === href || pathname.startsWith(href + "/");
          return (
            <ListItemButton
              key={href}
              component={Link}
              href={href}
              selected={selected}
              onClick={() => setMobileOpen(false)}
              sx={{ gap: 1.5, py: 1.25 }}
            >
              <ListItemIcon sx={{ minWidth: 0 }}>
                <NavIcon icon={Icon} active={selected} size={36} />
              </ListItemIcon>
              <ListItemText primary={label} />
            </ListItemButton>
          );
        })}
      </List>
    </Box>
  );

  return (
    <Box sx={{ display: "flex", minHeight: "100vh", bgcolor: "background.default" }}>
      {desktop ? (
        <Drawer
          variant="permanent"
          sx={{
            width: drawerWidth,
            [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: "border-box" },
          }}
        >
          {drawer}
        </Drawer>
      ) : (
        <Drawer open={mobileOpen} onClose={() => setMobileOpen(false)}>
          {drawer}
        </Drawer>
      )}
      <Box sx={{ flex: 1, display: "flex", flexDirection: "column", pb: { xs: MOBILE_DOCK_PADDING, md: 0 } }}>
        <AppBar position="sticky" color="inherit" elevation={0} sx={{ borderBottom: 1, borderColor: "divider" }}>
          <Toolbar>
            {!desktop && (
              <IconButton edge="start" onClick={() => setMobileOpen(true)}>
                <MenuIcon />
              </IconButton>
            )}
            <Typography variant="h6" sx={{ flex: 1 }}>
              {nav.find((n) => pathname.startsWith(n.href))?.label ?? "Money Manager"}
            </Typography>
            <IconButton onClick={() => setMode(mode === "dark" ? "light" : "dark")} aria-label="theme">
              <Brightness4Icon />
            </IconButton>
            <Avatar sx={{ width: 32, height: 32, ml: 1 }}>{(user?.displayName ?? "U").slice(0, 1)}</Avatar>
            <IconButton onClick={() => logout.mutate()} aria-label="logout">
              <LogoutIcon />
            </IconButton>
          </Toolbar>
        </AppBar>
        <Box component="main" sx={{ flex: 1, p: { xs: 2, md: 3 } }}>
          {children}
        </Box>
      </Box>
      {!desktop && <IphoneBottomDock />}
    </Box>
  );
}
