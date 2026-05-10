export function PageLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="flex h-screen overflow-hidden bg-background">
      {children}
    </main>
  )
}
