/*
 * Copyright (c) 2025-2026, s0up and the autobrr contributors.
 * SPDX-License-Identifier: GPL-2.0-or-later
 */

import { Outlet } from "@tanstack/react-router"
import { MobileFooterNav } from "@/components/layout/MobileFooterNav"
import { Header } from "@/components/layout/Header"
import { Sidebar } from "@/components/layout/Sidebar"
import { LayoutRouteProvider } from "@/contexts/LayoutRouteContext"
import { usePersistedSidebarState } from "@/hooks/usePersistedSidebarState"
import { Menu } from "lucide-react"
import { Button } from "@/components/ui/button"
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { MobileScrollProvider } from "@/contexts/MobileScrollContext"
import { TorrentSelectionProvider } from "@/contexts/TorrentSelectionContext"
import { ThemeValidator } from "@/components/themes/ThemeValidator"
import { ThemeEffectsLayer } from "@/components/themes/ThemeEffectsLayer"

function AppLayoutContent() {
  const [sidebarCollapsed, setSidebarCollapsed] = usePersistedSidebarState(false) // Desktop: persisted state

  return (
    <div className="relative flex h-[100dvh] bg-background">
      <ThemeEffectsLayer />
      {/* Desktop Sidebar - Collapsible */}
      <div className={cn(
        "relative z-10 hidden overflow-hidden transition-all duration-300 ease-out lg:flex",
        sidebarCollapsed ? "w-0 opacity-0" : "w-64 opacity-100"
      )}>
        <div className="w-64 flex-shrink-0">
          <Sidebar />
        </div>
      </div>

      <div className="relative z-10 flex min-w-0 flex-1 flex-col">
        <Header
          sidebarCollapsed={sidebarCollapsed}
        >
          {/* Desktop toggle button */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
                className="hidden lg:flex transition-transform duration-200 hover:scale-110"
              >
                <Menu className={cn(
                  "h-5 w-5 transition-transform duration-300",
                  sidebarCollapsed && "rotate-90"
                )} />
              </Button>
            </TooltipTrigger>
            <TooltipContent side="bottom">
              {sidebarCollapsed ? "Show sidebar" : "Hide sidebar"}
            </TooltipContent>
          </Tooltip>
        </Header>
        <main className={cn(
          "relative z-10 flex-1 overflow-y-auto",
          "pb-[calc(4rem+env(safe-area-inset-bottom))] lg:pb-0"
        )}>
          <Outlet />
        </main>
      </div>

      {/* Mobile Footer Navigation */}
      <MobileFooterNav />
    </div>
  )
}

export function AppLayout() {
  return (
    <LayoutRouteProvider>
      <ThemeValidator />
      <TorrentSelectionProvider>
        <MobileScrollProvider>
          <AppLayoutContent />
        </MobileScrollProvider>
      </TorrentSelectionProvider>
    </LayoutRouteProvider>
  )
}
