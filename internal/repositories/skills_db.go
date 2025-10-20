package repositories

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"fmt"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/utils"
)

const (
	skillsTableName = "skills"
)

var (
	ErrSkillNotFound = errors.New("skill not found")
	//errSkillMalformed = errors.New("skill malformed")
	//errFailedToListSkills = errors.New("failed to list skills")
	//errFailedToGetSkill = errors.New("failed to get skill")
)

func newSkillsDBRepository(db *sql.DB) SkillsRepository {
	return &skillsRepositoryImpl{
		db:        db,
		tableName: skillsTableName,
	}
}

type pgSkill struct {
	ID        int64
	Title     string
	Level     int32
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

func (p *pgSkill) toProto() *portfolio_grpc.Skill {

	return &portfolio_grpc.Skill{
		Id:        p.ID,
		Title:     p.Title,
		Level:     p.Level,
		CreatedAt: utils.TimeToProtoTimestamp(p.CreatedAt),
		UpdatedAt: utils.TimeToProtoTimestamp(p.UpdatedAt),
	}
}

type skillsRepositoryImpl struct {
	db        *sql.DB
	tableName string
}

func (s *skillsRepositoryImpl) ListSkills(ctx context.Context) ([]*portfolio_grpc.Skill, error) {
	// Using constant table name is safe, but parameterized query is best practice
	query := fmt.Sprintf("SELECT id, title, level, created_at, updated_at FROM %s LIMIT 1000", s.tableName)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := make([]*portfolio_grpc.Skill, 0)
	for rows.Next() {
		skill, err := s.decodeSkill(rows)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return skills, nil
}

func (s *skillsRepositoryImpl) GetSkill(ctx context.Context, id int) (*portfolio_grpc.Skill, error) {
	// Use parameterized query to prevent SQL injection
	query := fmt.Sprintf("SELECT id, title, level, created_at, updated_at FROM %s WHERE id = $1", s.tableName)
	row := s.db.QueryRowContext(ctx, query, id)

	skill, err := s.decodeSkill(row)
	if err != nil {
		return nil, err
	}

	return skill, nil
}

func (s *skillsRepositoryImpl) decodeSkill(row rowScanner) (*portfolio_grpc.Skill, error) {
	skill := new(pgSkill)
	err := row.Scan(&skill.ID, &skill.Title, &skill.Level, &skill.CreatedAt, &skill.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSkillNotFound
		}
		return nil, fmt.Errorf("failed to scan skill: %w", err)
	}

	return skill.toProto(), nil
}
