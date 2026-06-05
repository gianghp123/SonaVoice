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
        toast.error(result.error.message || "Failed to update profile")
        return
      }

      toast.success("Profile updated!")
      onSuccess?.()
    } catch {
      toast.error("Something went wrong")
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
                    htmlFor={`edit-english-level-${level.value}`}
                    className="flex items-center gap-3 rounded-lg border p-3 cursor-pointer has-data-[state=checked]:border-foreground has-data-[state=checked]:bg-muted"
                  >
                    <span className="flex-1">{level.label}</span>
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
              <FieldLabel>What do you want to improve most?</FieldLabel>
              <FieldDescription>Select all that apply.</FieldDescription>
              <FieldGroup className="grid grid-cols-2 gap-2">
                {IMPROVEMENT_GOALS.map((goal) => (
                  <FieldLabel
                    key={goal}
                    htmlFor={`edit-goal-${goal}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`edit-goal-${goal}`}
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
                    htmlFor={`edit-topic-${topic}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`edit-topic-${topic}`}
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
                <FieldLabel htmlFor="edit-customTopics">
                  What other topics?
                </FieldLabel>
                <FieldDescription>
                  Separate topics with commas.
                </FieldDescription>
                <Input
                  {...field}
                  id="edit-customTopics"
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
                    htmlFor={`edit-reason-${reason}`}
                    className="flex items-center gap-2"
                  >
                    <Checkbox
                      id={`edit-reason-${reason}`}
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
                <FieldLabel htmlFor="edit-customLearningReason">
                  What other reasons?
                </FieldLabel>
                <FieldDescription>
                  Separate reasons with commas.
                </FieldDescription>
                <Textarea
                  {...field}
                  id="edit-customLearningReason"
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
      <Button type="submit" disabled={loading} className="w-full rounded-xl py-5">
        {loading && <Loader2 className="size-4 animate-spin mr-2" />}
        Save
      </Button>
    </form>
  )
}
