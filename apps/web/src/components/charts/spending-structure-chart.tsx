"use client";

import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";

type Slice = { name: string; value: string; color: string };

export function SpendingStructureChart({ data }: { data: Slice[] }) {
  const chartData = data.map((d) => ({
    name: d.name,
    value: Number(d.value),
    color: d.color || "#7C3AED",
  }));

  if (chartData.length === 0) {
    return (
      <div className="flex h-[260px] items-center justify-center text-sm text-muted-foreground">
        Chưa có dữ liệu chi tiêu
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={260}>
      <PieChart>
        <Pie
          data={chartData}
          dataKey="value"
          nameKey="name"
          innerRadius={60}
          outerRadius={90}
          paddingAngle={2}
        >
          {chartData.map((entry, i) => (
            <Cell key={i} fill={entry.color} />
          ))}
        </Pie>
        <Tooltip formatter={(v) => Number(v).toLocaleString("vi-VN") + " đ"} />
      </PieChart>
    </ResponsiveContainer>
  );
}
