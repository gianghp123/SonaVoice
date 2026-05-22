"use client"

import Link from "next/link"
import { usePathname } from "next/navigation"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { SidebarFooterUI } from "@/components/common/SidebarFooter"
import type { ISession } from "@/lib/types/session.interface"
import { Logo } from "@/components/common/Logo"

interface HomePageContentProps {
  sessions: ISession[]
  children: React.ReactNode
}

export function HomePageLayout({ sessions, children }: HomePageContentProps) {
  const pathname = usePathname()

  return (
    <SidebarProvider defaultOpen={true}>
      <Sidebar>
        <SidebarHeader>
          <Logo />
        </SidebarHeader>
        <SidebarContent>
          <SidebarGroup>
            <SidebarGroupLabel>Recently sessions</SidebarGroupLabel>
            <SidebarMenu>
              {sessions.map((session) => (
                <SidebarMenuItem key={session.id}>
                  <SidebarMenuButton
                    asChild
                    isActive={pathname === `/chat/${session.id}`}
                  >
                    <Link href={`/chat/${session.id}`}>
                      <span className="truncate">
                        {new Date(session.createdAt).toLocaleDateString()}
                      </span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
              {sessions.length === 0 && (
                <SidebarMenuItem>
                  <p className="px-2 text-sm text-muted-foreground">
                    No sessions yet
                  </p>
                </SidebarMenuItem>
              )}
            </SidebarMenu>
          </SidebarGroup>
        </SidebarContent>
        <SidebarFooter>
          <SidebarFooterUI />
        </SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <SidebarTrigger className="absolute top-4 left-4 z-10" />
        {children}
      </SidebarInset>
    </SidebarProvider>
  )
}
