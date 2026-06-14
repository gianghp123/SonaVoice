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
import { formatDuration, formatTime, groupSessionsByDate, type DateGroup } from "@/lib/utils/date"
import { Show } from "@clerk/nextjs"
import { Plus, Podcast } from "lucide-react"
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

const DATE_GROUP_ORDER: DateGroup[] = ["today", "yesterday", "this_week", "older"]

export function HomePageLayout({ sessions, children, breadcrumb, sidebarFooter }: HomePageContentProps) {
  const pathname = usePathname()
  const { t } = useT('common')
  const groupedSessions = groupSessionsByDate(sessions)

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
              {sessions.length === 0 ? (
                <SidebarMenu>
                  <SidebarMenuItem>
                    <p className="px-2 text-sm text-muted-foreground">
                      {t('no_sessions')}
                    </p>
                  </SidebarMenuItem>
                </SidebarMenu>
              ) : (
                DATE_GROUP_ORDER.map((group) => {
                  const sessionsInGroup = groupedSessions[group]
                  if (sessionsInGroup.length === 0) return null

                  return (
                    <div key={group} className="mb-4">
                      <p className="px-2 py-1 text-xs font-medium text-muted-foreground">
                        {t(group)}
                      </p>
                      <SidebarMenu>
                        {sessionsInGroup.map((session) => {
                          const duration = formatDuration(session.actualUsage ?? 0)
                          return (
                            <SidebarMenuItem key={session.id}>
                              <SidebarMenuButton
                                asChild
                                isActive={pathname === PAGE_ROUTES.SESSION.DETAIL(session.id)}
                              >
                                <Link
                                  href={PAGE_ROUTES.SESSION.DETAIL(session.id)}
                                  className="flex items-center justify-between rounded-md px-2 py-2 hover:bg-muted"
                                >
                                  <div className="flex items-center gap-2 min-w-0">
                                    <Podcast className="size-4 shrink-0 text-primary" />
                                    <span className="truncate">
                                      {formatTime(session.createdAt)}
                                    </span>
                                  </div>

                                  {duration && (
                                    <span className="text-xs text-muted-foreground shrink-0">
                                      {duration}
                                    </span>
                                  )}
                                </Link>
                              </SidebarMenuButton>
                            </SidebarMenuItem>
                          )
                        })}
                      </SidebarMenu>
                    </div>
                  )
                })
              )}
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
        <main className="flex-1 overflow-auto">
          {children}
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
