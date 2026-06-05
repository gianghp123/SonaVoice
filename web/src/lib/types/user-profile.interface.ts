export interface IUserProfile {
  id: string
  userId: string
  displayName: string
  englishLevel: string
  preferences: IUserProfilePreferences
  createdAt: string
  updatedAt: string
}

export interface IUserProfilePreferences {
  nativeLanguage?: string
  improvementGoals?: string[]
  topics?: string[]
  customTopics?: string
  learningReason?: string[]
  customLearningReason?: string
}

export interface IUpsertProfileDto {
  displayName: string
  nativeLanguage?: string
  englishLevel: string
  improvementGoals?: string[]
  topics?: string[]
  customTopics?: string
  learningReason?: string[]
  customLearningReason?: string
}
