import { AuthGuard } from "@/components/auth-guard";
import { MuiAppShell } from "@/components/layout/mui-app-shell";

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <AuthGuard>
      <MuiAppShell>{children}</MuiAppShell>
    </AuthGuard>
  );
}
