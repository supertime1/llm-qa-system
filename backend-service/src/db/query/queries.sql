-- Patient operations
-- name: CreateQuestion :one
INSERT INTO questions (
    patient_id,
    question_text,
    question_type,
    department,
    urgency_level,
    status
) VALUES (
    $1, $2, $3, $4, $5, 'STATUS_PENDING'
) RETURNING *;

-- name: GetQuestionStatus :one
SELECT 
    q.*,
    a.answer_text,
    a.doctor_id as answered_by
FROM questions q
LEFT JOIN answers a ON q.id = a.question_id
WHERE q.id = $1 AND q.patient_id = $2
LIMIT 1;

-- name: GetAnswerHistory :many
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
LIMIT $5 OFFSET $6;

-- name: GetAnswerHistoryCount :one
SELECT COUNT(*)
FROM questions
WHERE patient_id = $1
    AND ($2::question_status IS NULL OR status = $2)
    AND ($3::timestamptz IS NULL OR created_at >= $3)
    AND ($4::timestamptz IS NULL OR created_at <= $4);

-- Doctor operations
-- name: GetPendingReviews :many
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
LIMIT $6 OFFSET $7;

-- name: GetPendingReviewsCount :one
SELECT COUNT(*)
FROM questions
WHERE status = 'STATUS_PENDING_REVIEW'
    AND ($1::department IS NULL OR department = $1)
    AND ($2::question_type IS NULL OR question_type = $2)
    AND ($3::int IS NULL OR urgency_level >= $3)
    AND ($4::timestamptz IS NULL OR created_at >= $4)
    AND ($5::timestamptz IS NULL OR created_at <= $5);

-- name: SubmitReview :one
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
RETURNING *;

-- name: SaveAIDraftAnswer :one
INSERT INTO answers (
    question_id,
    ai_draft_answer,
    ai_confidence,
    ai_references,
    review_status
) VALUES (
    $1, $2, $3, $4, 'DECISION_UNSPECIFIED'
)
RETURNING *;

-- Patient Context Operations
-- name: GetPatientWithContext :one
SELECT 
    p.*,
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
GROUP BY p.id;

-- Biometric Data Operations
-- name: CreateBiometricData :one
INSERT INTO biometric_data (
    patient_id,
    type,
    value,
    unit,
    measured_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPatientBiometricData :many
SELECT * FROM biometric_data
WHERE patient_id = $1
    AND ($2::varchar IS NULL OR type = $2)
    AND ($3::timestamptz IS NULL OR measured_at >= $3)
    AND ($4::timestamptz IS NULL OR measured_at <= $4)
ORDER BY measured_at DESC
LIMIT $5 OFFSET $6;

-- name: GetLatestBiometricsByType :many
SELECT DISTINCT ON (type) *
FROM biometric_data
WHERE patient_id = $1
ORDER BY type, measured_at DESC;

-- Medical History Operations
-- name: CreateMedicalHistory :one
INSERT INTO medical_history (
    patient_id,
    condition,
    diagnosed_date,
    status,
    notes
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetPatientMedicalHistory :many
SELECT * FROM medical_history
WHERE patient_id = $1
    AND ($2::varchar IS NULL OR status = $2)
ORDER BY diagnosed_date DESC
LIMIT $3 OFFSET $4;

-- name: UpdateMedicalHistoryStatus :one
UPDATE medical_history
SET 
    status = $3,
    notes = COALESCE($4, notes)
WHERE id = $1 AND patient_id = $2
RETURNING *;

-- name: GetActiveMedicalConditions :many
SELECT * FROM medical_history
WHERE patient_id = $1
    AND status IN ('ACTIVE', 'CHRONIC')
ORDER BY diagnosed_date DESC;

-- Patient Demographics Update
-- name: UpdatePatientDemographics :one
UPDATE patients
SET
    age = $2,
    gender = $3
WHERE id = $1
RETURNING *;

-- name: UpdateQuestionStatus :exec
UPDATE questions
SET 
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;