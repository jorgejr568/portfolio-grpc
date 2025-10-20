package server

import (
	"context"
	"errors"

	"github.com/jorgejr568/portfolio-grpc/gen/go/jorgejr568/portfolio_grpc"
	"github.com/jorgejr568/portfolio-grpc/internal/repositories"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server interface {
	portfolio_grpc.PortfolioServiceServer
}

func NewServer(
	skillsRepository repositories.SkillsRepository,
	experiencesRepository repositories.ExperiencesRepository,
	educationsRepository repositories.EducationsRepository,
) Server {
	return &serverImpl{
		skillsRepository:      skillsRepository,
		experiencesRepository: experiencesRepository,
		educationsRepository:  educationsRepository,
	}
}

type serverImpl struct {
	portfolio_grpc.UnimplementedPortfolioServiceServer

	skillsRepository      repositories.SkillsRepository
	experiencesRepository repositories.ExperiencesRepository
	educationsRepository  repositories.EducationsRepository
}

func (s *serverImpl) GetAllSkills(ctx context.Context, request *portfolio_grpc.GetAllSkillsRequest) (*portfolio_grpc.GetAllSkillsResponse, error) {
	skills, err := s.skillsRepository.ListSkills(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &portfolio_grpc.GetAllSkillsResponse{
		Skills: skills,
	}, nil
}

func (s *serverImpl) GetSkill(ctx context.Context, request *portfolio_grpc.GetSkillRequest) (*portfolio_grpc.GetSkillResponse, error) {
	skill, err := s.skillsRepository.GetSkill(ctx, int(request.Id))
	if err != nil {
		if errors.Is(err, repositories.ErrSkillNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &portfolio_grpc.GetSkillResponse{
		Skill: skill,
	}, nil
}

func (s *serverImpl) GetAllExperiences(ctx context.Context, request *portfolio_grpc.GetAllExperiencesRequest) (*portfolio_grpc.GetAllExperiencesResponse, error) {
	experiences, err := s.experiencesRepository.ListExperiences(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &portfolio_grpc.GetAllExperiencesResponse{
		Experiences: experiences,
	}, nil
}

func (s *serverImpl) GetExperience(ctx context.Context, request *portfolio_grpc.GetExperienceRequest) (*portfolio_grpc.GetExperienceResponse, error) {
	experience, err := s.experiencesRepository.GetExperience(ctx, int(request.Id))
	if err != nil {
		if errors.Is(err, repositories.ErrExperienceNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &portfolio_grpc.GetExperienceResponse{
		Experience: experience,
	}, nil
}

func (s *serverImpl) GetAllEducations(ctx context.Context, request *portfolio_grpc.GetAllEducationsRequest) (*portfolio_grpc.GetAllEducationsResponse, error) {
	educations, err := s.educationsRepository.ListEducations(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &portfolio_grpc.GetAllEducationsResponse{
		Educations: educations,
	}, nil
}

func (s *serverImpl) GetEducation(ctx context.Context, request *portfolio_grpc.GetEducationRequest) (*portfolio_grpc.GetEducationResponse, error) {
	education, err := s.educationsRepository.GetEducation(ctx, int(request.Id))
	if err != nil {
		if errors.Is(err, repositories.ErrEducationNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}

		return nil, status.Error(codes.Internal, err.Error())
	}

	return &portfolio_grpc.GetEducationResponse{
		Education: education,
	}, nil
}
