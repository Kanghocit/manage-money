"use client";

import { useQuery } from "@tanstack/react-query";
import { Card, CardContent, Typography } from "@mui/material";
import { apiFetch } from "@/lib/api-client";
import { MonthlyBarChart } from "@/components/charts/monthly-bar-chart";
import { SpendingStructureChart } from "@/components/charts/spending-structure-chart";

export default function StatisticsPage() {
  const { data: trends } = useQuery({
    queryKey: ["stats", "trends"],
    queryFn: () => apiFetch<{ month: string; income: string; expense: string }[]>("/stats/trends?months=6"),
  });
  const { data: structure } = useQuery({
    queryKey: ["stats", "structure"],
    queryFn: () => apiFetch<{ name: string; value: string; color: string }[]>("/stats/spending-structure"),
  });

  return (
    <>
      <Typography variant="h5" sx={{ fontWeight: 700 }} gutterBottom>
        Thống kê
      </Typography>
      <Card sx={{ mb: 2 }}>
        <CardContent>
          <MonthlyBarChart data={trends ?? []} />
        </CardContent>
      </Card>
      <Card>
        <CardContent>
          <SpendingStructureChart data={structure ?? []} />
        </CardContent>
      </Card>
    </>
  );
}
