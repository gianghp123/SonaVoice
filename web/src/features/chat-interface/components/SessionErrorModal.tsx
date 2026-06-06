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

interface SessionErrorModalProps {
  message: string
}

export function SessionErrorModal({ message }: SessionErrorModalProps) {
  const router = useRouter()

  return (
    <AlertDialog open>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Session Unavailable</AlertDialogTitle>
          <AlertDialogDescription>{message}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogAction onClick={() => router.push(PAGE_ROUTES.HOME)}>
            Return Home
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  )
}
