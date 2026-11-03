import type { ExpenseDayEntry, ExpenseDayLine } from "@/lib/expense-days-api";

export function parseLineAmount(amount: ExpenseDayLine["amount"] | number): number {
  if (amount == null) return 0;
  if (typeof amount === "number") return Number.isFinite(amount) ? amount : 0;
  const s = String(amount).trim();
  if (!s) return 0;
  const n = Number(s.replace(/\D/g, "") || s);
  return Number.isNaN(n) ? 0 : n;
}

/** Tổng hiển thị: ưu tiên totalAmount API, fallback cộng lines (list đôi khi total chưa cập nhật). */
export function expenseEntryTotal(entry: Pick<ExpenseDayEntry, "totalAmount" | "lines">): number {
  const fromField = Number(String(entry.totalAmount ?? "").replace(/\D/g, ""));
  if (!Number.isNaN(fromField) && fromField > 0) return fromField;
  return (entry.lines ?? []).reduce((sum, line) => sum + parseLineAmount(line.amount), 0);
}

export function expenseEntryHasData(entry: ExpenseDayEntry): boolean {
  if (entry.imageUrl) return true;
  if ((entry.lines?.length ?? 0) > 0) return true;
  if (expenseEntryTotal(entry) > 0) return true;
  return entry.status !== "draft" && entry.status !== "";
}
