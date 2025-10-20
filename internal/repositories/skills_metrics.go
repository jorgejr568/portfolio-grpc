package repositories

import (
	"context"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
)

type skillsMetricsRepositoryImpl struct {
	repo   SkillsRepository
	statsd statsd.Client
}

func (s *skillsMetricsRepositoryImpl) ListSkills(ctx context.Context) ([]*portfolio_grpc.Skill, error) {
	stat := s.statsd.Start("skills", "ListSkills")
	defer stat.Finished()

	skills, err := s.repo.ListSkills(ctx)
	if err != nil {
		stat.FailedWithError(err)
		return nil, err
	}

	stat.Succeeded()
	return skills, nil
}

func (s *skillsMetricsRepositoryImpl) GetSkill(ctx context.Context, id int) (*portfolio_grpc.Skill, error) {
	stat := s.statsd.Start("skills", "GetSkill")
	defer stat.Finished()

	skill, err := s.repo.GetSkill(ctx, id)
	if err != nil {
		stat.FailedWithError(err)
		return nil, err
	}

	stat.Succeeded()
	return skill, nil
}

func newSkillsMetricsRepository(repo SkillsRepository, statsdClient statsd.Client) SkillsRepository {
	return &skillsMetricsRepositoryImpl{
		repo:   repo,
		statsd: statsdClient,
	}
}
