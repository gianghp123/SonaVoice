"use client"

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { PAGE_ROUTES } from "@/lib/routes"
import { useRouter } from "next/navigation"
import { useT } from "next-i18next/client"

interface SessionErrorModalProps {
  message: string
}

export function SessionErrorModal({ message }: SessionErrorModalProps) {
  const router = useRouter()
  const { t } = useT("session")

  return (
    <AlertDialog open>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{t("session_unavailable")}</AlertDialogTitle>
          <AlertDialogDescription>{message}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogAction onClick={() => router.push(PAGE_ROUTES.HOME)}>
            {t("return_home")}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
