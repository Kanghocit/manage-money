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
import { formatVND } from "@/lib/format";
import { ExpenseCalendarDay, type DailySpend } from "@/components/expenses/expense-calendar-day";

function monthRange(view: Dayjs) {
  const start = view.startOf("month").format("YYYY-MM-DD");
  const end = view.endOf("month").format("YYYY-MM-DD");
  return { start, end };
}

type TxRow = {
  id: string;
  amount: string;
  note: string;
  type: string;
  category?: { name: string };
};

function DayDetail({
  date,
  imageUrl,
  transactions,
  total,
  onClose,
}: {
  date: string;
  imageUrl?: string;
  transactions: TxRow[];
  total: number;
  onClose: () => void;
}) {
  return (
    <Stack spacing={2} sx={{ p: 2, pb: 3 }}>
      <Box sx={{ width: 40, height: 4, bgcolor: "divider", borderRadius: 2, mx: "auto" }} />
      <Typography variant="h6">{date} — chi tiêu</Typography>
      {imageUrl && (
        <Box component="img" src={imageUrl} alt="" sx={{ width: "100%", borderRadius: 2, maxHeight: 280, objectFit: "cover" }} />
      )}
      <Typography gutterBottom>
        Tổng: <strong>{formatVND(total)}</strong> · {transactions.length} giao dịch
      </Typography>
      <List dense disablePadding>
        {transactions.map((tx) => (
          <ListItem key={tx.id} disableGutters>
            <ListItemText
              primary={tx.note || tx.category?.name || "Chi tiêu"}
              secondary={`${tx.category?.name ?? ""} · ${formatVND(tx.amount)}`}
            />
          </ListItem>
        ))}
      </List>
      <Button component={Link} href="/expenses/daily" variant="contained" onClick={onClose}>
        Chụp / chỉnh sửa ngày này
      </Button>
    </Stack>
  );
}

export default function ExpenseCalendarPage() {
  const theme = useTheme();
  const mobile = useMediaQuery(theme.breakpoints.down("md"));
  const [calendarMonth, setCalendarMonth] = useState(dayjs());
  const [selectedDate, setSelectedDate] = useState<string | null>(null);
  const { start, end } = monthRange(calendarMonth);

  const { data: entries } = useQuery({
    queryKey: ["expense-days", start, end],
    queryFn: () => apiFetch<ExpenseDayEntry[]>(`/expense-days?from=${start}&to=${end}`),
  });

  const { data: dailyTotals } = useQuery({
    queryKey: ["transactions-daily", start, end],
    queryFn: () => apiFetch<DailySpend[]>(`/transactions/daily?from=${start}&to=${end}`),
  });

  const byDate = useMemo(() => {
    const map = new Map<string, ExpenseDayEntry>();
    (entries ?? []).forEach((e) => map.set(e.entryDate, e));
    return map;
  }, [entries]);

  const totalsByDate = useMemo(() => {
    const map = new Map<string, DailySpend>();
    (dailyTotals ?? []).forEach((d) => map.set(d.date, d));
    return map;
  }, [dailyTotals]);

  const { data: dayTx } = useQuery({
    queryKey: ["transactions", "day", selectedDate],
    enabled: !!selectedDate,
    queryFn: () =>
      apiFetch<{ items: TxRow[] }>(
        `/transactions?date=${selectedDate}&type=expense&pageSize=100`,
      ),
  });

  const daysWithSpend = (dailyTotals ?? []).filter((d) => Number(d.total) > 0).length;

  const openDay = (dateKey: string) => {
    const spend = Number(totalsByDate.get(dateKey)?.total ?? 0);
    const hasImage = !!byDate.get(dateKey)?.imageUrl;
    if (spend > 0 || hasImage) setSelectedDate(dateKey);
  };

  const selectedTotal = selectedDate
    ? Number(totalsByDate.get(selectedDate)?.total ?? 0)
    : 0;

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
            openDay(d.format("YYYY-MM-DD"));
          }}
          slots={{ day: ExpenseCalendarDay }}
          slotProps={{
            day: { expenseEntries: byDate, dailyTotals: totalsByDate } as never,
          }}
        />
      </Paper>
      {daysWithSpend > 0 && (
        <Typography variant="body2" color="text.secondary">
          {daysWithSpend} ngày có chi tiêu (giao dịch) trong tháng — chạm ô để xem chi tiết.
        </Typography>
      )}
      <Drawer
        anchor="bottom"
        open={!!selectedDate}
        onClose={() => setSelectedDate(null)}
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
        {selectedDate && (
          <DayDetail
            date={selectedDate}
            imageUrl={byDate.get(selectedDate)?.imageUrl}
            transactions={dayTx?.items ?? []}
            total={selectedTotal}
            onClose={() => setSelectedDate(null)}
          />
        )}
      </Drawer>
    </Stack>
  );
}
