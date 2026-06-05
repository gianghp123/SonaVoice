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
import { zodResolver } from "@hookform/resolvers/zod"
import { Loader2 } from "lucide-react"
import { useState } from "react"
import { Controller, useForm } from "react-hook-form"
import { toast } from "sonner"
import { useT } from "next-i18next/client"
import {
  ENGLISH_LEVELS,
  IMPROVEMENT_GOALS,
  LEARNING_REASONS,
  TOPICS,
} from "@/features/onboarding/constants/onboarding.constants"
import { editProfileSchema, type EditProfileFormValues } from "../schemas/edit-profile.schema"
import { updateProfile } from "../services/profile.actions"
import type { IUserProfile } from "@/lib/types/user-profile.interface"

interface EditProfileFormProps {
  profile: IUserProfile | null
  onSuccess?: () => void
}

function mapProfileToFormValues(profile: IUserProfile): EditProfileFormValues {
  const prefs = profile.preferences || {}
  return {
    displayName: profile.displayName,
    englishLevel: profile.englishLevel as EditProfileFormValues["englishLevel"],
    nativeLanguage: prefs.nativeLanguage ?? "",
    improvementGoals: prefs.improvementGoals ?? [],
    topics: prefs.topics ?? [],
    customTopics: prefs.customTopics ?? "",
    learningReason: prefs.learningReason ?? [],
    customLearningReason: prefs.customLearningReason ?? "",
  }
}

export function EditProfileForm({ profile, onSuccess }: EditProfileFormProps) {
  const [loading, setLoading] = useState(false)
  const { t } = useT("onboarding")

  const form = useForm<EditProfileFormValues>({
    resolver: zodResolver(editProfileSchema),
    defaultValues: profile
      ? mapProfileToFormValues(profile)
      : {
          displayName: "",
          nativeLanguage: "",
          englishLevel: undefined,
          improvementGoals: [],
          topics: [],
          customTopics: "",
          learningReason: [],
          customLearningReason: "",
        },
  })

  const watchedTopics = form.watch("topics")
  const watchedLearningReason = form.watch("learningReason")
  const showCustomTopics = watchedTopics?.includes("Other")
  const showCustomLearningReason = watchedLearningReason?.includes("Other")

  async function onSubmit(data: EditProfileFormValues) {
    setLoading(true)

    try {
      const result = await updateProfile(data)

      if (result.error) {
        toast.error(t("failed_update_profile"))
        return
      }

      toast.success(t("profile_updated"))
      onSuccess?.()
    } catch {
      toast.error(t("something_went_wrong"))
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
      <FieldGroup>
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
                    htmlFor={`edit-english-level-${level.value}`}
                    className="flex items-center gap-3 rounded-lg border p-3 cursor-pointer has-data-[state=checked]:border-foreground has-data-[state=checked]:bg-muted"
                  >
                    <span className="flex-1">{t(level.i18nKey)}</span>
                    <RadioGroupItem
                      value={level.value}
                      id={`edit-english-level-${level.value}`}
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
                    htmlFor={`edit-goal-${goal.value}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`edit-goal-${goal.value}`}
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
                    htmlFor={`edit-topic-${topic.value}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`edit-topic-${topic.value}`}
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
                <FieldLabel htmlFor="edit-customTopics">
                  {t("what_other_topics")}
                </FieldLabel>
                <FieldDescription>
                  {t("separate_topics_commas")}
                </FieldDescription>
                <Input
                  {...field}
                  id="edit-customTopics"
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
                    htmlFor={`edit-reason-${reason.value}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`edit-reason-${reason.value}`}
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
                <FieldLabel htmlFor="edit-customLearningReason">
                  {t("what_other_reasons")}
                </FieldLabel>
                <FieldDescription>
                  {t("separate_reasons_commas")}
                </FieldDescription>
                <Textarea
                  {...field}
                  id="edit-customLearningReason"
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
      <Button type="submit" disabled={loading} className="w-full rounded-xl py-5">
        {loading && <Loader2 className="size-4 animate-spin mr-2" />}
        {t("save")}
      </Button>
    </form>
  )
}
