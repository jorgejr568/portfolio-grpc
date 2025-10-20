package repositories

import (
	"context"
	"database/sql"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
)

type ExperiencesRepository interface {
	ListExperiences(ctx context.Context) ([]*portfolio_grpc.Experience, error)
	GetExperience(ctx context.Context, id int) (*portfolio_grpc.Experience, error)
}

func NewExperiencesRepository(db *sql.DB, client statsd.Client) ExperiencesRepository {
	return newExperiencesMetricsRepository(
		newExperiencesDBRepository(db),
		client,
	)
}
