"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

type Row = { month: string; income: string; expense: string };

export function MonthlyBarChart({ data }: { data: Row[] }) {
  const chartData = data.map((d) => ({
    month: d.month,
    income: Number(d.income),
    expense: Number(d.expense),
  }));

  return (
    <ResponsiveContainer width="100%" height={320}>
      <BarChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="month" />
        <YAxis />
        <Tooltip formatter={(v) => Number(v).toLocaleString("vi-VN") + " đ"} />
        <Legend />
        <Bar dataKey="income" name="Thu" fill="#22c55e" radius={[4, 4, 0, 0]} />
        <Bar dataKey="expense" name="Chi" fill="#7C3AED" radius={[4, 4, 0, 0]} />
      </BarChart>
    </ResponsiveContainer>
  );
}
