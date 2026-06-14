import type { ISession } from "@/lib/types/session.interface"

export type DateGroup = "today" | "yesterday" | "this_week" | "older"

export function groupSessionsByDate(sessions: ISession[]): Record<DateGroup, ISession[]> {
  const today = new Date()
  const yesterday = new Date(today)
  yesterday.setDate(yesterday.getDate() - 1)

  const groups: Record<DateGroup, ISession[]> = {
    today: [],
    yesterday: [],
    this_week: [],
    older: [],
  }

  for (const session of sessions) {
    const date = new Date(session.createdAt)
    if (isSameDay(date, today)) {
      groups.today.push(session)
    } else if (isSameDay(date, yesterday)) {
      groups.yesterday.push(session)
    } else if (isThisWeek(date)) {
      groups.this_week.push(session)
    } else {
      groups.older.push(session)
    }
  }

  return groups
}

function isSameDay(d1: Date, d2: Date) {
  return d1.toDateString() === d2.toDateString()
}

function isThisWeek(date: Date) {
  const now = new Date()

  const startOfWeek = new Date(now)
  startOfWeek.setDate(now.getDate() - now.getDay())
  startOfWeek.setHours(0, 0, 0, 0)

  const todayStart = new Date(now)
  todayStart.setHours(0, 0, 0, 0)

  return date >= startOfWeek && date < todayStart
}

export function formatDuration(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  return mins < 1 ? "< 1 min" : `${mins} min`
}

export function formatTime(dateString: string): string {
  const date = new Date(dateString)
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" })
}
