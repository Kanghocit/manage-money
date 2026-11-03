export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-violet-50 to-white p-4 dark:from-zinc-950 dark:to-zinc-900">
      {children}
    </div>
  );
}
