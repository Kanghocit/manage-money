"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import {
  Alert,
  Button,
  IconButton,
  List,
  ListItem,
  ListItemText,
  Stack,
  TextField,
  ToggleButton,
  ToggleButtonGroup,
  Typography,
} from "@mui/material";
import DeleteOutlineOutlinedIcon from "@mui/icons-material/DeleteOutlineOutlined";
import { apiFetch } from "@/lib/api-client";

type Category = {
  id: string;
  name: string;
  type: string;
  isSystem?: boolean;
  color?: string;
};

export function CategoryManager() {
  const qc = useQueryClient();
  const [name, setName] = useState("");
  const [type, setType] = useState<"expense" | "income">("expense");

  const { data: categories } = useQuery({
    queryKey: ["categories"],
    queryFn: () => apiFetch<Category[]>("/categories"),
  });

  const create = useMutation({
    mutationFn: () =>
      apiFetch<Category>("/categories", {
        method: "POST",
        body: JSON.stringify({ name: name.trim(), type, color: "#7C3AED" }),
      }),
    onSuccess: () => {
      setName("");
      qc.invalidateQueries({ queryKey: ["categories"] });
    },
  });

  const remove = useMutation({
    mutationFn: (id: string) => apiFetch(`/categories/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["categories"] }),
  });

  const expense = (categories ?? []).filter((c) => c.type === "expense");
  const income = (categories ?? []).filter((c) => c.type === "income");

  const renderGroup = (title: string, items: Category[]) => (
    <>
      <Typography variant="subtitle2" color="text.secondary" sx={{ mt: 1 }}>
        {title}
      </Typography>
      <List dense disablePadding>
        {items.map((c) => (
          <ListItem
            key={c.id}
            secondaryAction={
              c.isSystem ? (
                <Typography variant="caption" color="text.secondary">
                  Hệ thống
                </Typography>
              ) : (
                <IconButton
                  edge="end"
                  aria-label="Xóa danh mục"
                  size="small"
                  onClick={() => {
                    if (window.confirm(`Xóa danh mục "${c.name}"?`)) remove.mutate(c.id);
                  }}
                  disabled={remove.isPending}
                >
                  <DeleteOutlineOutlinedIcon fontSize="small" />
                </IconButton>
              )
            }
          >
            <ListItemText primary={c.name} />
          </ListItem>
        ))}
      </List>
    </>
  );

  return (
    <Stack spacing={2}>
      <Typography variant="subtitle1" sx={{ fontWeight: 600 }}>
        Danh mục thu / chi
      </Typography>
      <Typography variant="body2" color="text.secondary">
        Thêm danh mục riêng để chọn khi tạo giao dịch hoặc xác nhận chi tiêu ngày.
      </Typography>
      <ToggleButtonGroup exclusive size="small" value={type} onChange={(_, v) => v && setType(v)}>
        <ToggleButton value="expense">Chi tiêu</ToggleButton>
        <ToggleButton value="income">Thu nhập</ToggleButton>
      </ToggleButtonGroup>
      <Stack direction="row" spacing={1}>
        <TextField
          label="Tên danh mục mới"
          size="small"
          fullWidth
          value={name}
          onChange={(e) => setName(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && name.trim()) create.mutate();
          }}
        />
        <Button variant="contained" disabled={!name.trim() || create.isPending} onClick={() => create.mutate()}>
          Thêm
        </Button>
      </Stack>
      {create.isError && <Alert severity="error">{create.error.message}</Alert>}
      {remove.isError && <Alert severity="error">{remove.error.message}</Alert>}
      {renderGroup("Chi tiêu", expense)}
      {renderGroup("Thu nhập", income)}
    </Stack>
  );
}
