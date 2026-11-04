"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs, { Dayjs } from "dayjs";
import { useEffect, useMemo, useState } from "react";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  IconButton,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import DeleteOutlineOutlinedIcon from "@mui/icons-material/DeleteOutlineOutlined";
import AddIcon from "@mui/icons-material/Add";
import LoadingButton from "@mui/lab/LoadingButton";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { apiFetch } from "@/lib/api-client";
import { hasPositiveAmount } from "@/lib/money-input";
import {
  uploadExpenseImage,
  type ExpenseDayEntry,
  type ExpenseDayLine,
} from "@/lib/expense-days-api";
import { parseLineAmount } from "@/lib/expense-day-totals";
import { formatVND } from "@/lib/format";
import { MoneyInput } from "@/components/money-input";

const emptyLine = (): ExpenseDayLine => ({ name: "", amount: "" });

function sanitizeLines(lines: ExpenseDayLine[]): ExpenseDayLine[] {
  return lines
    .map((l) => ({
      name: l.name.trim(),
      amount: String(l.amount).replace(/\D/g, ""),
    }))
    .filter((l) => l.name && hasPositiveAmount(l.amount));
}

export default function ExpenseDailyPage() {
  const qc = useQueryClient();
  const [date, setDate] = useState<Dayjs>(dayjs());
  const [entryId, setEntryId] = useState<string | null>(null);
  const [lines, setLines] = useState<ExpenseDayLine[]>([emptyLine()]);

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
    if (!entry) return;
    const fromServer = entry.lines ?? [];
    if (fromServer.length === 0) {
      setLines([emptyLine()]);
    } else {
      setLines(
        fromServer.map((l) => ({
          ...l,
          amount: String(l.amount).replace(/\D/g, ""),
        })),
      );
    }
  }, [entry?.id, entry?.updatedAt]);

  const upload = useMutation({
    mutationFn: (file: File) => uploadExpenseImage(entryId!, file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["expense-day", entryId] }),
  });

  const saveLines = useMutation({
    mutationFn: () => {
      const payload = sanitizeLines(lines);
      return apiFetch<ExpenseDayEntry>(`/expense-days/${entryId}/lines`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ lines: payload }),
      });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["expense-day", entryId] });
      qc.invalidateQueries({ queryKey: ["expense-days"] });
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["transactions-daily"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });

  const formTotal = useMemo(
    () => lines.reduce((sum, l) => sum + parseLineAmount(l.amount), 0),
    [lines],
  );

  const canSave = sanitizeLines(lines).length > 0;

  return (
    <Stack spacing={3} sx={{ maxWidth: 520, mx: "auto" }}>
      <Typography variant="h5" sx={{ fontWeight: 700 }}>
        Chụp chi tiêu
      </Typography>
      <DatePicker label="Ngày" value={date} onChange={(v) => v && setDate(v)} slotProps={{ textField: { fullWidth: true } }} />
      <Paper sx={{ p: 2 }}>
        <Stack spacing={2}>
          <Button variant="outlined" component="label" fullWidth>
            Chụp / chọn ảnh (tuỳ chọn)
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
          {upload.isPending && <CircularProgress size={24} sx={{ mx: "auto" }} />}
          {entry?.status === "processing" && (
            <Typography variant="body2" color="text.secondary" sx={{ textAlign: "center" }}>
              Đang OCR… {isFetching ? "" : ""}
            </Typography>
          )}
          {entry?.statusMessage && entry.status === "failed" && (
            <Alert severity="error">{entry.statusMessage}</Alert>
          )}
          {entry?.imageUrl && (
            <Box component="img" src={entry.imageUrl} alt="receipt" sx={{ maxHeight: 280, borderRadius: 2, width: "100%", objectFit: "cover" }} />
          )}
        </Stack>
      </Paper>

      <Paper sx={{ p: 2 }}>
        <Typography variant="h6" gutterBottom>
          Mục chi
        </Typography>
        <Stack spacing={2}>
          {lines.map((line, i) => (
            <Stack key={i} spacing={1} sx={{ border: 1, borderColor: "divider", borderRadius: 2, p: 1.5 }}>
              <TextField
                label="Tên / mô tả"
                fullWidth
                value={line.name}
                onChange={(e) => {
                  const next = [...lines];
                  next[i] = { ...line, name: e.target.value };
                  setLines(next);
                }}
              />
              <Stack direction="row" spacing={1} sx={{ alignItems: "center" }}>
                <MoneyInput
                  label="Số tiền"
                  fullWidth
                  value={String(line.amount).replace(/\D/g, "")}
                  onValueChange={(digits) => {
                    const next = [...lines];
                    next[i] = { ...line, amount: digits };
                    setLines(next);
                  }}
                />
                <IconButton
                  aria-label="Xóa mục"
                  disabled={lines.length <= 1}
                  onClick={() => setLines(lines.filter((_, j) => j !== i))}
                >
                  <DeleteOutlineOutlinedIcon />
                </IconButton>
              </Stack>
            </Stack>
          ))}
          <Button startIcon={<AddIcon />} onClick={() => setLines([...lines, emptyLine()])}>
            Thêm mục
          </Button>
          <Typography variant="body2">
            Tổng: <strong>{formatVND(formTotal)}</strong>
          </Typography>
          {saveLines.isError && <Alert severity="error">{saveLines.error.message}</Alert>}
          <LoadingButton
            loading={saveLines.isPending}
            variant="contained"
            size="large"
            disabled={!canSave}
            onClick={() => saveLines.mutate()}
          >
            Lưu
          </LoadingButton>
          <Typography variant="caption" color="text.secondary">
            Lưu sẽ tạo/cập nhật giao dịch chi tiêu trong ngày đã chọn (cùng nguồn với mục Giao dịch và Lịch).
          </Typography>
        </Stack>
      </Paper>
    </Stack>
  );
}
