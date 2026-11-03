"use client";

import Link from "next/link";
import { useState } from "react";
import { Alert, Box, Button, Paper, Stack, TextField, Typography } from "@mui/material";
import { useLogin } from "@/hooks/use-auth";

export default function LoginPage() {
  const login = useLogin();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  return (
    <Paper sx={{ p: 4, width: "100%", maxWidth: 420 }}>
      <Typography variant="h5" sx={{ fontWeight: 700 }} gutterBottom>
        Đăng nhập
      </Typography>
      <Stack spacing={2} component="form">
        <TextField label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} fullWidth />
        <TextField label="Mật khẩu" type="password" value={password} onChange={(e) => setPassword(e.target.value)} fullWidth />
        {login.isError && <Alert severity="error">{login.error.message}</Alert>}
        <Button variant="contained" size="large" disabled={login.isPending} onClick={() => login.mutate({ email, password })}>
          Đăng nhập
        </Button>
        <Box sx={{ textAlign: "center" }}>
          <Typography variant="body2" component={Link} href="/register" color="primary">
            Chưa có tài khoản? Đăng ký
          </Typography>
        </Box>
      </Stack>
    </Paper>
  );
}
