package postgre

import (
	"context"
	"fmt"
	"time"

	"github.com/Heatdog/Avito/internal/config"
	"github.com/Heatdog/Avito/pkg/client"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgreClient(ctx context.Context, cfg config.PostgreSettings) (client.Client, error) {
	time.Sleep(time.Duration(cfg.TimePrepare) * time.Second)
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeWait)*time.Second)
	defer cancel()

	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, err
	}

	return conn, nil
}
