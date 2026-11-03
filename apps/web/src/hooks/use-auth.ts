"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { apiFetch } from "@/lib/api-client";

export type User = {
  id: string;
  email: string;
  displayName: string;
  locale: string;
  avatarUrl?: string;
  role: string;
};

export function useMe() {
  return useQuery({
    queryKey: ["auth", "me"],
    queryFn: () => apiFetch<User>("/auth/me"),
    retry: false,
  });
}

export function useLogin() {
  const qc = useQueryClient();
  const router = useRouter();
  return useMutation({
    mutationFn: (body: { email: string; password: string }) =>
      apiFetch<User>("/auth/login", {
        method: "POST",
        body: JSON.stringify(body),
      }),
    onSuccess: (user) => {
      qc.setQueryData(["auth", "me"], user);
      router.push("/overview");
    },
  });
}

export function useRegister() {
  const qc = useQueryClient();
  const router = useRouter();
  return useMutation({
    mutationFn: (body: {
      email: string;
      password: string;
      displayName: string;
    }) =>
      apiFetch<User>("/auth/register", {
        method: "POST",
        body: JSON.stringify(body),
      }),
    onSuccess: (user) => {
      qc.setQueryData(["auth", "me"], user);
      router.push("/overview");
    },
  });
}

export function useLogout() {
  const qc = useQueryClient();
  const router = useRouter();
  return useMutation({
    mutationFn: () =>
      apiFetch<{ loggedOut: boolean }>("/auth/logout", { method: "POST" }),
    onSuccess: () => {
      qc.removeQueries({ queryKey: ["auth"] });
      router.push("/login");
    },
  });
}
