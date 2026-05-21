package domain

import (
	"encoding/json"

	"github.com/gianghp123/SonaVoice/api/internal/modules/session/dtos"
	"gorm.io/datatypes"
)

func ParseSessionConfig(raw datatypes.JSON) (*dtos.ConfigPayload, error) {
	if len(raw) == 0 {
		return &dtos.ConfigPayload{}, nil
	}

	var payload dtos.ConfigPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
