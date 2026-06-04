import type { Metadata } from "next";
import { Geist, Geist_Mono, Inter } from "next/font/google";
import "./globals.css";
import { ClerkProvider } from "@clerk/nextjs";
import { TooltipProvider } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import { PAGE_ROUTES, AUTH_ROUTES } from "@/lib/routes";
import { Toaster } from "sonner";

const inter = Inter({subsets:['latin'],variable:'--font-sans'});

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Sona Voice",
  description: "Real-time voice practice with AI",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={cn("h-full", "antialiased", geistSans.variable, geistMono.variable, "font-sans", inter.variable)}
    >
      <body className="min-h-full flex flex-col">
        <ClerkProvider
          signInUrl={AUTH_ROUTES.SIGN_IN}
          signUpUrl={AUTH_ROUTES.SIGN_UP}
          signInFallbackRedirectUrl={PAGE_ROUTES.HOME}
          signUpFallbackRedirectUrl={PAGE_ROUTES.HOME}
        >
          <TooltipProvider>{children}</TooltipProvider>
        </ClerkProvider>
        <Toaster />
      </body>
    </html>
  );
}
