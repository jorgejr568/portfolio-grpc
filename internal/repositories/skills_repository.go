package repositories

import (
	"context"
	"database/sql"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/client/statsd"
)

type SkillsRepository interface {
	ListSkills(ctx context.Context) ([]*portfolio_grpc.Skill, error)
	GetSkill(ctx context.Context, id int) (*portfolio_grpc.Skill, error)
}

func NewSkillsRepository(db *sql.DB, client statsd.Client) SkillsRepository {
	return newSkillsMetricsRepository(
		newSkillsDBRepository(db),
		client,
	)
}
