-- User related queries
-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
    email,
    name
) VALUES (
    $1, $2
) RETURNING *;

-- Patient related queries
-- name: CreatePatient :one
INSERT INTO patients (
    user_id,
    age,
    gender
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPatientByUserID :one
SELECT 
    u.id,
    u.email,
    u.name,
    p.age,
    p.gender,
    p.created_at
FROM users u
JOIN patients p ON p.user_id = u.id
WHERE u.id = $1;

-- Doctor related queries
-- name: CreateDoctor :one
INSERT INTO doctors (
    user_id,
    department_id,
    specialization,
    years_of_experience
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetDoctorByUserID :one
SELECT 
    u.id,
    u.email,
    u.name,
    d.department_id,
    dept.name as department_name,
    d.specialization,
    d.years_of_experience
FROM users u
JOIN doctors d ON d.user_id = u.id
JOIN ref_departments dept ON dept.id = d.department_id
WHERE u.id = $1;

-- Medical History
-- name: AddMedicalHistory :one
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
ORDER BY diagnosed_date DESC;

-- name: GetActiveConditions :many
SELECT * FROM medical_history 
WHERE patient_id = $1 
AND status = 'ACTIVE' 
ORDER BY diagnosed_date DESC;

-- Biometric Data
-- name: AddBiometricData :one
INSERT INTO biometric_data (
    patient_id,
    type_id,
    value,
    measured_at
) VALUES (
    $1, $2, $3, COALESCE($4, CURRENT_TIMESTAMP)
) RETURNING *;

-- name: GetLatestBiometrics :many
SELECT 
    rt.name as type_name,
    rt.unit_type,
    bd.value,
    bd.measured_at
FROM (
    SELECT DISTINCT ON (type_id) *
    FROM biometric_data
    WHERE patient_id = $1
    ORDER BY type_id, measured_at DESC
) bd
JOIN ref_biometric_types rt ON rt.id = bd.type_id;

-- Reference Data queries
-- name: ListActiveBiometricTypes :many
SELECT * FROM ref_biometric_types 
WHERE active = true 
ORDER BY name;

-- name: ListActiveDepartments :many
SELECT * FROM ref_departments 
WHERE active = true 
ORDER BY name;

-- name: GetActivePromptTemplate :one
SELECT version, template 
FROM ref_prompt_templates 
WHERE is_active = true 
ORDER BY created_at DESC 
LIMIT 1;

-- Chat Session Management
-- name: CreateChatSession :one
INSERT INTO chat_sessions (
    patient_id,
    status
) VALUES (
    $1,
    'CHAT_SESSION_STATUS_OPEN'
) RETURNING *;

-- name: UpdateChatSessionStatus :exec
UPDATE chat_sessions 
SET 
    status = $2,
    closed_at = CASE WHEN $2 = 'CHAT_SESSION_STATUS_CLOSED' THEN CURRENT_TIMESTAMP ELSE NULL END
WHERE id = $1;

-- name: GetActiveChatSession :one
SELECT * FROM chat_sessions 
WHERE patient_id = $1 
AND status = 'CHAT_SESSION_STATUS_OPEN' 
ORDER BY created_at DESC 
LIMIT 1;

-- Chat Messages
-- name: CreateChatMessage :one
INSERT INTO chat_messages (
    chat_session_id,
    sender_id,
    content,
    message_type,
    parent_message_id,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetChatHistory :many
SELECT 
    cm.id,
    cm.chat_session_id,
    cm.sender_id,
    u.name as sender_name,
    CASE 
        WHEN d.user_id IS NOT NULL THEN 'DOCTOR'
        WHEN p.user_id IS NOT NULL THEN 'PATIENT'
        ELSE 'SYSTEM'
    END as sender_role,
    cm.content,
    cm.message_type,
    cm.parent_message_id,
    cm.metadata,
    cm.created_at
FROM chat_messages cm
JOIN users u ON u.id = cm.sender_id
LEFT JOIN doctors d ON d.user_id = cm.sender_id
LEFT JOIN patients p ON p.user_id = cm.sender_id
WHERE cm.chat_session_id = $1
ORDER BY cm.created_at DESC
LIMIT $2;

-- Patient Context Query
-- name: GetPatientContext :one
SELECT 
    u.id,
    u.name,
    p.age,
    p.gender,
    (
        SELECT json_agg(json_build_object(
            'condition', condition,
            'status', status,
            'diagnosed_date', diagnosed_date
        ))
        FROM medical_history
        WHERE patient_id = p.user_id 
        AND status = 'ACTIVE'
    ) as active_conditions,
    (
        SELECT json_agg(json_build_object(
            'type', rt.name,
            'value', bd.value,
            'unit', rt.unit_type,
            'measured_at', bd.measured_at
        ))
        FROM (
            SELECT DISTINCT ON (type_id) *
            FROM biometric_data
            WHERE patient_id = p.user_id
            ORDER BY type_id, measured_at DESC
        ) bd
        JOIN ref_biometric_types rt ON rt.id = bd.type_id
    ) as recent_biometrics
FROM users u
JOIN patients p ON p.user_id = u.id
WHERE u.id = $1;

-- AI Interactions
-- name: CreateAIInteraction :one
INSERT INTO ai_interactions (
    chat_message_id,
    prompt_template_version,
    prompt_components,
    ai_response,
    confidence_score,
    references
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateAIInteractionReview :one
UPDATE ai_interactions
SET 
    review_status = $2,
    review_comment = $3,
    modified_content = $4,
    reviewed_at = CURRENT_TIMESTAMP,
    reviewed_by = $5
WHERE chat_message_id = $1
RETURNING *;

-- Training Data Collection
-- name: GetAITrainingData :many
SELECT 
    ai.id,
    ai.prompt_template_version,
    pt.template as prompt_template,
    ai.prompt_components,
    ai.ai_response,
    ai.confidence_score,
    ai.review_status,
    ai.modified_content,
    ai.review_comment,
    ai.created_at,
    ai.reviewed_at,
    d.user_id as reviewer_id,
    u.name as reviewer_name
FROM ai_interactions ai
LEFT JOIN ref_prompt_templates pt ON pt.version = ai.prompt_template_version
LEFT JOIN doctors d ON d.user_id = ai.reviewed_by
LEFT JOIN users u ON u.id = d.user_id
WHERE ai.created_at BETWEEN $1 AND $2
AND ai.review_status IS NOT NULL
ORDER BY ai.created_at DESC;

-- AI Performance Analytics
-- name: GetAIInteractionStats :one
SELECT 
    COUNT(*) as total_interactions,
    COUNT(CASE WHEN review_status = 'approved' THEN 1 END) as approved_count,
    COUNT(CASE WHEN review_status = 'rejected' THEN 1 END) as rejected_count,
    COUNT(CASE WHEN review_status = 'modified' THEN 1 END) as modified_count,
    AVG(confidence_score) as avg_confidence_score
FROM ai_interactions
WHERE created_at BETWEEN $1 AND $2;