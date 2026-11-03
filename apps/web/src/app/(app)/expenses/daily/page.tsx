"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs, { Dayjs } from "dayjs";
import { useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Paper,
  Stack,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from "@mui/material";
import LoadingButton from "@mui/lab/LoadingButton";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { apiFetch } from "@/lib/api-client";
import {
  uploadExpenseImage,
  type ExpenseDayEntry,
  type ExpenseDayLine,
} from "@/lib/expense-days-api";
import { formatVND } from "@/lib/format";

export default function ExpenseDailyPage() {
  const qc = useQueryClient();
  const [date, setDate] = useState<Dayjs>(dayjs());
  const [entryId, setEntryId] = useState<string | null>(null);
  const [lines, setLines] = useState<ExpenseDayLine[]>([]);

  const dateStr = date.format("YYYY-MM-DD");

  const ensureEntry = useMutation({
    mutationFn: () =>
      apiFetch<ExpenseDayEntry>("/expense-days", {
        method: "POST",
        body: JSON.stringify({ date: dateStr }),
      }),
    onSuccess: (e) => setEntryId(e.id),
  });

  useEffect(() => {
    ensureEntry.mutate();
  }, [dateStr]);

  const { data: entry, isFetching } = useQuery({
    queryKey: ["expense-day", entryId],
    enabled: !!entryId,
    queryFn: () => apiFetch<ExpenseDayEntry>(`/expense-days/${entryId}`),
    refetchInterval: (q) => {
      const st = q.state.data?.status;
      return st === "processing" ? 2000 : false;
    },
  });

  useEffect(() => {
    if (entry?.lines) setLines(entry.lines);
  }, [entry?.lines]);

  const upload = useMutation({
    mutationFn: (file: File) => uploadExpenseImage(entryId!, file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["expense-day", entryId] }),
  });

  const saveLines = useMutation({
    mutationFn: () =>
      apiFetch<ExpenseDayEntry>(`/expense-days/${entryId}/lines`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ lines }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["expense-day", entryId] }),
  });

  const confirm = useMutation({
    mutationFn: () =>
      apiFetch<ExpenseDayEntry>(`/expense-days/${entryId}/confirm`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({}),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["expense-day", entryId] }),
  });

  return (
      <Stack spacing={3} sx={{ maxWidth: 900, mx: "auto" }}>
        <Typography variant="h5" sx={{ fontWeight: 700 }}>
          Chi tiêu theo ngày
        </Typography>
        <DatePicker label="Ngày" value={date} onChange={(v) => v && setDate(v)} slotProps={{ textField: { fullWidth: true } }} />
        <Paper sx={{ p: 2 }}>
          <Stack direction={{ xs: "column", sm: "row" }} spacing={2} sx={{ alignItems: "center" }}>
            <Button variant="outlined" component="label">
              Chụp / chọn ảnh
              <input
                hidden
                type="file"
                accept="image/*"
                capture="environment"
                onChange={(e) => {
                  const f = e.target.files?.[0];
                  if (f && entryId) upload.mutate(f);
                }}
              />
            </Button>
            {upload.isPending && <CircularProgress size={24} />}
            {entry?.status && (
              <Typography variant="body2" color="text.secondary">
                Trạng thái: {entry.status}
                {isFetching && entry.status === "processing" ? " (đang OCR…)" : ""}
              </Typography>
            )}
          </Stack>
          {entry?.statusMessage && entry.status === "failed" && (
            <Alert severity="error" sx={{ mt: 2 }}>
              {entry.statusMessage}
            </Alert>
          )}
          {entry?.imageUrl && (
            <Box component="img" src={entry.imageUrl} alt="receipt" sx={{ mt: 2, maxHeight: 280, borderRadius: 2 }} />
          )}
        </Paper>
        {(entry?.status === "ready" || entry?.status === "confirmed") && (
          <Paper sx={{ p: 2 }}>
            <Typography variant="h6" gutterBottom>
              Danh sách mục chi
            </Typography>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Tên</TableCell>
                  <TableCell>Số tiền</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {lines.map((line, i) => (
                  <TableRow key={i}>
                    <TableCell>
                      <TextField
                        size="small"
                        fullWidth
                        value={line.name}
                        onChange={(e) => {
                          const next = [...lines];
                          next[i] = { ...line, name: e.target.value };
                          setLines(next);
                        }}
                      />
                    </TableCell>
                    <TableCell>
                      <TextField
                        size="small"
                        value={line.amount}
                        onChange={(e) => {
                          const next = [...lines];
                          next[i] = { ...line, amount: e.target.value };
                          setLines(next);
                        }}
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
            <Stack direction="row" spacing={2} sx={{ mt: 2 }}>
              <LoadingButton loading={saveLines.isPending} variant="outlined" onClick={() => saveLines.mutate()}>
                Lưu danh sách
              </LoadingButton>
              <LoadingButton
                loading={confirm.isPending}
                variant="contained"
                disabled={entry?.status === "confirmed"}
                onClick={() => confirm.mutate()}
              >
                Xác nhận → tạo giao dịch
              </LoadingButton>
            </Stack>
            <Typography variant="body2" sx={{ mt: 1 }}>
              Tổng: {formatVND(entry?.totalAmount ?? 0)}
            </Typography>
          </Paper>
        )}
      </Stack>
  );
}
