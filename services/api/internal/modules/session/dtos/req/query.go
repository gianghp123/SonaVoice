package req

import "github.com/gianghp123/SonaVoice/api/internal/core/enums"

type SessionListQuery struct {
	UserID *string              `form:"-"    json:"-"`
	Status *enums.SessionStatus `form:"-"    json:"-"`
	Page   int                  `form:"page" json:"page"`
	Limit  int                  `form:"limit" json:"limit"`
}
