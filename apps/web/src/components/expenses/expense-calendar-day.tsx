"use client";

import { PickerDay, PickerDayProps } from "@mui/x-date-pickers/PickerDay";
import { Avatar, Box, Typography } from "@mui/material";
import type { ExpenseDayEntry } from "@/lib/expense-days-api";
import { expenseEntryHasData, expenseEntryTotal } from "@/lib/expense-day-totals";
import { formatShortVND } from "@/lib/format";

export type ExpenseCalendarDayProps = PickerDayProps & {
  expenseEntries?: Map<string, ExpenseDayEntry>;
};

export function ExpenseCalendarDay(props: ExpenseCalendarDayProps) {
  const { expenseEntries, day, outsideCurrentMonth, today, selected, ...other } = props;
  const dateKey = day.format("YYYY-MM-DD");
  const entry = expenseEntries?.get(dateKey);
  const imageUrl = entry?.imageUrl;
  const showExpense = !!entry && expenseEntryHasData(entry) && !outsideCurrentMonth;
  const dayNum = day.date();

  const ring =
    selected || today
      ? {
          boxShadow: selected
            ? "0 0 0 2px #fff, 0 0 0 4px #7C3AED"
            : "0 0 0 2px rgba(124, 58, 237, 0.45)",
        }
      : {};

  if (imageUrl && showExpense) {
    return (
      <Box sx={{ position: "relative", display: "flex", justifyContent: "center", alignItems: "center" }}>
        <PickerDay
          {...other}
          day={day}
          outsideCurrentMonth={outsideCurrentMonth}
          today={today}
          selected={selected}
          sx={{
            p: 0,
            width: 44,
            height: 44,
            minWidth: 44,
            borderRadius: "50%",
            overflow: "hidden",
            fontSize: 0,
            ...ring,
          }}
        >
          <Avatar src={imageUrl} alt="" sx={{ width: 40, height: 40 }} />
        </PickerDay>
      </Box>
    );
  }

  if (showExpense && entry) {
    const lineCount = entry.lines?.length ?? 0;
    const totalLabel = formatShortVND(expenseEntryTotal(entry)).replace(" đ", "");
    return (
      <Box sx={{ display: "flex", justifyContent: "center", alignItems: "center" }}>
        <PickerDay
          {...other}
          day={day}
          outsideCurrentMonth={outsideCurrentMonth}
          today={today}
          selected={selected}
          sx={{
            p: 0,
            width: 44,
            height: 52,
            minWidth: 44,
            borderRadius: 2,
            flexDirection: "column",
            gap: 0,
            bgcolor: (t) => (t.palette.mode === "dark" ? "rgba(124, 58, 237, 0.2)" : "rgba(124, 58, 237, 0.12)"),
            border: 1,
            borderColor: "primary.main",
            ...ring,
          }}
        >
          <Typography variant="caption" sx={{ fontSize: "0.65rem", fontWeight: 700, lineHeight: 1.1 }}>
            {dayNum}
          </Typography>
          <Typography variant="caption" sx={{ fontSize: "0.55rem", fontWeight: 600, color: "primary.main", lineHeight: 1.1 }}>
            {totalLabel}
          </Typography>
          {lineCount > 0 && (
            <Typography variant="caption" sx={{ fontSize: "0.5rem", color: "text.secondary", lineHeight: 1 }}>
              ({lineCount})
            </Typography>
          )}
        </PickerDay>
      </Box>
    );
  }

  return (
    <PickerDay
      {...other}
      day={day}
      outsideCurrentMonth={outsideCurrentMonth}
      today={today}
      selected={selected}
      sx={{
        ...(today && !selected
          ? {
              border: "1px solid",
              borderColor: "primary.main",
            }
          : {}),
        ...ring,
      }}
    />
  );
}
