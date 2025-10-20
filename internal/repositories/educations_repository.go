package repositories

import (
	"context"
	"database/sql"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
)

type EducationsRepository interface {
	ListEducations(ctx context.Context) ([]*portfolio_grpc.Education, error)
	GetEducation(ctx context.Context, id int) (*portfolio_grpc.Education, error)
}

func NewEducationsRepository(db *sql.DB, client statsd.Client) EducationsRepository {
	return newEducationsMetricsRepository(
		newEducationsDBRepository(db),
		client,
	)
}
