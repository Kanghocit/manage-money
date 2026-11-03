const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "/api/v1";

export type ApiError = { code: string; message: string };

type Envelope<T> = { data: T; error: ApiError | null };

export async function apiFetch<T>(
  path: string,
  init?: RequestInit,
): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...init?.headers,
    },
  });
  const body = (await res.json()) as Envelope<T>;
  if (!res.ok || body.error) {
    throw new Error(body.error?.message ?? res.statusText);
  }
  return body.data;
}

export { API_URL };
