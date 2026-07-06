"use client"

import { SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuButton, SidebarMenuItem } from "@/components/ui/sidebar"
import { PAGE_ROUTES } from "@/lib/routes"
import type { ISession } from "@/lib/types/session.interface"
import { formatDuration, formatTime, groupSessionsByDate, type DateGroup } from "@/lib/utils/date"
import { Podcast } from "lucide-react"
import { useT } from "next-i18next/client"
import Link from "next/link"
import { usePathname } from "next/navigation"

const DATE_GROUP_ORDER: DateGroup[] = ["today", "yesterday", "this_week", "older"]

interface SessionListClientProps {
  sessions: ISession[]
}

export function SessionListClient({ sessions }: SessionListClientProps) {
  const pathname = usePathname()
  const { t } = useT('common')
  const groupedSessions = groupSessionsByDate(sessions)

  if (sessions.length === 0) {
    return (
      <SidebarGroup>
        <SidebarMenu>
          <SidebarMenuItem>
            <p className="px-2 text-sm text-muted-foreground">{t('no_sessions')}</p>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarGroup>
    )
  }

  return (
    <SidebarGroup>
      <SidebarGroupLabel>{t('recent_sessions')}</SidebarGroupLabel>
      {DATE_GROUP_ORDER.map((group) => {
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
                        className="flex min-h-11 items-center justify-between rounded-md px-2 py-2 hover:bg-muted"
                      >
                        <div className="flex items-center gap-2 min-w-0">
                          <Podcast className="size-4 shrink-0 text-primary" />
                          <span className="truncate">{formatTime(session.createdAt)}</span>
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
      })}
    </SidebarGroup>
  )
}
