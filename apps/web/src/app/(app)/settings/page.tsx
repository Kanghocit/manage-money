"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useColorScheme } from "@mui/material/styles";
import { Button, Card, CardContent, MenuItem, Stack, TextField, Typography } from "@mui/material";
import { apiFetch } from "@/lib/api-client";
import { useMe } from "@/hooks/use-auth";

type Settings = { theme: string; language: string };

export default function SettingsPage() {
  const { data: user } = useMe();
  const { setMode } = useColorScheme();
  const qc = useQueryClient();
  const { data: settings } = useQuery({
    queryKey: ["settings"],
    queryFn: () => apiFetch<Settings>("/settings"),
  });
  const patchSettings = useMutation({
    mutationFn: (body: Partial<Settings>) =>
      apiFetch("/settings", { method: "PATCH", body: JSON.stringify(body) }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["settings"] }),
  });
  const patchProfile = useMutation({
    mutationFn: (displayName: string) =>
      apiFetch("/settings/profile", {
        method: "PATCH",
        body: JSON.stringify({ displayName }),
      }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["auth", "me"] }),
  });

  return (
    <Stack spacing={2} sx={{ maxWidth: 520 }}>
      <Typography variant="h5" sx={{ fontWeight: 700 }}>
        Cài đặt
      </Typography>
      <Card>
        <CardContent>
          <Stack spacing={2}>
            <TextField label="Họ tên" defaultValue={user?.displayName} onBlur={(e) => patchProfile.mutate(e.target.value)} />
            <TextField label="Email" value={user?.email ?? ""} disabled />
            <TextField
              select
              label="Giao diện"
              value={settings?.theme ?? "system"}
              onChange={(e) => {
                const v = e.target.value;
                setMode(v === "system" ? "system" : v === "dark" ? "dark" : "light");
                patchSettings.mutate({ theme: v });
              }}
            >
              <MenuItem value="light">Sáng</MenuItem>
              <MenuItem value="dark">Tối</MenuItem>
              <MenuItem value="system">Hệ thống</MenuItem>
            </TextField>
            <TextField label="Ngôn ngữ" value="Tiếng Việt" disabled />
            <Button variant="outlined" onClick={() => patchProfile.mutate(user?.displayName ?? "")}>
              Lưu hồ sơ
            </Button>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}
