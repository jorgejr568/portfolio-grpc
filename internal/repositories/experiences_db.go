package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/utils"
)

const (
	experiencesTableName = "experiences"
)

var (
	ErrExperienceNotFound = errors.New("experience not found")
)

func newExperiencesDBRepository(db *sql.DB) ExperiencesRepository {
	return &experiencesRepositoryImpl{
		db:        db,
		tableName: experiencesTableName,
		selectColumns: strings.Join([]string{
			"id",
			"title",
			"description",
			"company_name",
			"company_url",
			"company_logo_url",
			"languages",
			"frameworks",
			"started_at",
			"finished_at as ended_at",
			"created_at",
			"updated_at",
		}, ","),
	}
}

type pgExperience struct {
	ID          int64
	Title       string
	Description string
	CompanyName string
	CompanyURL  string
	CompanyLogo string
	Languages   string
	Frameworks  string
	StartedAt   *time.Time
	EndedAt     *time.Time
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
}

func (p *pgExperience) toProto() *portfolio_grpc.Experience {
	exp := &portfolio_grpc.Experience{
		Id:          p.ID,
		Title:       p.Title,
		Description: p.Description,
		CreatedAt:   utils.TimeToProtoTimestamp(p.CreatedAt),
		UpdatedAt:   utils.TimeToProtoTimestamp(p.UpdatedAt),
	}

	exp.Company = &portfolio_grpc.Experience_Company{
		Name:    p.CompanyName,
		Url:     p.CompanyURL,
		LogoUrl: p.CompanyLogo,
	}

	var langs, frameworks []string
	_ = json.Unmarshal([]byte(p.Languages), &langs)
	_ = json.Unmarshal([]byte(p.Frameworks), &frameworks)

	exp.Technologies = append(exp.Technologies, langs...)

	exp.StartedAt = utils.TimeToProtoDate(p.StartedAt)
	exp.EndedAt = utils.TimeToProtoDate(p.EndedAt)
	return exp
}

type experiencesRepositoryImpl struct {
	db            *sql.DB
	tableName     string
	selectColumns string
}

func (e *experiencesRepositoryImpl) ListExperiences(ctx context.Context) ([]*portfolio_grpc.Experience, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s e
		ORDER BY e.started_at DESC
		LIMIT 1000`, e.selectColumns, e.tableName)

	rows, err := e.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	experiences := make([]*portfolio_grpc.Experience, 0)
	for rows.Next() {
		experience, err := e.decodeExperience(rows)
		if err != nil {
			return nil, err
		}
		experiences = append(experiences, experience)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return experiences, nil
}

func (e *experiencesRepositoryImpl) GetExperience(ctx context.Context, id int) (*portfolio_grpc.Experience, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE id = $1 LIMIT 1;`, e.selectColumns, e.tableName)

	row := e.db.QueryRowContext(ctx, query, id)

	experience, err := e.decodeExperience(row)
	if err != nil {
		return nil, err
	}

	return experience, nil
}

func (e *experiencesRepositoryImpl) decodeExperience(row rowScanner) (*portfolio_grpc.Experience, error) {
	exp := new(pgExperience)
	err := row.Scan(
		&exp.ID,
		&exp.Title,
		&exp.Description,
		&exp.CompanyName,
		&exp.CompanyURL,
		&exp.CompanyLogo,
		&exp.Languages,
		&exp.Frameworks,
		&exp.StartedAt,
		&exp.EndedAt,
		&exp.CreatedAt,
		&exp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrExperienceNotFound
		}
		return nil, fmt.Errorf("failed to scan experience: %w", err)
	}

	return exp.toProto(), nil
}
