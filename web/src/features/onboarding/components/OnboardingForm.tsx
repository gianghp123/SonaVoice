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
import { useT } from "next-i18next/client"
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
}

const TOTAL_STEPS = 3

const STEP_FIELDS: Record<number, (keyof OnboardingFormValues)[]> = {
  1: ["displayName", "nativeLanguage"],
  2: ["englishLevel", "improvementGoals"],
  3: ["topics", "customTopics", "learningReason", "customLearningReason"],
}

export function OnboardingForm({ defaultValues }: OnboardingFormProps) {
  const router = useRouter()
  const { user } = useUser()
  const [loading, setLoading] = useState(false)
  const [currentStep, setCurrentStep] = useState(1)
  const { t } = useT("onboarding")

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
        toast.error(t("failed_save_profile"))
        return
      }

      toast.success(t("profile_saved"))
      await user?.reload()
      router.push(PAGE_ROUTES.HOME)
    } catch {
      toast.error(t("something_went_wrong"))
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
              {t("your_ai_partner")}
            </p>
          </Onboarding.Header>
          <FieldGroup className="mt-6">
            <Controller
              name="displayName"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor="displayName">
                    {t("what_should_we_call_you")} *
                  </FieldLabel>
                  <Input
                    {...field}
                    id="displayName"
                    placeholder={t("display_name_placeholder")}
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
                    {t("native_language")}
                  </FieldLabel>
                  <Input
                    {...field}
                    id="nativeLanguage"
                    placeholder={t("native_language_placeholder")}
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
            title={t("your_english")}
            description={t("english_level_description")}
          />
          <FieldGroup className="mt-6">
            <Controller
              name="englishLevel"
              control={form.control}
              render={({ field, fieldState }) => (
                <FieldSet data-invalid={fieldState.invalid}>
                  <FieldLabel>{t("what_is_english_level")} *</FieldLabel>
                  <FieldDescription>
                    {t("english_level_help")}
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
                        <span className="flex-1">{t(level.i18nKey)}</span>
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
                  <FieldLabel>{t("what_to_improve")}</FieldLabel>
                  <FieldDescription>{t("select_all_apply")}</FieldDescription>
                  <FieldGroup className="grid grid-cols-2 gap-2">
                    {IMPROVEMENT_GOALS.map((goal) => (
                      <FieldLabel
                        key={goal.value}
                        htmlFor={`goal-${goal.value}`}
                        className="flex items-center gap-2"
                      >
                        <Checkbox
                          id={`goal-${goal.value}`}
                          checked={field.value?.includes(goal.value)}
                          onCheckedChange={(checked) => {
                            const current = field.value || []
                            if (checked) {
                              field.onChange([...current, goal.value])
                            } else {
                              field.onChange(current.filter((g) => g !== goal.value))
                            }
                          }}
                        />
                        <span>{t(goal.i18nKey)}</span>
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
            title={t("preferences")}
            description={t("choose_topics_reasons")}
          />
          <FieldGroup className="mt-6">
            <Controller
              name="topics"
              control={form.control}
              render={({ field, fieldState }) => (
                <FieldSet data-invalid={fieldState.invalid}>
                  <FieldLabel>{t("what_topics_enjoy")}</FieldLabel>
                  <FieldDescription>{t("select_all_apply")}</FieldDescription>
                  <FieldGroup className="grid grid-cols-2 gap-2">
                    {TOPICS.map((topic) => (
                      <FieldLabel
                        key={topic.value}
                        htmlFor={`topic-${topic.value}`}
                        className="flex items-center gap-2"
                      >
                        <Checkbox
                          id={`topic-${topic.value}`}
                          checked={field.value?.includes(topic.value)}
                          onCheckedChange={(checked) => {
                            const current = field.value || []
                            if (checked) {
                              field.onChange([...current, topic.value])
                            } else {
                              field.onChange(current.filter((t) => t !== topic.value))
                            }
                          }}
                        />
                        <span>{t(topic.i18nKey)}</span>
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
                      {t("what_other_topics")}
                    </FieldLabel>
                    <FieldDescription>
                      {t("separate_topics_commas")}
                    </FieldDescription>
                    <Input
                      {...field}
                      id="customTopics"
                      placeholder={t("custom_topics_placeholder")}
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
                  <FieldLabel>{t("why_learning_english")}</FieldLabel>
                  <FieldDescription>{t("select_all_apply")}</FieldDescription>
                  <FieldGroup className="grid grid-cols-2 gap-2">
                    {LEARNING_REASONS.map((reason) => (
                      <FieldLabel
                        key={reason.value}
                        htmlFor={`reason-${reason.value}`}
                        className="flex items-center gap-2"
                      >
                        <Checkbox
                          id={`reason-${reason.value}`}
                          checked={field.value?.includes(reason.value)}
                          onCheckedChange={(checked) => {
                            const current = field.value || []
                            if (checked) {
                              field.onChange([...current, reason.value])
                            } else {
                              field.onChange(current.filter((r) => r !== reason.value))
                            }
                          }}
                        />
                        <span>{t(reason.i18nKey)}</span>
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
                      {t("what_other_reasons")}
                    </FieldLabel>
                    <FieldDescription>
                      {t("separate_reasons_commas")}
                    </FieldDescription>
                    <Textarea
                      {...field}
                      id="customLearningReason"
                      placeholder={t("custom_reasons_placeholder")}
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
            {t("back")}
          </Button>
          {isLastStep ? (
            <Button
              type="submit"
              disabled={loading}
              className="flex-1 rounded-xl py-5"
            >
              {loading && <Loader2 className="size-4 animate-spin mr-2" />}
              {t("get_started")}
            </Button>
          ) : (
            <Button
              type="button"
              className="flex-1 rounded-xl py-5"
              onClick={handleNext}
            >
              {t("next")}
            </Button>
          )}
        </div>
      </form>
    </Onboarding>
  )
}
