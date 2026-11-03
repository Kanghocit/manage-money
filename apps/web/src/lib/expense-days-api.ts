import { API_URL } from "@/lib/api-client";

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
};

export async function apiFetchRaw<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    credentials: "include",
    headers: init?.headers,
  });
  const body = await res.json();
  if (!res.ok || body.error) {
    throw new Error(body.error?.message ?? res.statusText);
  }
  return body.data as T;
}

export async function uploadExpenseImage(entryId: string, file: File) {
  const form = new FormData();
  form.append("image", file);
  const res = await fetch(`${API_URL}/expense-days/${entryId}/image`, {
    method: "POST",
    credentials: "include",
    body: form,
  });
  const body = await res.json();
  if (!res.ok || body.error) {
    throw new Error(body.error?.message ?? res.statusText);
  }
  return body.data as ExpenseDayEntry;
}
