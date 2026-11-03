"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Box, Button, Card, CardContent, LinearProgress, Stack, TextField, Typography } from "@mui/material";
import { apiFetch } from "@/lib/api-client";
import { formatVND } from "@/lib/format";
import { MoneyInput } from "@/components/money-input";
import { digitsOnly } from "@/lib/money-input";
import { useState } from "react";

type Category = { id: string; name: string; type: string };
type BudgetLine = { id: string; categoryId: string; limitAmount: string; spentAmount: string; category?: Category };

export default function BudgetPage() {
  const qc = useQueryClient();
  const { data } = useQuery({
    queryKey: ["budget"],
    queryFn: () =>
      apiFetch<{
        budget: { lines: BudgetLine[] };
        usedPercent: number;
        totalLimit: string;
        totalSpent: string;
      }>("/budgets/current"),
  });
  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: () => apiFetch<Category[]>("/categories"),
  });
  const [limits, setLimits] = useState<Record<string, string>>({});
  const expenseCats = (categories ?? []).filter((c) => c.type === "expense");

  const save = useMutation({
    mutationFn: () =>
      apiFetch("/budgets/current", {
        method: "PUT",
        body: JSON.stringify({
          lines: expenseCats.map((c) => ({
            categoryId: c.id,
            limitAmount: digitsOnly(limits[c.id] ?? "") || "0",
          })),
        }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["budget"] }),
  });

  const used = data?.usedPercent ?? 0;

  return (
    <Stack spacing={3} sx={{ maxWidth: 640, mx: "auto" }}>
      <Typography variant="h5" sx={{ fontWeight: 700 }}>
        Ngân sách
      </Typography>
      <Card sx={{ textAlign: "center", py: 4 }}>
        <Typography variant="h3" color="primary">
          {used}%
        </Typography>
        <Typography color="text.secondary">
          {formatVND(data?.totalSpent ?? 0)} / {formatVND(data?.totalLimit ?? 0)}
        </Typography>
      </Card>
      <Card>
        <CardContent>
          {(data?.budget?.lines ?? []).map((line) => {
            const pct =
              line.limitAmount === "0"
                ? 0
                : Math.min(100, Math.round((Number(line.spentAmount) / Number(line.limitAmount)) * 100));
            return (
              <Box key={line.id} sx={{ mb: 2 }}>
                <Typography variant="body2">
                  {line.category?.name} — {formatVND(line.spentAmount)} / {formatVND(line.limitAmount)}
                </Typography>
                <LinearProgress variant="determinate" value={pct} />
              </Box>
            );
          })}
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Thiết lập hạn mức
          </Typography>
          {expenseCats.map((c) => (
            <Stack direction="row" spacing={2} sx={{ alignItems: "center", mb: 1 }} key={c.id}>
              <Typography sx={{ width: 100 }}>{c.name}</Typography>
              <MoneyInput
                size="small"
                sx={{ flex: 1 }}
                value={limits[c.id] ?? ""}
                onValueChange={(digits) => setLimits((s) => ({ ...s, [c.id]: digits }))}
              />
            </Stack>
          ))}
          <Button variant="contained" onClick={() => save.mutate()} disabled={save.isPending}>
            Lưu
          </Button>
        </CardContent>
      </Card>
    </Stack>
  );
}
