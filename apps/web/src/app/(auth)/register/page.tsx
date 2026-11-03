"use client";

import Link from "next/link";
import { useState } from "react";
import { Alert, Box, Button, Paper, Stack, TextField, Typography } from "@mui/material";
import { useRegister } from "@/hooks/use-auth";

export default function RegisterPage() {
  const register = useRegister();
  const [displayName, setDisplayName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  return (
    <Paper sx={{ p: 4, width: "100%", maxWidth: 420 }}>
      <Typography variant="h5" sx={{ fontWeight: 700 }} gutterBottom>
        Đăng ký
      </Typography>
      <Stack spacing={2}>
        <TextField label="Họ tên" value={displayName} onChange={(e) => setDisplayName(e.target.value)} fullWidth />
        <TextField label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} fullWidth />
        <TextField label="Mật khẩu" type="password" value={password} onChange={(e) => setPassword(e.target.value)} fullWidth />
        {register.isError && <Alert severity="error">{register.error.message}</Alert>}
        <Button variant="contained" size="large" disabled={register.isPending} onClick={() => register.mutate({ email, password, displayName })}>
          Đăng ký
        </Button>
        <Box sx={{ textAlign: "center" }}>
          <Typography variant="body2" component={Link} href="/login" color="primary">
            Đã có tài khoản? Đăng nhập
          </Typography>
        </Box>
      </Stack>
    </Paper>
  );
}
