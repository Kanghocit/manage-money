"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Box, CircularProgress } from "@mui/material";
import { useMe } from "@/hooks/use-auth";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { data, isLoading, isError } = useMe();
  const router = useRouter();

  useEffect(() => {
    if (!isLoading && (isError || !data)) {
      router.replace("/login");
    }
  }, [isLoading, isError, data, router]);

  if (isLoading || !data) {
    return (
      <Box sx={{ display: "flex", minHeight: "100vh", alignItems: "center", justifyContent: "center" }}>
        <CircularProgress />
      </Box>
    );
  }

  return <>{children}</>;
}
