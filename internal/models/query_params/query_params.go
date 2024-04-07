package query_params

import (
	"fmt"
	"strconv"
)

type BannerUserParams struct {
	TagID            int
	FeatureID        int
	UseLastrRevision bool
}

func ValidateUserBannerParams(tagIdStr, featureIdStr,
	useLastRevisionStr string) (BannerUserParams, error) {

	tagId, err := strconv.Atoi(tagIdStr)
	if err != nil {
		return BannerUserParams{}, err
	}

	featureId, err := strconv.Atoi(featureIdStr)
	if err != nil {
		return BannerUserParams{}, err
	}

	useLastRevision := false
	if useLastRevisionStr != "" {
		switch useLastRevisionStr {
		case "true":
			useLastRevision = true
		case "false":
			useLastRevision = false
		default:
			err = fmt.Errorf("incorrect use_last_revision value")
			return BannerUserParams{}, err
		}
	}
	return BannerUserParams{
		TagID:            tagId,
		FeatureID:        featureId,
		UseLastrRevision: useLastRevision,
	}, nil
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
		featureId, err := strconv.Atoi(featureStr)
		if err != nil {
			return BannerParams{}, err
		}
		res.FeatureID = &featureId
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
