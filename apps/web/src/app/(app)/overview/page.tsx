"use client";

import { useQuery } from "@tanstack/react-query";
import { Box, Card, CardContent, Grid, Tab, Tabs, Typography } from "@mui/material";
import { apiFetch } from "@/lib/api-client";
import { formatShortVND, formatVND } from "@/lib/format";
import { useUIStore } from "@/stores/ui-store";
import { SpendingTrendChart } from "@/components/charts/spending-trend-chart";
import { SpendingStructureChart } from "@/components/charts/spending-structure-chart";
import { useMe } from "@/hooks/use-auth";
import LinearProgress from "@mui/material/LinearProgress";

type DashboardSummary = {
  totalIncome: string;
  totalExpense: string;
  savingsDeltaPct: number;
  goalProgressPct: number;
  trend: { date: string; income: string; expense: string }[];
  structure: { name: string; value: string; color: string }[];
  recentTransactions: {
    id: string;
    amount: string;
    type: string;
    note: string;
    occurredAt: string;
    category?: { name: string };
  }[];
  budgetLines: { categoryName: string; usedPercent: number }[];
};

export default function OverviewPage() {
  const { data: user } = useMe();
  const range = useUIStore((s) => s.dashboardRange);
  const setRange = useUIStore((s) => s.setDashboardRange);
  const { data } = useQuery({
    queryKey: ["dashboard", range],
    queryFn: () => apiFetch<DashboardSummary>(`/dashboard/summary?range=${range}`),
  });

  return (
    <Box>
      <Typography variant="h5" sx={{ fontWeight: 700 }} gutterBottom>
        Xin chào, {user?.displayName}!
      </Typography>
      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat title="Tổng thu" value={formatShortVND(data?.totalIncome ?? 0)} />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat title="Tổng chi" value={formatShortVND(data?.totalExpense ?? 0)} />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat title="Tiết kiệm" value={`${data?.savingsDeltaPct ?? 0}%`} />
        </Grid>
        <Grid size={{ xs: 6, md: 3 }}>
          <Stat title="Mục tiêu" value={`${data?.goalProgressPct ?? 0}%`} />
        </Grid>
      </Grid>
      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 8 }}>
          <Card>
            <CardContent>
              <Tabs value={range} onChange={(_, v) => setRange(v)}>
                <Tab value="7d" label="7 ngày" />
                <Tab value="30d" label="30 ngày" />
                <Tab value="60d" label="60 ngày" />
              </Tabs>
              <SpendingTrendChart data={data?.trend ?? []} />
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 12, lg: 4 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Cơ cấu chi
              </Typography>
              <SpendingStructureChart data={data?.structure ?? []} />
            </CardContent>
          </Card>
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Giao dịch gần đây
              </Typography>
              {(data?.recentTransactions ?? []).map((tx) => (
                <Box key={tx.id} sx={{ display: "flex", justifyContent: "space-between", py: 1 }}>
                  <Box>
                    <Typography variant="body2" sx={{ fontWeight: 600 }}>
                      {tx.category?.name}
                    </Typography>
                    <Typography variant="caption" color="text.secondary">
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
        </Grid>
        <Grid size={{ xs: 12, md: 6 }}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                Ngân sách
              </Typography>
              {(data?.budgetLines ?? []).map((line) => (
                <Box key={line.categoryName} sx={{ mb: 2 }}>
                  <Typography variant="body2">
                    {line.categoryName} — {line.usedPercent}%
                  </Typography>
                  <LinearProgress variant="determinate" value={Math.min(line.usedPercent, 100)} />
                </Box>
              ))}
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
}

function Stat({ title, value }: { title: string; value: string }) {
  return (
    <Card>
      <CardContent>
        <Typography variant="body2" color="text.secondary">
          {title}
        </Typography>
        <Typography variant="h6" sx={{ fontWeight: 700 }}>
          {value}
        </Typography>
      </CardContent>
    </Card>
  );
}
