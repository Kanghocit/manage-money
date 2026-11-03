"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import dayjs, { Dayjs } from "dayjs";
import { Box, Button, Card, CardContent, Stack, TextField, Typography } from "@mui/material";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { apiFetch } from "@/lib/api-client";
import { formatVND } from "@/lib/format";
import { useMemo, useState } from "react";

type Tx = {
  id: string;
  amount: string;
  type: string;
  note: string;
  occurredAt: string;
  category?: { name: string };
};

export default function TransactionsPage() {
  const [q, setQ] = useState("");
  const [monthValue, setMonthValue] = useState<Dayjs | null>(null);
  const month = useMemo(() => (monthValue ? monthValue.format("YYYY-MM") : ""), [monthValue]);

  const { data } = useQuery({
    queryKey: ["transactions", q, month],
    queryFn: () => {
      const params = new URLSearchParams();
      if (q) params.set("q", q);
      if (month) params.set("month", month);
      return apiFetch<{ items: Tx[]; total: number }>(`/transactions?${params}`);
    },
  });

  return (
    <Stack spacing={2}>
      <Stack direction="row" sx={{ justifyContent: "space-between", alignItems: "center" }}>
        <Typography variant="h5" sx={{ fontWeight: 700 }}>
          Giao dịch
        </Typography>
        <Button component={Link} href="/transactions/new" variant="contained">
          Thêm
        </Button>
      </Stack>
      <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
        <TextField label="Tìm kiếm" value={q} onChange={(e) => setQ(e.target.value)} fullWidth />
        <DatePicker
          label="Tháng"
          views={["year", "month"]}
          openTo="month"
          value={monthValue}
          onChange={setMonthValue}
          referenceDate={dayjs()}
          slotProps={{ textField: { fullWidth: true } }}
        />
      </Stack>
      <Card>
        <CardContent>
          <Typography variant="subtitle1" gutterBottom>
            Danh sách ({data?.total ?? 0})
          </Typography>
          {(data?.items ?? []).map((tx) => (
            <Box key={tx.id} sx={{ display: "flex", justifyContent: "space-between", py: 1.5, borderBottom: 1, borderColor: "divider" }}>
              <Box>
                <Typography sx={{ fontWeight: 600 }}>{tx.category?.name}</Typography>
                <Typography variant="body2" color="text.secondary">
                  {tx.note || new Date(tx.occurredAt).toLocaleString("vi-VN")}
                </Typography>
              </Box>
              <Typography color={tx.type === "income" ? "success.main" : "error.main"}>
                {tx.type === "income" ? "+" : "-"}
                {formatVND(tx.amount)}
              </Typography>
            </Box>
          ))}
        </CardContent>
      </Card>
    </Stack>
  );
}
