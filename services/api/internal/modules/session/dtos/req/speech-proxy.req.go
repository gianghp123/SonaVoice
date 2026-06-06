package req

type StartConnectionReq struct {
	EnableDefaultIceServers bool                `json:"enableDefaultIceServers"`
	Body                    StartConnectionBody `json:"body"`
}

type StartConnectionBody struct {
	UserID      string          `json:"user_id"`
	SessionID   string          `json:"session_id"`
	MaxDuration int64           `json:"max_duration"`
	UserProfile *UserProfileDTO `json:"user_profile,omitempty"`
}

type UserProfileDTO struct {
	DisplayName  string              `json:"display_name"`
	EnglishLevel string              `json:"english_level"`
	Preferences  *UserPreferencesDTO `json:"preferences,omitempty"`
}

type UserPreferencesDTO struct {
	NativeLanguage       string   `json:"native_language"`
	ImprovementGoals     []string `json:"improvement_goals"`
	Topics               []string `json:"topics"`
	CustomTopics         string   `json:"custom_topics"`
	LearningReason       []string `json:"learning_reason"`
	CustomLearningReason string   `json:"custom_learning_reason"`
}

type UserPreferencesPayload struct {
	NativeLanguage       string   `json:"native_language"`
	ImprovementGoals     []string `json:"improvement_goals"`
	Topics               []string `json:"topics"`
	CustomTopics         string   `json:"custom_topics"`
	LearningReason       []string `json:"learning_reason"`
	CustomLearningReason string   `json:"custom_learning_reason"`
}
