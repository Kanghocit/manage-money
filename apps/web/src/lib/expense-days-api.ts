import { apiFetchRaw, API_URL } from "@/lib/api-client";

export type ExpenseDayLine = {
  id?: string;
  name: string;
  amount: string;
  categoryId?: string;
  sortOrder?: number;
};

export type ExpenseDayEntry = {
  id: string;
  entryDate: string;
  note: string;
  status: string;
  statusMessage?: string;
  imageUrl?: string;
  totalAmount: string;
  lines: ExpenseDayLine[];
  updatedAt?: string;
};

export async function uploadExpenseImage(entryId: string, file: File) {
  const form = new FormData();
  form.append("image", file);
  const { res, body } = await apiFetchRaw<ExpenseDayEntry>(`/expense-days/${entryId}/image`, {
    method: "POST",
    body: form,
  });
  if (!res.ok || body.error) {
    throw new Error(body.error?.message ?? res.statusText);
  }
  return body.data as ExpenseDayEntry;
}

export { API_URL };
