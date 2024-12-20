// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: queries.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createBiometricData = `-- name: CreateBiometricData :one
INSERT INTO biometric_data (
    patient_id,
    type,
    value,
    unit,
    measured_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, patient_id, type, value, unit, measured_at, created_at
`

type CreateBiometricDataParams struct {
	PatientID  pgtype.UUID        `json:"patient_id"`
	Type       string             `json:"type"`
	Value      string             `json:"value"`
	Unit       string             `json:"unit"`
	MeasuredAt pgtype.Timestamptz `json:"measured_at"`
}

// Biometric Data Operations
func (q *Queries) CreateBiometricData(ctx context.Context, arg CreateBiometricDataParams) (BiometricDatum, error) {
	row := q.db.QueryRow(ctx, createBiometricData,
		arg.PatientID,
		arg.Type,
		arg.Value,
		arg.Unit,
		arg.MeasuredAt,
	)
	var i BiometricDatum
	err := row.Scan(
		&i.ID,
		&i.PatientID,
		&i.Type,
		&i.Value,
		&i.Unit,
		&i.MeasuredAt,
		&i.CreatedAt,
	)
	return i, err
}

const createMedicalHistory = `-- name: CreateMedicalHistory :one
INSERT INTO medical_history (
    patient_id,
    condition,
    diagnosed_date,
    status,
    notes
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING id, patient_id, condition, diagnosed_date, status, notes, created_at
`

type CreateMedicalHistoryParams struct {
	PatientID     pgtype.UUID        `json:"patient_id"`
	Condition     string             `json:"condition"`
	DiagnosedDate pgtype.Timestamptz `json:"diagnosed_date"`
	Status        string             `json:"status"`
	Notes         pgtype.Text        `json:"notes"`
}

// Medical History Operations
func (q *Queries) CreateMedicalHistory(ctx context.Context, arg CreateMedicalHistoryParams) (MedicalHistory, error) {
	row := q.db.QueryRow(ctx, createMedicalHistory,
		arg.PatientID,
		arg.Condition,
		arg.DiagnosedDate,
		arg.Status,
		arg.Notes,
	)
	var i MedicalHistory
	err := row.Scan(
		&i.ID,
		&i.PatientID,
		&i.Condition,
		&i.DiagnosedDate,
		&i.Status,
		&i.Notes,
		&i.CreatedAt,
	)
	return i, err
}

const createQuestion = `-- name: CreateQuestion :one
INSERT INTO questions (
    patient_id,
    question_text,
    question_type,
    department,
    urgency_level,
    status
) VALUES (
    $1, $2, $3, $4, $5, 'STATUS_PENDING'
) RETURNING id, patient_id, question_text, question_type, department, urgency_level, status, created_at, answered_at, answered_by
`

type CreateQuestionParams struct {
	PatientID    pgtype.UUID `json:"patient_id"`
	QuestionText string      `json:"question_text"`
	QuestionType string      `json:"question_type"`
	Department   string      `json:"department"`
	UrgencyLevel pgtype.Int4 `json:"urgency_level"`
}

// Patient operations
func (q *Queries) CreateQuestion(ctx context.Context, arg CreateQuestionParams) (Question, error) {
	row := q.db.QueryRow(ctx, createQuestion,
		arg.PatientID,
		arg.QuestionText,
		arg.QuestionType,
		arg.Department,
		arg.UrgencyLevel,
	)
	var i Question
	err := row.Scan(
		&i.ID,
		&i.PatientID,
		&i.QuestionText,
		&i.QuestionType,
		&i.Department,
		&i.UrgencyLevel,
		&i.Status,
		&i.CreatedAt,
		&i.AnsweredAt,
		&i.AnsweredBy,
	)
	return i, err
}

const getActiveMedicalConditions = `-- name: GetActiveMedicalConditions :many
SELECT id, patient_id, condition, diagnosed_date, status, notes, created_at FROM medical_history
WHERE patient_id = $1
    AND status IN ('ACTIVE', 'CHRONIC')
ORDER BY diagnosed_date DESC
`

func (q *Queries) GetActiveMedicalConditions(ctx context.Context, patientID pgtype.UUID) ([]MedicalHistory, error) {
	rows, err := q.db.Query(ctx, getActiveMedicalConditions, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MedicalHistory{}
	for rows.Next() {
		var i MedicalHistory
		if err := rows.Scan(
			&i.ID,
			&i.PatientID,
			&i.Condition,
			&i.DiagnosedDate,
			&i.Status,
			&i.Notes,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAnswerHistory = `-- name: GetAnswerHistory :many
SELECT 
    q.id as question_id,
    q.question_text,
    q.question_type,
    q.department,
    a.answer_text as answer,
    q.status,
    q.created_at,
    a.created_at as answered_at,
    d.name as answered_by
FROM questions q
LEFT JOIN answers a ON q.id = a.question_id
LEFT JOIN doctors d ON a.doctor_id = d.id
WHERE q.patient_id = $1
    AND ($2::question_status IS NULL OR q.status = $2)
    AND ($3::timestamptz IS NULL OR q.created_at >= $3)
    AND ($4::timestamptz IS NULL OR q.created_at <= $4)
ORDER BY q.created_at DESC
LIMIT $5 OFFSET $6
`

type GetAnswerHistoryParams struct {
	PatientID pgtype.UUID        `json:"patient_id"`
	Column2   interface{}        `json:"column_2"`
	Column3   pgtype.Timestamptz `json:"column_3"`
	Column4   pgtype.Timestamptz `json:"column_4"`
	Limit     int32              `json:"limit"`
	Offset    int32              `json:"offset"`
}

type GetAnswerHistoryRow struct {
	QuestionID   pgtype.UUID        `json:"question_id"`
	QuestionText string             `json:"question_text"`
	QuestionType string             `json:"question_type"`
	Department   string             `json:"department"`
	Answer       pgtype.Text        `json:"answer"`
	Status       string             `json:"status"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	AnsweredAt   pgtype.Timestamptz `json:"answered_at"`
	AnsweredBy   pgtype.Text        `json:"answered_by"`
}

func (q *Queries) GetAnswerHistory(ctx context.Context, arg GetAnswerHistoryParams) ([]GetAnswerHistoryRow, error) {
	rows, err := q.db.Query(ctx, getAnswerHistory,
		arg.PatientID,
		arg.Column2,
		arg.Column3,
		arg.Column4,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAnswerHistoryRow{}
	for rows.Next() {
		var i GetAnswerHistoryRow
		if err := rows.Scan(
			&i.QuestionID,
			&i.QuestionText,
			&i.QuestionType,
			&i.Department,
			&i.Answer,
			&i.Status,
			&i.CreatedAt,
			&i.AnsweredAt,
			&i.AnsweredBy,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAnswerHistoryCount = `-- name: GetAnswerHistoryCount :one
SELECT COUNT(*)
FROM questions
WHERE patient_id = $1
    AND ($2::question_status IS NULL OR status = $2)
    AND ($3::timestamptz IS NULL OR created_at >= $3)
    AND ($4::timestamptz IS NULL OR created_at <= $4)
`

type GetAnswerHistoryCountParams struct {
	PatientID pgtype.UUID        `json:"patient_id"`
	Column2   interface{}        `json:"column_2"`
	Column3   pgtype.Timestamptz `json:"column_3"`
	Column4   pgtype.Timestamptz `json:"column_4"`
}

func (q *Queries) GetAnswerHistoryCount(ctx context.Context, arg GetAnswerHistoryCountParams) (int64, error) {
	row := q.db.QueryRow(ctx, getAnswerHistoryCount,
		arg.PatientID,
		arg.Column2,
		arg.Column3,
		arg.Column4,
	)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getLatestBiometricsByType = `-- name: GetLatestBiometricsByType :many
SELECT DISTINCT ON (type) id, patient_id, type, value, unit, measured_at, created_at
FROM biometric_data
WHERE patient_id = $1
ORDER BY type, measured_at DESC
`

func (q *Queries) GetLatestBiometricsByType(ctx context.Context, patientID pgtype.UUID) ([]BiometricDatum, error) {
	rows, err := q.db.Query(ctx, getLatestBiometricsByType, patientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []BiometricDatum{}
	for rows.Next() {
		var i BiometricDatum
		if err := rows.Scan(
			&i.ID,
			&i.PatientID,
			&i.Type,
			&i.Value,
			&i.Unit,
			&i.MeasuredAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPatientBiometricData = `-- name: GetPatientBiometricData :many
SELECT id, patient_id, type, value, unit, measured_at, created_at FROM biometric_data
WHERE patient_id = $1
    AND ($2::varchar IS NULL OR type = $2)
    AND ($3::timestamptz IS NULL OR measured_at >= $3)
    AND ($4::timestamptz IS NULL OR measured_at <= $4)
ORDER BY measured_at DESC
LIMIT $5 OFFSET $6
`

type GetPatientBiometricDataParams struct {
	PatientID pgtype.UUID        `json:"patient_id"`
	Column2   string             `json:"column_2"`
	Column3   pgtype.Timestamptz `json:"column_3"`
	Column4   pgtype.Timestamptz `json:"column_4"`
	Limit     int32              `json:"limit"`
	Offset    int32              `json:"offset"`
}

func (q *Queries) GetPatientBiometricData(ctx context.Context, arg GetPatientBiometricDataParams) ([]BiometricDatum, error) {
	rows, err := q.db.Query(ctx, getPatientBiometricData,
		arg.PatientID,
		arg.Column2,
		arg.Column3,
		arg.Column4,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []BiometricDatum{}
	for rows.Next() {
		var i BiometricDatum
		if err := rows.Scan(
			&i.ID,
			&i.PatientID,
			&i.Type,
			&i.Value,
			&i.Unit,
			&i.MeasuredAt,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPatientMedicalHistory = `-- name: GetPatientMedicalHistory :many
SELECT id, patient_id, condition, diagnosed_date, status, notes, created_at FROM medical_history
WHERE patient_id = $1
    AND ($2::varchar IS NULL OR status = $2)
ORDER BY diagnosed_date DESC
LIMIT $3 OFFSET $4
`

type GetPatientMedicalHistoryParams struct {
	PatientID pgtype.UUID `json:"patient_id"`
	Column2   string      `json:"column_2"`
	Limit     int32       `json:"limit"`
	Offset    int32       `json:"offset"`
}

func (q *Queries) GetPatientMedicalHistory(ctx context.Context, arg GetPatientMedicalHistoryParams) ([]MedicalHistory, error) {
	rows, err := q.db.Query(ctx, getPatientMedicalHistory,
		arg.PatientID,
		arg.Column2,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []MedicalHistory{}
	for rows.Next() {
		var i MedicalHistory
		if err := rows.Scan(
			&i.ID,
			&i.PatientID,
			&i.Condition,
			&i.DiagnosedDate,
			&i.Status,
			&i.Notes,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPatientWithContext = `-- name: GetPatientWithContext :one
SELECT 
    p.id, p.name, p.email, p.created_at, p.age, p.gender,
    json_agg(DISTINCT jsonb_build_object(
        'type', b.type,
        'value', b.value,
        'unit', b.unit,
        'measured_at', b.measured_at
    )) FILTER (WHERE b.id IS NOT NULL) as biometric_data,
    json_agg(DISTINCT jsonb_build_object(
        'condition', m.condition,
        'status', m.status,
        'diagnosed_date', m.diagnosed_date
    )) FILTER (WHERE m.id IS NOT NULL) as medical_history
FROM patients p
LEFT JOIN biometric_data b ON b.patient_id = p.id
LEFT JOIN medical_history m ON m.patient_id = p.id
WHERE p.id = $1
GROUP BY p.id
`

type GetPatientWithContextRow struct {
	ID             pgtype.UUID        `json:"id"`
	Name           string             `json:"name"`
	Email          string             `json:"email"`
	CreatedAt      pgtype.Timestamptz `json:"created_at"`
	Age            int32              `json:"age"`
	Gender         string             `json:"gender"`
	BiometricData  []byte             `json:"biometric_data"`
	MedicalHistory []byte             `json:"medical_history"`
}

// Patient Context Operations
func (q *Queries) GetPatientWithContext(ctx context.Context, id pgtype.UUID) (GetPatientWithContextRow, error) {
	row := q.db.QueryRow(ctx, getPatientWithContext, id)
	var i GetPatientWithContextRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.CreatedAt,
		&i.Age,
		&i.Gender,
		&i.BiometricData,
		&i.MedicalHistory,
	)
	return i, err
}

const getPendingReviews = `-- name: GetPendingReviews :many
SELECT 
    q.id as question_id,
    q.patient_id,
    q.question_text,
    q.question_type,
    q.department,
    q.urgency_level,
    a.ai_draft_answer,
    a.ai_confidence,
    a.ai_references,
    q.created_at
FROM questions q
LEFT JOIN answers a ON q.id = a.question_id
WHERE q.status = 'STATUS_PENDING_REVIEW'
    AND ($1::department IS NULL OR q.department = $1)
    AND ($2::question_type IS NULL OR q.question_type = $2)
    AND ($3::int IS NULL OR q.urgency_level >= $3)
    AND ($4::timestamptz IS NULL OR q.created_at >= $4)
    AND ($5::timestamptz IS NULL OR q.created_at <= $5)
ORDER BY q.urgency_level DESC, q.created_at ASC
LIMIT $6 OFFSET $7
`

type GetPendingReviewsParams struct {
	Column1 interface{}        `json:"column_1"`
	Column2 interface{}        `json:"column_2"`
	Column3 int32              `json:"column_3"`
	Column4 pgtype.Timestamptz `json:"column_4"`
	Column5 pgtype.Timestamptz `json:"column_5"`
	Limit   int32              `json:"limit"`
	Offset  int32              `json:"offset"`
}

type GetPendingReviewsRow struct {
	QuestionID    pgtype.UUID        `json:"question_id"`
	PatientID     pgtype.UUID        `json:"patient_id"`
	QuestionText  string             `json:"question_text"`
	QuestionType  string             `json:"question_type"`
	Department    string             `json:"department"`
	UrgencyLevel  pgtype.Int4        `json:"urgency_level"`
	AiDraftAnswer pgtype.Text        `json:"ai_draft_answer"`
	AiConfidence  pgtype.Float8      `json:"ai_confidence"`
	AiReferences  []string           `json:"ai_references"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
}

// Doctor operations
func (q *Queries) GetPendingReviews(ctx context.Context, arg GetPendingReviewsParams) ([]GetPendingReviewsRow, error) {
	rows, err := q.db.Query(ctx, getPendingReviews,
		arg.Column1,
		arg.Column2,
		arg.Column3,
		arg.Column4,
		arg.Column5,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetPendingReviewsRow{}
	for rows.Next() {
		var i GetPendingReviewsRow
		if err := rows.Scan(
			&i.QuestionID,
			&i.PatientID,
			&i.QuestionText,
			&i.QuestionType,
			&i.Department,
			&i.UrgencyLevel,
			&i.AiDraftAnswer,
			&i.AiConfidence,
			&i.AiReferences,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPendingReviewsCount = `-- name: GetPendingReviewsCount :one
SELECT COUNT(*)
FROM questions
WHERE status = 'STATUS_PENDING_REVIEW'
    AND ($1::department IS NULL OR department = $1)
    AND ($2::question_type IS NULL OR question_type = $2)
    AND ($3::int IS NULL OR urgency_level >= $3)
    AND ($4::timestamptz IS NULL OR created_at >= $4)
    AND ($5::timestamptz IS NULL OR created_at <= $5)
`

type GetPendingReviewsCountParams struct {
	Column1 interface{}        `json:"column_1"`
	Column2 interface{}        `json:"column_2"`
	Column3 int32              `json:"column_3"`
	Column4 pgtype.Timestamptz `json:"column_4"`
	Column5 pgtype.Timestamptz `json:"column_5"`
}

func (q *Queries) GetPendingReviewsCount(ctx context.Context, arg GetPendingReviewsCountParams) (int64, error) {
	row := q.db.QueryRow(ctx, getPendingReviewsCount,
		arg.Column1,
		arg.Column2,
		arg.Column3,
		arg.Column4,
		arg.Column5,
	)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getQuestionStatus = `-- name: GetQuestionStatus :one
SELECT 
    q.id, q.patient_id, q.question_text, q.question_type, q.department, q.urgency_level, q.status, q.created_at, q.answered_at, q.answered_by,
    a.answer_text,
    a.doctor_id as answered_by
FROM questions q
LEFT JOIN answers a ON q.id = a.question_id
WHERE q.id = $1 AND q.patient_id = $2
LIMIT 1
`

type GetQuestionStatusParams struct {
	ID        pgtype.UUID `json:"id"`
	PatientID pgtype.UUID `json:"patient_id"`
}

type GetQuestionStatusRow struct {
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
	AnswerText   pgtype.Text        `json:"answer_text"`
	AnsweredBy_2 pgtype.UUID        `json:"answered_by_2"`
}

func (q *Queries) GetQuestionStatus(ctx context.Context, arg GetQuestionStatusParams) (GetQuestionStatusRow, error) {
	row := q.db.QueryRow(ctx, getQuestionStatus, arg.ID, arg.PatientID)
	var i GetQuestionStatusRow
	err := row.Scan(
		&i.ID,
		&i.PatientID,
		&i.QuestionText,
		&i.QuestionType,
		&i.Department,
		&i.UrgencyLevel,
		&i.Status,
		&i.CreatedAt,
		&i.AnsweredAt,
		&i.AnsweredBy,
		&i.AnswerText,
		&i.AnsweredBy_2,
	)
	return i, err
}

const saveAIDraftAnswer = `-- name: SaveAIDraftAnswer :one
INSERT INTO answers (
    question_id,
    ai_draft_answer,
    ai_confidence,
    ai_references,
    review_status
) VALUES (
    $1, $2, $3, $4, 'DECISION_UNSPECIFIED'
)
RETURNING id, question_id, doctor_id, answer_text, ai_draft_answer, ai_confidence, ai_references, review_status, review_comment, created_at, updated_at
`

type SaveAIDraftAnswerParams struct {
	QuestionID    pgtype.UUID   `json:"question_id"`
	AiDraftAnswer pgtype.Text   `json:"ai_draft_answer"`
	AiConfidence  pgtype.Float8 `json:"ai_confidence"`
	AiReferences  []string      `json:"ai_references"`
}

func (q *Queries) SaveAIDraftAnswer(ctx context.Context, arg SaveAIDraftAnswerParams) (Answer, error) {
	row := q.db.QueryRow(ctx, saveAIDraftAnswer,
		arg.QuestionID,
		arg.AiDraftAnswer,
		arg.AiConfidence,
		arg.AiReferences,
	)
	var i Answer
	err := row.Scan(
		&i.ID,
		&i.QuestionID,
		&i.DoctorID,
		&i.AnswerText,
		&i.AiDraftAnswer,
		&i.AiConfidence,
		&i.AiReferences,
		&i.ReviewStatus,
		&i.ReviewComment,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const submitReview = `-- name: SubmitReview :one
WITH updated_question AS (
    UPDATE questions
    SET status = $2
    WHERE id = $1
    RETURNING id
)
INSERT INTO answers (
    question_id,
    doctor_id,
    answer_text,
    ai_draft_answer,
    review_status,
    review_comment
) VALUES (
    $1, $3, $4, $5, $6, $7
)
RETURNING id, question_id, doctor_id, answer_text, ai_draft_answer, ai_confidence, ai_references, review_status, review_comment, created_at, updated_at
`

type SubmitReviewParams struct {
	QuestionID    pgtype.UUID `json:"question_id"`
	Status        string      `json:"status"`
	DoctorID      pgtype.UUID `json:"doctor_id"`
	AnswerText    string      `json:"answer_text"`
	AiDraftAnswer pgtype.Text `json:"ai_draft_answer"`
	ReviewStatus  string      `json:"review_status"`
	ReviewComment pgtype.Text `json:"review_comment"`
}

func (q *Queries) SubmitReview(ctx context.Context, arg SubmitReviewParams) (Answer, error) {
	row := q.db.QueryRow(ctx, submitReview,
		arg.QuestionID,
		arg.Status,
		arg.DoctorID,
		arg.AnswerText,
		arg.AiDraftAnswer,
		arg.ReviewStatus,
		arg.ReviewComment,
	)
	var i Answer
	err := row.Scan(
		&i.ID,
		&i.QuestionID,
		&i.DoctorID,
		&i.AnswerText,
		&i.AiDraftAnswer,
		&i.AiConfidence,
		&i.AiReferences,
		&i.ReviewStatus,
		&i.ReviewComment,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const updateMedicalHistoryStatus = `-- name: UpdateMedicalHistoryStatus :one
UPDATE medical_history
SET 
    status = $3,
    notes = COALESCE($4, notes)
WHERE id = $1 AND patient_id = $2
RETURNING id, patient_id, condition, diagnosed_date, status, notes, created_at
`

type UpdateMedicalHistoryStatusParams struct {
	ID        pgtype.UUID `json:"id"`
	PatientID pgtype.UUID `json:"patient_id"`
	Status    string      `json:"status"`
	Notes     pgtype.Text `json:"notes"`
}

func (q *Queries) UpdateMedicalHistoryStatus(ctx context.Context, arg UpdateMedicalHistoryStatusParams) (MedicalHistory, error) {
	row := q.db.QueryRow(ctx, updateMedicalHistoryStatus,
		arg.ID,
		arg.PatientID,
		arg.Status,
		arg.Notes,
	)
	var i MedicalHistory
	err := row.Scan(
		&i.ID,
		&i.PatientID,
		&i.Condition,
		&i.DiagnosedDate,
		&i.Status,
		&i.Notes,
		&i.CreatedAt,
	)
	return i, err
}

const updatePatientDemographics = `-- name: UpdatePatientDemographics :one
UPDATE patients
SET
    age = $2,
    gender = $3
WHERE id = $1
RETURNING id, name, email, created_at, age, gender
`

type UpdatePatientDemographicsParams struct {
	ID     pgtype.UUID `json:"id"`
	Age    int32       `json:"age"`
	Gender string      `json:"gender"`
}

// Patient Demographics Update
func (q *Queries) UpdatePatientDemographics(ctx context.Context, arg UpdatePatientDemographicsParams) (Patient, error) {
	row := q.db.QueryRow(ctx, updatePatientDemographics, arg.ID, arg.Age, arg.Gender)
	var i Patient
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Email,
		&i.CreatedAt,
		&i.Age,
		&i.Gender,
	)
	return i, err
}

const updateQuestionStatus = `-- name: UpdateQuestionStatus :exec
UPDATE questions
SET 
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
`

type UpdateQuestionStatusParams struct {
	ID     pgtype.UUID `json:"id"`
	Status string      `json:"status"`
}

func (q *Queries) UpdateQuestionStatus(ctx context.Context, arg UpdateQuestionStatusParams) error {
	_, err := q.db.Exec(ctx, updateQuestionStatus, arg.ID, arg.Status)
	return err
}
