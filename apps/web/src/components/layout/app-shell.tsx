"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  BarChart3,
  Bell,
  Home,
  Menu,
  PiggyBank,
  Plus,
  Search,
  Settings,
  Target,
  Wallet,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { vi } from "@/lib/i18n/vi";
import { ThemeToggle } from "@/components/layout/theme-toggle";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { useMe, useLogout } from "@/hooks/use-auth";

const navItems = [
  { href: "/overview", label: vi.nav.overview, icon: Home },
  { href: "/transactions", label: vi.nav.transactions, icon: Wallet },
  { href: "/budget", label: vi.nav.budget, icon: PiggyBank },
  { href: "/goals", label: vi.nav.goals, icon: Target },
  { href: "/statistics", label: vi.nav.statistics, icon: BarChart3 },
  { href: "/settings", label: vi.nav.settings, icon: Settings },
];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const { data: user } = useMe();
  const logout = useLogout();

  const sidebar = (
    <div className="flex h-full flex-col bg-sidebar text-sidebar-foreground">
      <div className="border-b border-sidebar-border p-4">
        <p className="text-xs text-muted-foreground">{vi.appName}</p>
        <p className="font-semibold">{user?.displayName ?? "..."}</p>
      </div>
      <nav className="flex-1 space-y-1 p-3">
        {navItems.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || pathname.startsWith(href + "/");
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
                active
                  ? "bg-sidebar-primary text-sidebar-primary-foreground"
                  : "hover:bg-sidebar-accent hover:text-sidebar-accent-foreground",
              )}
            >
              <Icon className="h-4 w-4" />
              {label}
            </Link>
          );
        })}
      </nav>
    </div>
  );

  return (
    <div className="min-h-screen bg-muted/30">
      <div className="mx-auto flex min-h-screen max-w-[1600px]">
        <aside className="hidden w-64 shrink-0 border-r md:block">{sidebar}</aside>
        <div className="flex min-h-screen flex-1 flex-col">
          <header className="sticky top-0 z-20 flex items-center gap-3 border-b bg-background/95 px-4 py-3 backdrop-blur">
            <Sheet>
              <SheetTrigger className="md:hidden">
                <Button variant="outline" size="icon" type="button">
                  <Menu className="h-4 w-4" />
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="w-64 p-0">
                {sidebar}
              </SheetContent>
            </Sheet>
            <div className="relative hidden flex-1 md:block">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input className="pl-9" placeholder="Tìm giao dịch..." />
            </div>
            <Button variant="ghost" size="icon" className="md:hidden">
              <Search className="h-4 w-4" />
            </Button>
            <ThemeToggle />
            <Button variant="ghost" size="icon">
              <Bell className="h-5 w-5" />
            </Button>
            <Avatar className="h-8 w-8">
              <AvatarFallback>
                {(user?.displayName ?? "U").slice(0, 1).toUpperCase()}
              </AvatarFallback>
            </Avatar>
            <Button variant="outline" size="sm" onClick={() => logout.mutate()}>
              Thoát
            </Button>
          </header>
          <main className="flex-1 p-4 pb-24 md:p-6 md:pb-6">{children}</main>
        </div>
      </div>
      <nav className="fixed bottom-0 left-0 right-0 z-30 border-t bg-background md:hidden">
        <div className="mx-auto grid max-w-lg grid-cols-5 items-end px-2 py-2">
          {[
            { href: "/overview", icon: Home, label: "Home" },
            { href: "/transactions", icon: Wallet, label: "GD" },
            { href: "/transactions/new", icon: Plus, label: "Thêm", fab: true },
            { href: "/budget", icon: PiggyBank, label: "NS" },
            { href: "/settings", icon: Menu, label: "More" },
          ].map(({ href, icon: Icon, label, fab }) =>
            fab ? (
              <Link
                key={href}
                href={href}
                className="mx-auto -mt-6 flex h-14 w-14 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg"
              >
                <Icon className="h-6 w-6" />
              </Link>
            ) : (
              <Link
                key={href}
                href={href}
                className={cn(
                  "flex flex-col items-center gap-1 text-xs",
                  pathname.startsWith(href)
                    ? "text-primary"
                    : "text-muted-foreground",
                )}
              >
                <Icon className="h-5 w-5" />
                {label}
              </Link>
            ),
          )}
        </div>
      </nav>
    </div>
  );
}
