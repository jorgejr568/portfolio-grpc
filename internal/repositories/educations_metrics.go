package repositories

import (
	"context"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
)

type educationsMetricsRepositoryImpl struct {
	repo   EducationsRepository
	statsd statsd.Client
}

func (e *educationsMetricsRepositoryImpl) ListEducations(ctx context.Context) ([]*portfolio_grpc.Education, error) {
	stat := e.statsd.Start("educations", "ListEducations")
	defer stat.Finished()

	educations, err := e.repo.ListEducations(ctx)
	if err != nil {
		stat.FailedWithError(err)
		return nil, err
	}

	stat.Succeeded()
	return educations, nil
}

func (e *educationsMetricsRepositoryImpl) GetEducation(ctx context.Context, id int) (*portfolio_grpc.Education, error) {
	stat := e.statsd.Start("educations", "GetEducation")
	defer stat.Finished()

	education, err := e.repo.GetEducation(ctx, id)
	if err != nil {
		stat.FailedWithError(err)
		return nil, err
	}

	stat.Succeeded()
	return education, nil
}

func newEducationsMetricsRepository(repo EducationsRepository, statsdClient statsd.Client) EducationsRepository {
	return &educationsMetricsRepositoryImpl{
		repo:   repo,
		statsd: statsdClient,
	}
}
