"use client";

import "@mui/x-date-pickers/themeAugmentation";
import { createTheme } from "@mui/material/styles";

export const moneyManagerTheme = createTheme({
  cssVariables: { colorSchemeSelector: "class" },
  colorSchemes: {
    light: {
      palette: {
        primary: { main: "#7C3AED", light: "#A78BFA", dark: "#6D28D9" },
        success: { main: "#22C55E" },
        error: { main: "#EF4444" },
        background: { default: "#F4F4F5", paper: "#FFFFFF" },
      },
    },
    dark: {
      palette: {
        primary: { main: "#A78BFA", light: "#C4B5FD", dark: "#7C3AED" },
        background: { default: "#0F0F12", paper: "#18181B" },
      },
    },
  },
  typography: {
    fontFamily: "var(--font-sans), Inter, system-ui, sans-serif",
  },
  shape: { borderRadius: 12 },
  components: {
    MuiPickerDay: {
      styleOverrides: {
        root: {
          "&.Mui-selected": {
            backgroundColor: "var(--mui-palette-primary-main)",
            "&:hover": {
              backgroundColor: "var(--mui-palette-primary-dark)",
            },
          },
        },
      },
    },
  },
});

/** Radiant gradient for FAB / accents */
export const radiantGradient = "linear-gradient(145deg, #A78BFA 0%, #7C3AED 50%, #4F46E5 100%)";
