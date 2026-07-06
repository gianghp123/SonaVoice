import { NewSessionButton } from "@/components/common/NewSessionButton"
import { ProfileSkeleton } from "@/components/common/ProfileSkeleton"
import { SessionList } from "@/components/common/SessionList"
import { SessionSkeleton } from "@/components/common/SessionSkeleton"
import { UserSetting } from "@/components/common/UserSetting"
import { LanguageSwitcher } from "@/components/common/LanguageSwitcher"
import { Logo } from "@/components/common/Logo"
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Separator } from "@/components/ui/separator"
import { Show } from "@clerk/nextjs"
import { auth } from "@clerk/nextjs/server"
import { Suspense } from "react"

export default async function SidebarLayout({
  children,
  breadcrumb,
}: {
  children: React.ReactNode
  breadcrumb: React.ReactNode
}) {
  await auth()

  return (
    <SidebarProvider>
      <Sidebar>
        <SidebarHeader className="p-4">
          <Logo className="text-xl" />
        </SidebarHeader>
        <SidebarContent>
          <Show when="signed-in">
            <NewSessionButton />

            <Suspense fallback={<SessionSkeleton />}>
              <SessionList />
            </Suspense>

            <Suspense fallback={<ProfileSkeleton />}>
              <UserSetting />
            </Suspense>
          </Show>
        </SidebarContent>
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
