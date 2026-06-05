import type { Metadata } from "next"
import { Geist, Geist_Mono, Inter } from "next/font/google"
import "../globals.css"
import { ClerkProvider } from "@clerk/nextjs"
import { TooltipProvider } from "@/components/ui/tooltip"
import { cn } from "@/lib/utils"
import { PAGE_ROUTES, AUTH_ROUTES } from "@/lib/routes"
import { Toaster } from "sonner"
import { initServerI18next, getT, getResources, generateI18nStaticParams } from "next-i18next/server"
import { I18nProvider } from "next-i18next/client"
import i18nConfig from "../../../i18n.config"
import { getClerkLanguageKey } from "@/lib/i18n/clerk-localization"
import { SupportedLanguage } from "@/lib/i18n/i18n"

initServerI18next(i18nConfig)

const inter = Inter({ subsets: ["latin"], variable: "--font-sans" })
const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] })
const geistMono = Geist_Mono({ variable: "--font-geist-mono", subsets: ["latin"] })

export async function generateMetadata(): Promise<Metadata> {
  const { t } = await getT('common')

  return {
    title: t('site_title'),
    description: t('site_description'),
  }
}

export async function generateStaticParams() {
  return generateI18nStaticParams()
}

export default async function RootLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ lng: string }>
}) {
  const { lng } = await params
  const { i18n } = await getT()
  const resources = getResources(i18n)

  return (
    <html
      lang={lng}
      className={cn(
        "h-full",
        "antialiased",
        geistSans.variable,
        geistMono.variable,
        "font-sans",
        inter.variable
      )}
    >
      <body className="min-h-full flex flex-col">
        <ClerkProvider
          signInUrl={AUTH_ROUTES.SIGN_IN}
          signUpUrl={AUTH_ROUTES.SIGN_UP}
          signInFallbackRedirectUrl={PAGE_ROUTES.HOME}
          signUpFallbackRedirectUrl={PAGE_ROUTES.HOME}
          localization={getClerkLanguageKey(lng as SupportedLanguage)}
        >
          <I18nProvider language={lng} resources={resources}>
            <TooltipProvider>{children}</TooltipProvider>
          </I18nProvider>
        </ClerkProvider>
        <Toaster />
      </body>
    </html>
  )
}
