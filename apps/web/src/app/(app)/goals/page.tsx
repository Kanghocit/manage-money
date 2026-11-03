"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Box, Button, Card, CardContent, Grid, LinearProgress, Stack, TextField, Typography } from "@mui/material";
import { apiFetch } from "@/lib/api-client";
import { formatVND } from "@/lib/format";
import { MoneyInput } from "@/components/money-input";
import { hasPositiveAmount } from "@/lib/money-input";
import { useState } from "react";

type Goal = { id: string; name: string; targetAmount: string; currentAmount: string };

export default function GoalsPage() {
  const qc = useQueryClient();
  const [name, setName] = useState("");
  const [target, setTarget] = useState("");
  const { data: goals } = useQuery({
    queryKey: ["goals"],
    queryFn: () => apiFetch<Goal[]>("/goals"),
  });
  const create = useMutation({
    mutationFn: () =>
      apiFetch("/goals", {
        method: "POST",
        body: JSON.stringify({ name, targetAmount: target, currentAmount: "0" }),
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["goals"] });
      setName("");
      setTarget("");
    },
  });

  return (
    <Stack spacing={3}>
      <Typography variant="h5" sx={{ fontWeight: 700 }}>
        Mục tiêu
      </Typography>
      <Grid container spacing={2}>
        {(goals ?? []).map((g) => {
          const pct = Math.min(100, Math.round((Number(g.currentAmount) / Number(g.targetAmount)) * 100) || 0);
          return (
            <Grid size={{ xs: 12, md: 6 }} key={g.id}>
              <Card>
                <Box sx={{ height: 80, background: "linear-gradient(90deg,#7C3AED,#4F46E5)" }} />
                <CardContent>
                  <Typography variant="h6">{g.name}</Typography>
                  <LinearProgress variant="determinate" value={pct} sx={{ my: 1 }} />
                  <Typography variant="body2">
                    {formatVND(g.currentAmount)} / {formatVND(g.targetAmount)} ({pct}%)
                  </Typography>
                </CardContent>
              </Card>
            </Grid>
          );
        })}
      </Grid>
      <Card sx={{ maxWidth: 480 }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Thêm mục tiêu
          </Typography>
          <Stack spacing={2}>
            <TextField label="Tên" value={name} onChange={(e) => setName(e.target.value)} />
            <MoneyInput label="Số tiền mục tiêu" fullWidth value={target} onValueChange={setTarget} />
            <Button variant="contained" disabled={!name || !hasPositiveAmount(target)} onClick={() => create.mutate()}>
              Tạo
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}
