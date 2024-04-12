package queryparams

import (
	"strconv"
)

type BannerUserParams struct {
	TagID            string `validate:"required,numeric"`
	FeatureID        string `validate:"required,numeric"`
	UseLastrRevision string `validate:"omitempty,boolean"`
	Version          string `validate:"omitempty,numeric,min=1,max=3"`
	Token            string
}

type BannerParams struct {
	TagID     *int
	FeatureID *int
	Limit     *int
	Offset    *int
}

func ValidateBannersParams(tagStr, featureStr, limitStr,
	offsetStr string) (BannerParams, error) {
	res := BannerParams{}

	if tagStr != "" {
		tag, err := strconv.Atoi(tagStr)
		if err != nil {
			return BannerParams{}, err
		}

		res.TagID = &tag
	}

	if featureStr != "" {
		featureID, err := strconv.Atoi(featureStr)
		if err != nil {
			return BannerParams{}, err
		}

		res.FeatureID = &featureID
	}

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return BannerParams{}, err
		}

		res.Limit = &limit
	}

	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return BannerParams{}, err
		}

		res.Offset = &offset
	}

	return res, nil
}

type DeleteBannerParams struct {
	TagID     *int
	FeatureID *int
}

func ValidateDeleteBannerParams(tagStr, featureStr string) (DeleteBannerParams, error) {
	res := DeleteBannerParams{}

	if tagStr != "" {
		tag, err := strconv.Atoi(tagStr)
		if err != nil {
			return DeleteBannerParams{}, err
		}

		res.TagID = &tag
	}

	if featureStr != "" {
		featureID, err := strconv.Atoi(featureStr)
		if err != nil {
			return DeleteBannerParams{}, err
		}

		res.FeatureID = &featureID
	}

	return res, nil
}
