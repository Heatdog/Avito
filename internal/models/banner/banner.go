package banner_model

import "time"

func init() {

}

type BannerInsert struct {
	TagsID    []int  `json:"tag_id,omitempty" validate:"required,min=1,dive,numeric,required"`
	FeatureID int    `json:"feature_id,omitempty" validate:"numeric,required"`
	Content   string `json:"content,omitempty" validate:"json,required"`
	IsActive  bool   `json:"is_active,omitempty" validate:"boolean,required"`
}

type BannerUserParams struct {
	TagID            int
	FeatureID        int
	UseLastrRevision bool
}

type Banner struct {
	ID        int
	TagsID    []int
	FeatureID int
	Content   string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type BannerParams struct {
	TagID     *int
	FeatureID *int
	Limit     *int
	Offset    *int
}
