package repositories

import (
	"context"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
)

type experiencesMetricsRepositoryImpl struct {
	repo   ExperiencesRepository
	statsd statsd.Client
}

func (e *experiencesMetricsRepositoryImpl) ListExperiences(ctx context.Context) ([]*portfolio_grpc.Experience, error) {
	stat := e.statsd.Start("experiences", "ListExperiences")
	defer stat.Finished()

	experiences, err := e.repo.ListExperiences(ctx)
	if err != nil {
		stat.FailedWithError(err)
		return nil, err
	}

	stat.Succeeded()
	return experiences, nil
}

func (e *experiencesMetricsRepositoryImpl) GetExperience(ctx context.Context, id int) (*portfolio_grpc.Experience, error) {
	stat := e.statsd.Start("experiences", "GetExperience")
	defer stat.Finished()

	experience, err := e.repo.GetExperience(ctx, id)
	if err != nil {
		stat.FailedWithError(err)
		return nil, err
	}

	stat.Succeeded()
	return experience, nil
}

func newExperiencesMetricsRepository(repo ExperiencesRepository, statsdClient statsd.Client) ExperiencesRepository {
	return &experiencesMetricsRepositoryImpl{
		repo:   repo,
		statsd: statsdClient,
	}
}
