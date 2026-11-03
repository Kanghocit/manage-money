"use client";

import { useQuery } from "@tanstack/react-query";
import dayjs, { Dayjs } from "dayjs";
import { useMemo, useState } from "react";
import {
  Box,
  Button,
  Drawer,
  List,
  ListItem,
  ListItemText,
  Paper,
  Stack,
  Typography,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import { DateCalendar } from "@mui/x-date-pickers/DateCalendar";
import Link from "next/link";
import { apiFetch } from "@/lib/api-client";
import type { ExpenseDayEntry } from "@/lib/expense-days-api";
import { expenseEntryHasData, expenseEntryTotal, parseLineAmount } from "@/lib/expense-day-totals";
import { formatVND } from "@/lib/format";
import { ExpenseCalendarDay } from "@/components/expenses/expense-calendar-day";

function monthRange(view: Dayjs) {
  const start = view.startOf("month").format("YYYY-MM-DD");
  const end = view.endOf("month").format("YYYY-MM-DD");
  return { start, end };
}

function EntryDetail({
  entry,
  onClose,
}: {
  entry: ExpenseDayEntry;
  onClose: () => void;
}) {
  const total = expenseEntryTotal(entry);
  return (
    <Stack spacing={2} sx={{ p: 2, pb: 3 }}>
      <Box sx={{ width: 40, height: 4, bgcolor: "divider", borderRadius: 2, mx: "auto" }} />
      <Typography variant="h6">{entry.entryDate} — chi tiêu</Typography>
      {entry.imageUrl && (
        <Box component="img" src={entry.imageUrl} alt="" sx={{ width: "100%", borderRadius: 2, maxHeight: 280, objectFit: "cover" }} />
      )}
      <Typography gutterBottom>
        Tổng: <strong>{formatVND(total)}</strong> · {entry.status}
      </Typography>
      <List dense disablePadding>
        {(entry.lines ?? []).map((l, i) => (
          <ListItem key={l.id ?? i} disableGutters>
            <ListItemText primary={l.name} secondary={formatVND(parseLineAmount(l.amount))} />
          </ListItem>
        ))}
      </List>
      <Button component={Link} href="/expenses/daily" variant="contained" onClick={onClose}>
        Chỉnh sửa ngày này
      </Button>
    </Stack>
  );
}

export default function ExpenseCalendarPage() {
  const theme = useTheme();
  const mobile = useMediaQuery(theme.breakpoints.down("md"));
  /** Tháng đang xem — tách khỏi ngày được chọn để đổi tháng không nhảy focus ngày 1. */
  const [calendarMonth, setCalendarMonth] = useState(dayjs());
  const [selectedEntryId, setSelectedEntryId] = useState<string | null>(null);
  const { start, end } = monthRange(calendarMonth);

  const { data: entries } = useQuery({
    queryKey: ["expense-days", start, end],
    queryFn: () => apiFetch<ExpenseDayEntry[]>(`/expense-days?from=${start}&to=${end}`),
  });

  const byDate = useMemo(() => {
    const map = new Map<string, ExpenseDayEntry>();
    (entries ?? []).forEach((e) => map.set(e.entryDate, e));
    return map;
  }, [entries]);

  const { data: selectedEntry } = useQuery({
    queryKey: ["expense-day", selectedEntryId],
    enabled: !!selectedEntryId,
    queryFn: () => apiFetch<ExpenseDayEntry>(`/expense-days/${selectedEntryId}`),
  });

  const withData = (entries ?? []).filter((e) => expenseEntryHasData(e));

  return (
    <Stack spacing={2} sx={{ maxWidth: 720, mx: "auto" }}>
      <Typography variant="h5" sx={{ fontWeight: 700 }}>
        Lịch chi tiêu tháng
      </Typography>
      <Paper sx={{ p: { xs: 1, sm: 2 }, borderRadius: 3 }}>
        <DateCalendar
          referenceDate={calendarMonth}
          onMonthChange={(m) => setCalendarMonth(m)}
          onChange={(d) => {
            if (!d) return;
            const entry = byDate.get(d.format("YYYY-MM-DD"));
            if (entry) setSelectedEntryId(entry.id);
          }}
          slots={{ day: ExpenseCalendarDay }}
          slotProps={{
            day: { expenseEntries: byDate } as never,
          }}
        />
      </Paper>
      {withData.length > 0 && (
        <Typography variant="body2" color="text.secondary">
          {withData.length} ngày có chi tiêu trong tháng — chạm ô trên lịch để xem chi tiết (ảnh hoặc tổng tiền).
        </Typography>
      )}
      <Typography variant="caption" color="text.secondary">
        Lịch chỉ hiển thị chi tiêu theo ngày (ảnh/OCR). Giao dịch thêm thủ công xem tại mục Giao dịch.
      </Typography>
      <Drawer
        anchor="bottom"
        open={!!selectedEntryId && !!selectedEntry}
        onClose={() => setSelectedEntryId(null)}
        slotProps={{
          paper: {
            sx: {
              borderTopLeftRadius: 16,
              borderTopRightRadius: 16,
              maxHeight: mobile ? "85vh" : "70vh",
              maxWidth: mobile ? "100%" : 480,
              mx: mobile ? 0 : "auto",
              left: mobile ? 0 : "50%",
              right: mobile ? 0 : "auto",
              transform: mobile ? undefined : "translateX(-50%)",
            },
          },
        }}
      >
        {selectedEntry && (
          <EntryDetail entry={selectedEntry} onClose={() => setSelectedEntryId(null)} />
        )}
      </Drawer>
    </Stack>
  );
}
