import { z } from "zod"

export const onboardingSchema = z
  .object({
    displayName: z.string().min(1, "Name is required"),
    nativeLanguage: z.string().optional(),
    englishLevel: z.enum(
      ["beginner", "intermediate", "advanced", "not_sure"],
      { message: "English level is required" }
    ),
    improvementGoals: z
      .array(z.string())
      .min(1, "Select at least one improvement goal"),
    topics: z
      .array(z.string())
      .min(1, "Select at least one topic"),
    customTopics: z.string().optional(),
    learningReason: z
      .array(z.string())
      .min(1, "Select at least one reason"),

    customLearningReason: z.string().optional(),
  })
  .superRefine((data, ctx) => {
    if (data.topics.includes("Other") && !data.customTopics?.trim()) {
      ctx.addIssue({
        code: "custom",
        path: ["customTopics"],
        message: "Please enter your custom topics",
      })
    }
    if (
      data.learningReason.includes("Other") &&
      !data.customLearningReason?.trim()
    ) {
      ctx.addIssue({
        code: "custom",
        path: ["customLearningReason"],
        message: "Please enter your custom reason",
      })
    }
  })

export type OnboardingFormValues = z.infer<typeof onboardingSchema>