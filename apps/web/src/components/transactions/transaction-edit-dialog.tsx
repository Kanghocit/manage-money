"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs, { Dayjs } from "dayjs";
import { useEffect, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
} from "@mui/material";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { apiFetch } from "@/lib/api-client";
import { hasPositiveAmount } from "@/lib/money-input";
import { MoneyInput } from "@/components/money-input";

type Category = { id: string; name: string; type: string };

export type TxEdit = {
  id: string;
  amount: string;
  type: string;
  note: string;
  occurredAt: string;
  categoryId?: string;
  category?: { id: string; name: string };
};

export function TransactionEditDialog({
  tx,
  open,
  onClose,
}: {
  tx: TxEdit | null;
  open: boolean;
  onClose: () => void;
}) {
  const qc = useQueryClient();
  const [type, setType] = useState<"expense" | "income">("expense");
  const [amount, setAmount] = useState("");
  const [categoryId, setCategoryId] = useState("");
  const [note, setNote] = useState("");
  const [occurredAt, setOccurredAt] = useState<Dayjs>(dayjs());

  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: () => apiFetch<Category[]>("/categories"),
  });

  useEffect(() => {
    if (!tx || !open) return;
    setType(tx.type as "expense" | "income");
    setAmount(String(tx.amount).replace(/\D/g, ""));
    setCategoryId(tx.categoryId ?? tx.category?.id ?? "");
    setNote(tx.note ?? "");
    setOccurredAt(dayjs(tx.occurredAt));
  }, [tx, open]);

  const save = useMutation({
    mutationFn: () =>
      apiFetch(`/transactions/${tx!.id}`, {
        method: "PATCH",
        body: JSON.stringify({
          type,
          amount,
          categoryId,
          note,
          occurredAt: occurredAt.startOf("day").toISOString(),
        }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
      onClose();
    },
  });

  const filtered = (categories ?? []).filter((c) => c.type === type);

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="xs">
      <DialogTitle>Sửa giao dịch</DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <ToggleButtonGroup exclusive value={type} onChange={(_, v) => v && setType(v)} fullWidth>
            <ToggleButton value="expense">Chi tiêu</ToggleButton>
            <ToggleButton value="income">Thu nhập</ToggleButton>
          </ToggleButtonGroup>
          <MoneyInput label="Số tiền" fullWidth value={amount} onValueChange={setAmount} />
          <DatePicker
            label="Ngày"
            value={occurredAt}
            onChange={(d) => d && setOccurredAt(d)}
            slotProps={{ textField: { fullWidth: true } }}
          />
          <Box sx={{ display: "flex", flexWrap: "wrap", gap: 1 }}>
            {filtered.map((c) => (
              <Button key={c.id} variant={categoryId === c.id ? "contained" : "outlined"} size="small" onClick={() => setCategoryId(c.id)}>
                {c.name}
              </Button>
            ))}
          </Box>
          <TextField label="Ghi chú" fullWidth value={note} onChange={(e) => setNote(e.target.value)} />
          {save.isError && <Alert severity="error">{save.error.message}</Alert>}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>Huỷ</Button>
        <Button variant="contained" disabled={!hasPositiveAmount(amount) || !categoryId || save.isPending} onClick={() => save.mutate()}>
          Lưu
        </Button>
      </DialogActions>
    </Dialog>
  );
}
