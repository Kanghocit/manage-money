"use client";

import { PickerDay, PickerDayProps } from "@mui/x-date-pickers/PickerDay";
import { Avatar, Box } from "@mui/material";
import type { ExpenseDayEntry } from "@/lib/expense-days-api";

export type ExpenseCalendarDayProps = PickerDayProps & {
  expenseEntries?: Map<string, ExpenseDayEntry>;
};

export function ExpenseCalendarDay(props: ExpenseCalendarDayProps) {
  const { expenseEntries, day, outsideCurrentMonth, today, selected, ...other } = props;
  const dateKey = day.format("YYYY-MM-DD");
  const entry = expenseEntries?.get(dateKey);
  const imageUrl = entry?.imageUrl;

  const ring =
    selected || today
      ? {
          boxShadow: selected
            ? "0 0 0 2px #fff, 0 0 0 4px #7C3AED"
            : "0 0 0 2px rgba(124, 58, 237, 0.45)",
        }
      : {};

  if (imageUrl && !outsideCurrentMonth) {
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
            width: 40,
            height: 40,
            minWidth: 40,
            borderRadius: "50%",
            overflow: "hidden",
            fontSize: 0,
            ...ring,
          }}
        >
          <Avatar src={imageUrl} alt="" sx={{ width: 36, height: 36 }} />
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
