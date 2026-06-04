"use client"

import { Button } from "@/components/ui/button"
import {
  Item,
  ItemContent,
  ItemHeader
} from "@/components/ui/item"
import { ConnectNow } from "@/features/landing/components/ConnectNow"
import { useT } from "next-i18next/client"
import { Show, SignInButton } from "@clerk/nextjs"
import { Trans } from "react-i18next"

const audioBars = [30, 60, 90, 70, 40, 80, 50]

export function LandingHero() {
  const { t } = useT('home')
  return (
    <section className="relative flex min-h-[calc(100vh-4rem)] flex-1 items-center justify-center overflow-hidden px-4 py-16">
      <div className="absolute inset-0 -z-10 bg-background" />
      <div className="absolute left-1/2 top-1/2 -z-10 h-[760px] w-[760px] -translate-x-1/2 -translate-y-1/2 rounded-full bg-primary/10 blur-[120px]" />
      <div className="absolute left-1/2 top-[58%] -z-10 h-72 w-72 -translate-x-1/2 rounded-full bg-primary/15 blur-3xl" />

      <style>{`
        @keyframes audioPulse {
          0%, 100% {
            height: 10%;
          }
          50% {
            height: 100%;
          }
        }
      `}</style>

      <Item variant="muted" className="w-full flex-col items-center text-center bg-transparent">
        <ItemHeader>
          <div className="w-full max-w-5xl mx-auto px-4 py-12 flex flex-col items-center justify-center text-center gap-6">
            <h1 className="text-4xl font-bold tracking-tight text-primary lg:text-5xl leading-[1.15]">
              <Trans
                t={t}
                i18nKey="hero_title"
                components={{ 1: <span className="inline-block text-[#5c539b]" /> }}
              />
            </h1>

            <p className="max-w-2xl text-lg text-muted-foreground leading-relaxed">
              {t('hero_subtitle')}
            </p>
          </div>
        </ItemHeader>

        <ItemContent className="items-center space-y-10 px-6 pb-10 md:px-12 md:pb-14">
          <div className="relative mx-auto flex h-40 w-full max-w-md items-center justify-center">
            <div className="flex h-36 items-center justify-center gap-1.5">
              {audioBars.map((height, index) => (
                <span
                  key={index}
                  className={`block w-3 rounded-full shadow-sm md:w-3.5 ${index === 5 ? "bg-primary/25" : "bg-primary"
                    }`}
                  style={{
                    height: `${height}%`,
                    animation: "audioPulse 1.15s ease-in-out infinite",
                    animationDelay: `${index * 0.08}s`,
                  }}
                />
              ))}
            </div>
          </div>

          <Show when="signed-in">
            <ConnectNow />
          </Show>

          <Show when="signed-out">
            <div className="flex flex-col items-center gap-4">
              <p className="text-sm text-muted-foreground">
                {t('sign_in_prompt')}
              </p>
              <SignInButton>
                <Button size="lg" className="rounded-xl px-8">
                  {t('sign_in_button')}
                </Button>
              </SignInButton>
            </div>
          </Show>
        </ItemContent>
      </Item>
    </section>
  )
}
