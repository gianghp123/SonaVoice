"use client"

import { ConnectNow } from "@/features/landing/components/ConnectNow"
import { SidebarGroup, SidebarMenu, SidebarMenuItem } from "@/components/ui/sidebar"
import { Plus } from "lucide-react"
import { useT } from "next-i18next/client"

export function NewSessionButton() {
  const { t } = useT('common')

  return (
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
  )
}
