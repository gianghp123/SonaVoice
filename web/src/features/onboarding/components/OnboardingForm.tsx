"use client"

import { Button } from "@/components/ui/button"
import { Checkbox } from "@/components/ui/checkbox"
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
  FieldSet,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { RadioGroup, RadioGroupItem } from "@/components/ui/radio-group"
import { Textarea } from "@/components/ui/textarea"
import { PAGE_ROUTES } from "@/lib/routes"
import { useUser } from "@clerk/nextjs"
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"
import { useRouter } from "next/navigation"
import { useState } from "react"
import { Controller, useForm } from "react-hook-form"
import { toast } from "sonner"
import {
  ENGLISH_LEVELS,
  IMPROVEMENT_GOALS,
  LEARNING_REASONS,
  TOPICS,
} from "../constants/onboarding.constants"
import { onboardingSchema, type OnboardingFormValues } from "../schemas/onboarding.schema"
import { upsertProfile } from "../services/profile.actions"

interface OnboardingFormProps {
  defaultValues?: Partial<OnboardingFormValues>
}

export function OnboardingForm({ defaultValues }: OnboardingFormProps) {
  const router = useRouter()
  const { user } = useUser()
  const [loading, setLoading] = useState(false)

  const form = useForm<OnboardingFormValues>({
    resolver: zodResolver(onboardingSchema),
    defaultValues: {
      displayName: "",
      nativeLanguage: "",
      englishLevel: undefined,
      improvementGoals: [],
      topics: [],
      customTopics: "",
      learningReason: [],
      customLearningReason: "",
      ...defaultValues,
    },
  })

  const watchedTopics = form.watch("topics")
  const watchedLearningReason = form.watch("learningReason")
  const showCustomTopics = watchedTopics?.includes("Other")
  const showCustomLearningReason = watchedLearningReason?.includes("Other")

  async function onSubmit(data: OnboardingFormValues) {
    setLoading(true)

    try {
      const result = await upsertProfile(data)

      if (result.error) {
        toast.error(result.error.message || "Failed to save profile")
        return
      }

      toast.success("Profile saved!")
      await user?.reload()
      router.push(PAGE_ROUTES.HOME)
    } catch {
      toast.error("Something went wrong")
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-8">
      <FieldGroup>
        {/* Display Name */}
        <Controller
          name="displayName"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="displayName">
                What should we call you? *
              </FieldLabel>
              <Input
                {...field}
                id="displayName"
                placeholder='e.g. "Giang"'
                aria-invalid={fieldState.invalid}
              />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />

        {/* Native Language */}
        <Controller
          name="nativeLanguage"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="nativeLanguage">
                What is your native language?
              </FieldLabel>
              <Input
                {...field}
                id="nativeLanguage"
                placeholder='e.g. "Vietnamese"'
                aria-invalid={fieldState.invalid}
              />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />

        {/* English Level */}
        <Controller
          name="englishLevel"
          control={form.control}
          render={({ field, fieldState }) => (
            <FieldSet data-invalid={fieldState.invalid}>
              <FieldLabel>What is your English level? *</FieldLabel>
              <FieldDescription>
                This helps us adjust the difficulty of conversations.
              </FieldDescription>
              <RadioGroup
                name={field.name}
                value={field.value}
                onValueChange={field.onChange}
              >
                {ENGLISH_LEVELS.map((level) => (
                  <FieldLabel
                    key={level.value}
                    htmlFor={`english-level-${level.value}`}
                  >
                    <Field orientation="horizontal" data-invalid={fieldState.invalid}>
                      <span>{level.label}</span>
                      <RadioGroupItem
                        value={level.value}
                        id={`english-level-${level.value}`}
                        aria-invalid={fieldState.invalid}
                      />
                    </Field>
                  </FieldLabel>
                ))}
              </RadioGroup>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </FieldSet>
          )}
        />

        {/* Improvement Goals */}
        <Controller
          name="improvementGoals"
          control={form.control}
          render={({ field, fieldState }) => (
            <FieldSet data-invalid={fieldState.invalid}>
              <FieldLabel>What do you want to improve most?</FieldLabel>
              <FieldDescription>Select all that apply.</FieldDescription>
              <FieldGroup className="grid grid-cols-2 gap-2">
                {IMPROVEMENT_GOALS.map((goal) => (
                  <FieldLabel
                    key={goal}
                    htmlFor={`goal-${goal}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`goal-${goal}`}
                      checked={field.value?.includes(goal)}
                      onCheckedChange={(checked) => {
                        const current = field.value || []
                        if (checked) {
                          field.onChange([...current, goal])
                        } else {
                          field.onChange(current.filter((g) => g !== goal))
                        }
                      }}
                    />
                    <span>{goal}</span>
                  </FieldLabel>
                ))}
              </FieldGroup>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </FieldSet>
          )}
        />

        {/* Topics */}
        <Controller
          name="topics"
          control={form.control}
          render={({ field, fieldState }) => (
            <FieldSet data-invalid={fieldState.invalid}>
              <FieldLabel>What topics do you enjoy talking about?</FieldLabel>
              <FieldDescription>Select all that apply.</FieldDescription>
              <FieldGroup className="grid grid-cols-2 gap-2">
                {TOPICS.map((topic) => (
                  <FieldLabel
                    key={topic}
                    htmlFor={`topic-${topic}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`topic-${topic}`}
                      checked={field.value?.includes(topic)}
                      onCheckedChange={(checked) => {
                        const current = field.value || []
                        if (checked) {
                          field.onChange([...current, topic])
                        } else {
                          field.onChange(current.filter((t) => t !== topic))
                        }
                      }}
                    />
                    <span>{topic}</span>
                  </FieldLabel>
                ))}
              </FieldGroup>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </FieldSet>
          )}
        />

        {/* Custom Topics (conditional) */}
        {showCustomTopics && (
          <Controller
            name="customTopics"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="customTopics">
                  What other topics?
                </FieldLabel>
                <FieldDescription>
                  Separate topics with commas.
                </FieldDescription>
                <Input
                  {...field}
                  id="customTopics"
                  placeholder='e.g. "AI, blockchain, music"'
                  aria-invalid={fieldState.invalid}
                />
                {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
              </Field>
            )}
          />
        )}

        {/* Learning Reason */}
        <Controller
          name="learningReason"
          control={form.control}
          render={({ field, fieldState }) => (
            <FieldSet data-invalid={fieldState.invalid}>
              <FieldLabel>Why are you learning English?</FieldLabel>
              <FieldDescription>Select all that apply.</FieldDescription>
              <FieldGroup className="grid grid-cols-2 gap-2">
                {LEARNING_REASONS.map((reason) => (
                  <FieldLabel
                    key={reason}
                    htmlFor={`reason-${reason}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`reason-${reason}`}
                      checked={field.value?.includes(reason)}
                      onCheckedChange={(checked) => {
                        const current = field.value || []
                        if (checked) {
                          field.onChange([...current, reason])
                        } else {
                          field.onChange(current.filter((r) => r !== reason))
                        }
                      }}
                    />
                    <span>{reason}</span>
                  </FieldLabel>
                ))}
              </FieldGroup>
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </FieldSet>
          )}
        />

        {/* Custom Learning Reason (conditional) */}
        {showCustomLearningReason && (
          <Controller
            name="customLearningReason"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="customLearningReason">
                  What other reasons?
                </FieldLabel>
                <FieldDescription>
                  Separate reasons with commas.
                </FieldDescription>
                <Textarea
                  {...field}
                  id="customLearningReason"
                  placeholder='e.g. "For remote work, for travel"'
                  aria-invalid={fieldState.invalid}
                  className="min-h-[80px]"
                />
                {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
              </Field>
            )}
          />
        )}
      </FieldGroup>

      <Button type="submit" disabled={loading} className="w-full">
        {loading && <Loader2 className="size-4 animate-spin mr-2" />}
        Get Started
      </Button>
    </form>
  )
}
