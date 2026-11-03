"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useState } from "react";
import dayjs, { Dayjs } from "dayjs";
import { Alert, Box, Button, Card, CardContent, Stack, TextField, ToggleButton, ToggleButtonGroup, Typography } from "@mui/material";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { apiFetch } from "@/lib/api-client";
import { hasPositiveAmount } from "@/lib/money-input";
import { MoneyInput } from "@/components/money-input";

type Category = { id: string; name: string; type: string };

export default function NewTransactionPage() {
  const router = useRouter();
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

  const create = useMutation({
    mutationFn: () =>
      apiFetch("/transactions", {
        method: "POST",
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
      router.push("/transactions");
    },
  });

  const filtered = (categories ?? []).filter((c) => c.type === type);

  return (
    <Card sx={{ maxWidth: 480, mx: "auto" }}>
      <CardContent>
        <Typography variant="h6" gutterBottom>
          Thêm giao dịch
        </Typography>
        <Stack spacing={2}>
          <ToggleButtonGroup exclusive value={type} onChange={(_, v) => v && setType(v)} fullWidth>
            <ToggleButton value="expense">Chi tiêu</ToggleButton>
            <ToggleButton value="income">Thu nhập</ToggleButton>
          </ToggleButtonGroup>
          <MoneyInput label="Số tiền" fullWidth value={amount} onValueChange={setAmount} />
          <DatePicker
            label="Ngày giao dịch"
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
          {create.isError && <Alert severity="error">{create.error.message}</Alert>}
          <Button
            variant="contained"
            disabled={!hasPositiveAmount(amount) || !categoryId || create.isPending}
            onClick={() => create.mutate()}
          >
            Lưu
          </Button>
        </Stack>
      </CardContent>
    </Card>
  );
}
