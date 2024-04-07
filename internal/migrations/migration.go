package migrations

import (
	"context"

	"github.com/Heatdog/Avito/pkg/client"
)

func initTags(client client.Client) error {
	tags := []struct {
		name string
	}{
		{
			name: "animals",
		},
		{
			name: "cars",
		},
		{
			name: "job",
		},
		{
			name: "avito",
		},
		{
			name: "moscow",
		},
	}

	q := `
		INSERT INTO tags (name)
		VALUES ($1)
	`

	for _, tag := range tags {
		client.QueryRow(context.Background(), q, tag.name)
	}
	return nil
}

func initFeatures(client client.Client) error {
	features := []struct {
		name string
	}{
		{
			name: "gettable",
		},
		{
			name: "settable",
		},
		{
			name: "readable",
		},
		{
			name: "writtable",
		},
		{
			name: "likable",
		},
	}

	q := `
		INSERT INTO features (name)
		VALUES ($1)
	`
	for _, feature := range features {
		client.QueryRow(context.Background(), q, feature.name)
	}
	return nil
}

func InitDb(client client.Client) error {
	if err := initTags(client); err != nil {
		return err
	}
	return initFeatures(client)
}
