import { z } from "zod"

export const onboardingSchema = z.object({
  displayName: z.string().min(1, "Name is required"),
  nativeLanguage: z.string().optional(),
  englishLevel: z.enum(
    ["beginner", "intermediate", "advanced", "not_sure"],
    { message: "English level is required" }
  ),
  improvementGoals: z.array(z.string()).optional(),
  topics: z.array(z.string()).optional(),
  customTopics: z.string().optional(),
  learningReason: z.array(z.string()).optional(),
  customLearningReason: z.string().optional(),
})

export type OnboardingFormValues = z.infer<typeof onboardingSchema>
