package banner_model

type BannerInsert struct {
	TagsId    []int  `json:"tag_id,omitempty" validate:"required,min=1,dive,int,required"`
	FeatureId int    `json:"feature_id,omitempty" validate:"int,required"`
	Content   string `json:"content,omitempty" validate:"json,required"`
	IsActive  bool   `json:"is_active,omitempty" validate:"boolean,required"`
}
