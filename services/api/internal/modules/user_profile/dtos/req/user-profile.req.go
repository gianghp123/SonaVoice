package req

type UpsertProfileReq struct {
	DisplayName          string   `json:"display_name" binding:"required"`
	NativeLanguage       string   `json:"native_language"`
	EnglishLevel         string   `json:"english_level" binding:"required,oneof=beginner intermediate advanced not_sure"`
	ImprovementGoals     []string `json:"improvement_goals"`
	Topics               []string `json:"topics"`
	CustomTopics         string   `json:"custom_topics"`
	LearningReason       []string `json:"learning_reason"`
	CustomLearningReason string   `json:"custom_learning_reason"`
}
