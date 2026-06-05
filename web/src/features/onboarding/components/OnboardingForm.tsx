"use client"

import { Logo } from "@/components/common/Logo"
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
import { Onboarding } from "@/components/ui/onboarding"
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
import { completeOnboarding } from "../services/onboarding.actions"

interface OnboardingFormProps {
  defaultValues?: Partial<OnboardingFormValues>
  onSuccess?: () => void
}

const TOTAL_STEPS = 3

const STEP_FIELDS: Record<number, (keyof OnboardingFormValues)[]> = {
  1: ["displayName", "nativeLanguage"],
  2: ["englishLevel", "improvementGoals"],
  3: ["topics", "customTopics", "learningReason", "customLearningReason"],
}

export function OnboardingForm({ defaultValues, onSuccess }: OnboardingFormProps) {
  const router = useRouter()
  const { user } = useUser()
  const [loading, setLoading] = useState(false)
  const [currentStep, setCurrentStep] = useState(1)

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
    }
  })

  const watchedTopics = form.watch("topics")
  const watchedLearningReason = form.watch("learningReason")
  const showCustomTopics = watchedTopics?.includes("Other")
  const showCustomLearningReason = watchedLearningReason?.includes("Other")

  async function handleNext(event: React.MouseEvent<HTMLButtonElement>) {
    event.preventDefault()

    const fields = STEP_FIELDS[currentStep]
    const valid = fields.length > 0 ? await form.trigger(fields) : true

    if (valid && currentStep < TOTAL_STEPS) {
      setCurrentStep((step) => step + 1)
    }
  }

  function handleBack() {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1)
    }
  }

  async function onSubmit(data: OnboardingFormValues) {
    setLoading(true)

    try {
      const result = await completeOnboarding(data)

      if (result.error) {
        toast.error(result.error.message || "Failed to save profile")
        return
      }

      toast.success("Profile saved!")
      await user?.reload()
      onSuccess?.()
      router.push(PAGE_ROUTES.HOME)
    } catch {
      toast.error("Something went wrong")
    } finally {
      setLoading(false)
    }
  }

  const isLastStep = currentStep === TOTAL_STEPS

  return (
    <Onboarding
      value={currentStep}
      onValueChange={setCurrentStep}
      totalSteps={TOTAL_STEPS}
      className="w-full max-w-lg"
    >
      <Onboarding.StepIndicator variant="pills" className="mb-6" />

      <form onSubmit={form.handleSubmit(onSubmit)}>
        {/* Step 1: Welcome + About You */}
        <Onboarding.Step step={1}>
          <Onboarding.Header>
            <Logo className="justify-center text-3xl mb-3" />
            <p className="text-muted-foreground text-base">
              Your AI English speaking partner
            </p>
          </Onboarding.Header>
          <FieldGroup className="mt-6">
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
          </FieldGroup>
        </Onboarding.Step>

        {/* Step 2: Your English */}
        <Onboarding.Step step={2}>
          <Onboarding.Header
            title="Your English"
            description="Help us understand your current level."
          />
          <FieldGroup className="mt-6">
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
                    value={field.value ?? ""}
                    onValueChange={field.onChange}
                    className="gap-2"
                  >
                    {ENGLISH_LEVELS.map((level) => (
                      <label
                        key={level.value}
                        htmlFor={`english-level-${level.value}`}
                        className="flex items-center gap-3 rounded-lg border p-3 cursor-pointer has-data-[state=checked]:border-foreground has-data-[state=checked]:bg-muted"
                      >
                        <span className="flex-1">{level.label}</span>
                        <RadioGroupItem
                          value={level.value}
                          id={`english-level-${level.value}`}
                        />
                      </label>
                    ))}
                  </RadioGroup>
                  {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                </FieldSet>
              )}
            />
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
          </FieldGroup>
        </Onboarding.Step>

        {/* Step 3: Preferences */}
        <Onboarding.Step step={3}>
          <Onboarding.Header
            title="Preferences"
            description="Choose topics and reasons for learning."
          />
          <FieldGroup className="mt-6">
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
                      className="min-h-20"
                    />
                    {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
                  </Field>
                )}
              />
            )}
          </FieldGroup>
        </Onboarding.Step>

        {/* Navigation */}
        <div className="flex gap-3 mt-8" data-slot="onboarding-navigation">
          <Button
            type="button"
            variant="outline"
            className="flex-1 rounded-xl py-5"
            disabled={currentStep === 1}
            onClick={handleBack}
          >
            Back
          </Button>
          {isLastStep ? (
            <Button
              type="submit"
              disabled={loading}
              className="flex-1 rounded-xl py-5"
            >
              {loading && <Loader2 className="size-4 animate-spin mr-2" />}
              Get Started
            </Button>
          ) : (
            <Button
              type="button"
              className="flex-1 rounded-xl py-5"
              onClick={handleNext}
            >
              Next
            </Button>
          )}
        </div>
      </form>
    </Onboarding>
  )
}
