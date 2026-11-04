const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "/api/v1";

export type ApiError = { code: string; message: string };

type Envelope<T> = { data: T; error: ApiError | null };

let refreshPromise: Promise<boolean> | null = null;

async function tryRefreshSession(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = (async () => {
    try {
      const res = await fetch(`${API_URL}/auth/refresh`, {
        method: "POST",
        credentials: "include",
      });
      if (!res.ok) return false;
      const body = (await res.json()) as Envelope<unknown>;
      return !body.error;
    } catch {
      return false;
    } finally {
      refreshPromise = null;
    }
  })();
  return refreshPromise;
}

export async function apiFetchRaw<T>(
  path: string,
  init?: RequestInit,
  retried = false,
): Promise<{ res: Response; body: Envelope<T> }> {
  const res = await fetch(`${API_URL}${path}`, {
    ...init,
    credentials: "include",
    headers: {
      ...(init?.body instanceof FormData ? {} : { "Content-Type": "application/json" }),
      ...init?.headers,
    },
  });
  const body = (await res.json()) as Envelope<T>;
  if (res.status === 401 && !retried && !path.startsWith("/auth/")) {
    const ok = await tryRefreshSession();
    if (ok) return apiFetchRaw(path, init, true);
  }
  return { res, body };
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const { res, body } = await apiFetchRaw<T>(path, init);
  if (!res.ok || body.error) {
    throw new Error(body.error?.message ?? res.statusText);
  }
  return body.data;
}

export { API_URL };
