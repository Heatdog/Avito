package migrations

import (
	"context"

	"github.com/Heatdog/Avito/pkg/client/postgre"
)

func initTags(client postgre.Client) error {
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

	var id int
	for _, tag := range tags {
		row := client.QueryRow(context.Background(), q, tag.name)
		if err := row.Scan(&id); err != nil {
			return err
		}
	}
	return nil
}

func initFeatures(client postgre.Client) error {
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
	var id int
	for _, feature := range features {
		row := client.QueryRow(context.Background(), q, feature.name)
		if err := row.Scan(&id); err != nil {
			return err
		}
	}
	return nil
}

func InitDb(client postgre.Client) error {
	if err := initTags(client); err != nil {
		return err
	}
	return initFeatures(client)
}
