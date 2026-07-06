import { Skeleton } from "@/components/ui/skeleton"
import { SidebarGroup, SidebarGroupLabel, SidebarMenu, SidebarMenuSkeleton } from "@/components/ui/sidebar"

export function SessionSkeleton() {
  return (
    <SidebarGroup>
      <SidebarGroupLabel>
        <Skeleton className="h-4 w-24" />
      </SidebarGroupLabel>
      <SidebarMenu>
        {Array.from({ length: 5 }).map((_, i) => (
          <SidebarMenuSkeleton key={i} showIcon />
        ))}
      </SidebarMenu>
    </SidebarGroup>
  )
}
