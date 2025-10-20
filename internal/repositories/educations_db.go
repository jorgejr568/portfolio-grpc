package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/utils"
)

const (
	educationsTableName = "education"
)

var (
	ErrEducationNotFound = errors.New("education not found")
)

func newEducationsDBRepository(db *sql.DB) EducationsRepository {
	return &educationsRepositoryImpl{
		db:        db,
		tableName: educationsTableName,
		selectColumns: strings.Join([]string{
			"id",
			"title",
			"institution_name",
			"institution_url",
			"started_at",
			"finished_at as ended_at",
			"created_at",
			"updated_at",
		}, ","),
	}
}

type pgEducation struct {
	ID              int64
	Title           string
	Description     string
	InstitutionName string
	InstitutionURL  string
	StartedAt       *time.Time
	EndedAt         *time.Time
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
}

func (p *pgEducation) toProto() *portfolio_grpc.Education {
	edu := &portfolio_grpc.Education{
		Id:        p.ID,
		Title:     p.Title,
		CreatedAt: utils.TimeToProtoTimestamp(p.CreatedAt),
		UpdatedAt: utils.TimeToProtoTimestamp(p.UpdatedAt),
		StartedAt: utils.TimeToProtoDate(p.StartedAt),
		EndedAt:   utils.TimeToProtoDate(p.EndedAt),
	}

	edu.Institution = &portfolio_grpc.Education_Institution{
		Name: p.InstitutionName,
		Url:  p.InstitutionURL,
	}

	if p.StartedAt != nil {
		edu.StartedAt = utils.TimeToProtoDate(p.StartedAt)
	}

	if p.EndedAt != nil {
		edu.EndedAt = utils.TimeToProtoDate(p.EndedAt)
	}

	return edu
}

type educationsRepositoryImpl struct {
	db            *sql.DB
	tableName     string
	selectColumns string
}

func (e *educationsRepositoryImpl) ListEducations(ctx context.Context) ([]*portfolio_grpc.Education, error) {
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

	educations := make([]*portfolio_grpc.Education, 0)
	for rows.Next() {
		education, err := e.decodeEducation(rows)
		if err != nil {
			return nil, err
		}
		educations = append(educations, education)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return educations, nil
}

func (e *educationsRepositoryImpl) GetEducation(ctx context.Context, id int) (*portfolio_grpc.Education, error) {
	query := fmt.Sprintf(`
		SELECT %s
		FROM %s
		WHERE id = $1`, e.selectColumns, e.tableName)

	row := e.db.QueryRowContext(ctx, query, id)

	education, err := e.decodeEducation(row)
	if err != nil {
		return nil, err
	}

	return education, nil
}

func (e *educationsRepositoryImpl) decodeEducation(row rowScanner) (*portfolio_grpc.Education, error) {
	edu := new(pgEducation)
	err := row.Scan(
		&edu.ID,
		&edu.Title,
		&edu.InstitutionName,
		&edu.InstitutionURL,
		&edu.StartedAt,
		&edu.EndedAt,
		&edu.CreatedAt,
		&edu.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEducationNotFound
		}
		return nil, fmt.Errorf("failed to scan education: %w", err)
	}

	return edu.toProto(), nil
}
