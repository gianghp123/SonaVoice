import { Button } from "@/components/ui/button"
import {Show, SignInButton, UserButton } from "@clerk/nextjs"
import { LanguageSwitcher } from "@/components/common/LanguageSwitcher"

export function SidebarFooterUI() {
  return (
    <div className="px-4 py-5">
      <Show when="signed-in">
        <UserButton />
      </Show>

      <Show when="signed-out">
        <div className="space-y-5">
          <div className="space-y-3">
            <h3 className="text-[15px] font-semibold leading-none text-foreground">
              Get responses tailored to you
            </h3>

            <p className="text-[14px] leading-5 text-muted-foreground">
              Log in to get answers based on saved chats
            </p>
          </div>

          <SignInButton>
            <Button
              // variant="outline"
              className="w-full rounded-full border-border"
            >
              Log in
            </Button>
          </SignInButton>
        </div>
      </Show>

      <div className="mt-4">
        <LanguageSwitcher />
      </div>
    </div>
  )
}