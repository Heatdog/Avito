package banner_model

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
)

func ValidateJson(fl validator.FieldLevel) bool {
	data, err := json.Marshal(fl.Field().Interface())
	if err != nil {
		return false
	}

	var js map[string]interface{}
	if err := json.Unmarshal(data, &js); err != nil {
		return false
	}
	return true
}

type BannerInsert struct {
	TagsID    []int       `json:"tag_id,omitempty" validate:"required,min=1,dive,numeric"`
	FeatureID int         `json:"feature_id,omitempty" validate:"numeric"`
	Content   interface{} `json:"content,omitempty" validate:"json,required" swaggertype:"object"`
	IsActive  bool        `json:"is_active,omitempty" validate:"boolean"`
}

type Banner struct {
	ID        int         `json:"banner_id"`
	TagsID    []int       `json:"tag_ids"`
	FeatureID int         `json:"feature_id"`
	Content   interface{} `json:"content" swaggertype:"object"`
	IsActive  bool        `json:"is_active"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type BannerKey struct {
	TagID     int
	FeatureID int
}

type BannerParams struct {
	TagIDs    []int
	FeatureID int
}
