export function formatVND(amount: string | number): string {
  const n = typeof amount === "string" ? Number(amount) : amount;
  if (Number.isNaN(n)) return "0 đ";
  return new Intl.NumberFormat("vi-VN").format(n) + " đ";
}

export function formatShortVND(amount: string | number): string {
  const n = typeof amount === "string" ? Number(amount) : amount;
  if (Number.isNaN(n)) return "0";
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1).replace(/\.0$/, "") + "M đ";
  if (n >= 1_000) return (n / 1_000).toFixed(0) + "K đ";
  return formatVND(n);
}
