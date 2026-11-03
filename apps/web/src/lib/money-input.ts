/** Strip everything except digits (API amount string). */
export function digitsOnly(value: string): string {
  return value.replace(/\D/g, "");
}

/** Format digit string with vi-VN grouping for display while typing. */
export function formatMoneyTyping(digits: string): string {
  if (!digits) return "";
  const n = Number(digits);
  if (Number.isNaN(n)) return "";
  return new Intl.NumberFormat("vi-VN").format(n);
}

export function hasPositiveAmount(digits: string): boolean {
  return digitsOnly(digits).length > 0 && Number(digitsOnly(digits)) > 0;
}
