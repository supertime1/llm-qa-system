// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Answer struct {
	ID            pgtype.UUID        `json:"id"`
	QuestionID    pgtype.UUID        `json:"question_id"`
	DoctorID      pgtype.UUID        `json:"doctor_id"`
	AnswerText    string             `json:"answer_text"`
	AiDraftAnswer pgtype.Text        `json:"ai_draft_answer"`
	AiConfidence  pgtype.Float8      `json:"ai_confidence"`
	AiReferences  []string           `json:"ai_references"`
	ReviewStatus  string             `json:"review_status"`
	ReviewComment pgtype.Text        `json:"review_comment"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
}

type BiometricDatum struct {
	ID         pgtype.UUID        `json:"id"`
	PatientID  pgtype.UUID        `json:"patient_id"`
	Type       string             `json:"type"`
	Value      string             `json:"value"`
	Unit       string             `json:"unit"`
	MeasuredAt pgtype.Timestamptz `json:"measured_at"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
}

type Doctor struct {
	ID         pgtype.UUID        `json:"id"`
	Name       string             `json:"name"`
	Email      string             `json:"email"`
	Department string             `json:"department"`
	CreatedAt  pgtype.Timestamptz `json:"created_at"`
}

type MedicalHistory struct {
	ID            pgtype.UUID        `json:"id"`
	PatientID     pgtype.UUID        `json:"patient_id"`
	Condition     string             `json:"condition"`
	DiagnosedDate pgtype.Timestamptz `json:"diagnosed_date"`
	Status        string             `json:"status"`
	Notes         pgtype.Text        `json:"notes"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
}

type Patient struct {
	ID        pgtype.UUID        `json:"id"`
	Name      string             `json:"name"`
	Email     string             `json:"email"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
	Age       int32              `json:"age"`
	Gender    string             `json:"gender"`
}

type Question struct {
	ID           pgtype.UUID        `json:"id"`
	PatientID    pgtype.UUID        `json:"patient_id"`
	QuestionText string             `json:"question_text"`
	QuestionType string             `json:"question_type"`
	Department   string             `json:"department"`
	UrgencyLevel pgtype.Int4        `json:"urgency_level"`
	Status       string             `json:"status"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	AnsweredAt   pgtype.Timestamptz `json:"answered_at"`
	AnsweredBy   pgtype.UUID        `json:"answered_by"`
}
