"use client";

import Link from "next/link";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import dayjs, { Dayjs } from "dayjs";
import { useMemo, useState } from "react";
import {
  Box,
  Button,
  Card,
  CardContent,
  FormControl,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";
import DeleteOutlineOutlinedIcon from "@mui/icons-material/DeleteOutlineOutlined";
import EditOutlinedIcon from "@mui/icons-material/EditOutlined";
import { DatePicker } from "@mui/x-date-pickers/DatePicker";
import { apiFetch } from "@/lib/api-client";
import { formatVND } from "@/lib/format";
import { TransactionEditDialog, type TxEdit } from "@/components/transactions/transaction-edit-dialog";

type Category = { id: string; name: string; type: string };

type Tx = TxEdit & {
  category?: { name: string; id?: string };
};

export default function TransactionsPage() {
  const qc = useQueryClient();
  const [q, setQ] = useState("");
  const [monthValue, setMonthValue] = useState<Dayjs | null>(null);
  const [typeFilter, setTypeFilter] = useState<"" | "expense" | "income">("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const [editTx, setEditTx] = useState<Tx | null>(null);

  const month = useMemo(() => (monthValue ? monthValue.format("YYYY-MM") : ""), [monthValue]);

  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: () => apiFetch<Category[]>("/categories"),
  });

  const categoryOptions = (categories ?? []).filter((c) => !typeFilter || c.type === typeFilter);

  const { data } = useQuery({
    queryKey: ["transactions", q, month, typeFilter, categoryFilter],
    queryFn: () => {
      const params = new URLSearchParams();
      if (q) params.set("q", q);
      if (month) params.set("month", month);
      if (typeFilter) params.set("type", typeFilter);
      if (categoryFilter) params.set("category", categoryFilter);
      return apiFetch<{ items: Tx[]; total: number }>(`/transactions?${params}`);
    },
  });

  const remove = useMutation({
    mutationFn: (id: string) => apiFetch(`/transactions/${id}`, { method: "DELETE" }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["transactions"] });
      qc.invalidateQueries({ queryKey: ["dashboard"] });
    },
  });

  const handleDelete = (tx: Tx) => {
    if (window.confirm(`Xóa giao dịch ${formatVND(tx.amount)}?`)) {
      remove.mutate(tx.id);
    }
  };

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
      <Stack spacing={2}>
        <ToggleButtonGroup
          exclusive
          value={typeFilter}
          onChange={(_, v) => {
            setTypeFilter(v ?? "");
            setCategoryFilter("");
          }}
          size="small"
        >
          <ToggleButton value="">Tất cả</ToggleButton>
          <ToggleButton value="expense">Chi</ToggleButton>
          <ToggleButton value="income">Thu</ToggleButton>
        </ToggleButtonGroup>
        <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
          <TextField label="Tìm kiếm" value={q} onChange={(e) => setQ(e.target.value)} fullWidth />
          <FormControl fullWidth size="small">
            <InputLabel>Danh mục</InputLabel>
            <Select
              label="Danh mục"
              value={categoryFilter}
              onChange={(e) => setCategoryFilter(e.target.value)}
            >
              <MenuItem value="">Tất cả</MenuItem>
              {categoryOptions.map((c) => (
                <MenuItem key={c.id} value={c.id}>
                  {c.name}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <DatePicker
            label="Tháng"
            views={["year", "month"]}
            openTo="month"
            value={monthValue}
            onChange={setMonthValue}
            referenceDate={dayjs()}
            slotProps={{ textField: { fullWidth: true, size: "small" } }}
          />
        </Stack>
      </Stack>
      <Card>
        <CardContent>
          <Typography variant="subtitle1" gutterBottom>
            Danh sách ({data?.total ?? 0})
          </Typography>
          {(data?.items ?? []).map((tx) => (
            <Box
              key={tx.id}
              sx={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", py: 1.5, borderBottom: 1, borderColor: "divider", gap: 1 }}
            >
              <Box sx={{ flex: 1, minWidth: 0 }}>
                <Typography sx={{ fontWeight: 600 }}>{tx.category?.name}</Typography>
                <Typography variant="body2" color="text.secondary">
                  {tx.note || new Date(tx.occurredAt).toLocaleString("vi-VN")}
                </Typography>
              </Box>
              <Typography color={tx.type === "income" ? "success.main" : "error.main"} sx={{ whiteSpace: "nowrap" }}>
                {tx.type === "income" ? "+" : "-"}
                {formatVND(tx.amount)}
              </Typography>
              <Stack direction="row" sx={{ flexShrink: 0 }}>
                <IconButton size="small" aria-label="Sửa" onClick={() => setEditTx(tx)}>
                  <EditOutlinedIcon fontSize="small" />
                </IconButton>
                <IconButton size="small" aria-label="Xóa" onClick={() => handleDelete(tx)} disabled={remove.isPending}>
                  <DeleteOutlineOutlinedIcon fontSize="small" />
                </IconButton>
              </Stack>
            </Box>
          ))}
        </CardContent>
      </Card>
      <TransactionEditDialog tx={editTx} open={!!editTx} onClose={() => setEditTx(null)} />
    </Stack>
  );
}
