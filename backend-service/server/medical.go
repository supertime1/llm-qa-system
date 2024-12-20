package server

import (
	"context"
	"fmt"
	"strconv"

	db "llm-qa-system/backend-service/src/db"
	pb "llm-qa-system/backend-service/src/proto"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"google.golang.org/grpc"
)

type MedicalServer struct {
	pb.UnimplementedMedicalServiceServer
	*BaseServer
	llmClient pb.MedicalQAServiceClient
}

func NewMedicalServer(pool *pgxpool.Pool, llmServiceAddr string) (*MedicalServer, error) {
	conn, err := grpc.Dial(llmServiceAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LLM service: %v", err)
	}

	return &MedicalServer{
		BaseServer: NewBaseServer(pool),
		llmClient:  pb.NewMedicalQAServiceClient(conn),
	}, nil
}

// SubmitQuestion handles the initial question submission from a patient
func (s *MedicalServer) SubmitQuestion(ctx context.Context, req *pb.PatientQuestionRequest) (*pb.PatientQuestionResponse, error) {
	// Create question in database
	patientUUID, err := uuid.Parse(req.PatientId)
	if err != nil {
		return nil, fmt.Errorf("invalid patient ID: %v", err)
	}

	question, err := s.dbq.CreateQuestion(ctx, db.CreateQuestionParams{
		PatientID:    pgtype.UUID{Bytes: patientUUID[:], Valid: true},
		QuestionText: req.QuestionText,
		QuestionType: req.QuestionType.String(),
		Department:   req.Department.String(),
		UrgencyLevel: pgtype.Int4{Int32: int32(req.UrgencyLevel), Valid: true},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create question: %v", err)
	}

	// Get user context from database by patient ID
	userContext, err := s.dbq.GetPatientWithContext(ctx, patientUUID)

	if err != nil {
		return nil, fmt.Errorf("failed to get user context: %v", err)
	}

	questionUUID, err := uuid.FromBytes(question.ID.Bytes[:])
	if err != nil {
		return nil, fmt.Errorf("failed to convert UUID: %v", err)
	}
	// Send to LLM service for draft answer
	llmResp, err := s.llmClient.GenerateDraftAnswer(ctx, &pb.QuestionRequest{
		QuestionId:   questionUUID.String(),
		QuestionText: req.QuestionText,
		UserContext: &pb.UserContext{
			UserInfo: &pb.UserInfo{
				Age:    strconv.Itoa(int(userContext.Age)),
				Gender: userContext.Gender,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to generate draft answer: %v", err)
	}

	// Save AI draft answer
	_, err = s.dbq.SaveAIDraftAnswer(ctx, db.SaveAIDraftAnswerParams{
		QuestionID:    question.ID,
		AiDraftAnswer: pgtype.Text{String: llmResp.DraftAnswer, Valid: true},
		AiConfidence:  pgtype.Float8{Float64: float64(llmResp.ConfidenceScore), Valid: true},
		AiReferences:  llmResp.References,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to save draft answer: %v", err)
	}

	// Update question status to PENDING_REVIEW
	err = s.dbq.UpdateQuestionStatus(ctx, db.UpdateQuestionStatusParams{
		ID:     question.ID,
		Status: "STATUS_PENDING_REVIEW",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update question status: %v", err)
	}

	return &pb.PatientQuestionResponse{
		QuestionId: questionUUID.String(),
		Status:     pb.QuestionStatus_STATUS_PENDING_REVIEW,
	}, nil

}
