export const ENGLISH_LEVELS = [
  { value: "beginner", i18nKey: "level_beginner" },
  { value: "intermediate", i18nKey: "level_intermediate" },
  { value: "advanced", i18nKey: "level_advanced" },
  { value: "not_sure", i18nKey: "level_not_sure" },
] as const

export const IMPROVEMENT_GOALS = [
  { value: "Fluency", i18nKey: "goal_fluency" },
  { value: "Pronunciation", i18nKey: "goal_pronunciation" },
  { value: "Vocabulary", i18nKey: "goal_vocabulary" },
  { value: "Grammar", i18nKey: "goal_grammar" },
  { value: "Listening", i18nKey: "goal_listening" },
  { value: "Interview", i18nKey: "goal_interview" },
  { value: "IELTS/TOEFL", i18nKey: "goal_ielts_toefl" },
  { value: "Business English", i18nKey: "goal_business_english" },
  { value: "Daily conversation", i18nKey: "goal_daily_conversation" },
] as const

export const TOPICS = [
  { value: "Technology", i18nKey: "topic_technology" },
  { value: "Movies", i18nKey: "topic_movies" },
  { value: "School", i18nKey: "topic_school" },
  { value: "Work", i18nKey: "topic_work" },
  { value: "Travel", i18nKey: "topic_travel" },
  { value: "Food", i18nKey: "topic_food" },
  { value: "Games", i18nKey: "topic_games" },
  { value: "Daily life", i18nKey: "topic_daily_life" },
  { value: "Other", i18nKey: "topic_other" },
] as const

export const LEARNING_REASONS = [
  { value: "For interviews", i18nKey: "reason_interviews" },
  { value: "For studying abroad", i18nKey: "reason_studying_abroad" },
  { value: "For work", i18nKey: "reason_work" },
  { value: "For travel", i18nKey: "reason_travel" },
  { value: "Other", i18nKey: "reason_other" },
] as const

export const IMPROVEMENT_GOAL_VALUES = IMPROVEMENT_GOALS.map((g) => g.value)
export const TOPIC_VALUES = TOPICS.map((t) => t.value)
export const LEARNING_REASON_VALUES = LEARNING_REASONS.map((r) => r.value)
