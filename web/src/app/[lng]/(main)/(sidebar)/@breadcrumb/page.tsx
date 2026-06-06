import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage
} from "@/components/ui/breadcrumb"
import { getT } from "next-i18next/server"

export default async function SessionBreadcrumb({
  params,
}: {
  params: Promise<{ lng: string }>
}) {
  const { lng } = await params
  const { t } = await getT("session", { lng })

  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbPage>{t("home")}</BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
  )
}
