"use client";

import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";

type Point = { date: string; income: string; expense: string };

export function SpendingTrendChart({ data }: { data: Point[] }) {
  const chartData = data.map((d) => ({
    date: d.date.slice(5),
    income: Number(d.income),
    expense: Number(d.expense),
  }));

  return (
    <ResponsiveContainer width="100%" height={280}>
      <LineChart data={chartData}>
        <CartesianGrid strokeDasharray="3 3" className="stroke-border" />
        <XAxis dataKey="date" tick={{ fontSize: 12 }} />
        <YAxis tick={{ fontSize: 12 }} />
        <Tooltip formatter={(v) => Number(v).toLocaleString("vi-VN") + " đ"} />
        <Legend />
        <Line type="monotone" dataKey="income" name="Thu" stroke="#22c55e" strokeWidth={2} dot={false} />
        <Line type="monotone" dataKey="expense" name="Chi" stroke="#ef4444" strokeWidth={2} dot={false} />
      </LineChart>
    </ResponsiveContainer>
  );
}
