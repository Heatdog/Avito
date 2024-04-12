package bannermodel

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator/v10"
)

func ValidateJSON(fl validator.FieldLevel) bool {
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
	Content   interface{} `json:"content,omitempty" validate:"json,required" swaggertype:"object"`
	TagsID    []int       `json:"tag_id,omitempty" validate:"required,min=1,dive,numeric"`
	FeatureID int         `json:"feature_id,omitempty" validate:"required,numeric"`
	IsActive  bool        `json:"is_active,omitempty" validate:"omitempty,boolean"`
}

type BannerUpdate struct {
	Content   interface{} `json:"content,omitempty" validate:"omitnil,json" swaggertype:"object"`
	TagsID    *[]int      `json:"tag_id,omitempty" validate:"omitnil,min=1,dive,numeric"`
	ID        int         `json:"banner_id," validate:"numeric,required" swaggerignore:"true"`
	FeatureID *int        `json:"feature_id,omitempty" validate:"omitnil,numeric"`
	IsActive  *bool       `json:"is_active,omitempty" validate:"omitnil,boolean"`
}

type Banner struct {
	ContentV1 interface{} `json:"content_v1" swaggertype:"object"`
	ContentV2 interface{} `json:"content_v2" swaggertype:"object"`
	ContentV3 interface{} `json:"content_v3" swaggertype:"object"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	TagsID    []int       `json:"tag_ids"`
	ID        int         `json:"banner_id"`
	FeatureID int         `json:"feature_id"`
	IsActive  bool        `json:"is_active"`
}

type BannerKey struct {
	TagID     string
	FeatureID string
}

type BannerParams struct {
	TagIDs    []int
	FeatureID int
}
