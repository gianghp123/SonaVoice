"use client"

import { Logo } from "@/components/common/Logo"
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
import { ConnectNow } from "@/features/landing/components/ConnectNow"
import { PAGE_ROUTES } from "@/lib/routes"
import type { ISession } from "@/lib/types/session.interface"
import { Show } from "@clerk/nextjs"
import { Plus } from "lucide-react"
import { useT } from "next-i18next/client"
import Link from "next/link"
import { usePathname } from "next/navigation"
import { Separator } from "../ui/separator"
import { LanguageSwitcher } from "./LanguageSwitcher"

interface HomePageContentProps {
  sessions: ISession[]
  children: React.ReactNode
  breadcrumb?: React.ReactNode
  sidebarFooter?: React.ReactNode
}

export function HomePageLayout({ sessions, children, breadcrumb, sidebarFooter }: HomePageContentProps) {
  const pathname = usePathname()
  const { t } = useT('common')

  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarHeader className="p-4">
          <Logo className="text-xl" />
        </SidebarHeader>
        <SidebarContent>
          <Show when="signed-in">
            <SidebarGroup>
              <SidebarMenu className="pb-2">
                <SidebarMenuItem>
                  <ConnectNow
                    variant="ghost"
                    className="justify-start px-2 font-normal"
                  >
                    <Plus className="size-4" />
                    <span>{t('new_session')}</span>
                  </ConnectNow>
                </SidebarMenuItem>
              </SidebarMenu>
            </SidebarGroup>
            <SidebarGroup>
              <SidebarGroupLabel>{t('recent_sessions')}</SidebarGroupLabel>
              <SidebarMenu>
                {sessions.map((session) => (
                  <SidebarMenuItem key={session.id}>
                    <SidebarMenuButton
                      asChild
                      isActive={pathname === PAGE_ROUTES.SESSION.DETAIL(session.id)}
                    >
                      <Link href={PAGE_ROUTES.SESSION.DETAIL(session.id)}>
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
                      {t('no_sessions')}
                    </p>
                  </SidebarMenuItem>
                )}
              </SidebarMenu>
            </SidebarGroup>
          </Show>
        </SidebarContent>
        <SidebarFooter>
          {sidebarFooter}
        </SidebarFooter>
      </Sidebar>
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center justify-between gap-2 border-b px-3">
          <div className="flex items-center gap-2">
            <SidebarTrigger />
            <Separator orientation="vertical" className="my-auto mr-2 h-4" />
            {breadcrumb}
          </div>

            <LanguageSwitcher />
        </header>
        {children}
      </SidebarInset>
    </SidebarProvider>
  )
}
