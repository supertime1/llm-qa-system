-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS doctors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    department VARCHAR(50) NOT NULL CHECK (department IN 
        ('DEPT_UNSPECIFIED', 'DEPT_GENERAL_MEDICINE', 'DEPT_CARDIOLOGY', 'DEPT_PEDIATRICS', 'DEPT_DERMATOLOGY')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(id),
    question_text TEXT NOT NULL,
    question_type VARCHAR(50) NOT NULL CHECK (question_type IN 
        ('TYPE_UNSPECIFIED', 'TYPE_GENERAL', 'TYPE_MEDICATION', 'TYPE_DIAGNOSIS', 'TYPE_FOLLOW_UP')),
    department VARCHAR(50) NOT NULL CHECK (department IN 
        ('DEPT_UNSPECIFIED', 'DEPT_GENERAL_MEDICINE', 'DEPT_CARDIOLOGY', 'DEPT_PEDIATRICS', 'DEPT_DERMATOLOGY')),
    urgency_level INTEGER CHECK (urgency_level BETWEEN 1 AND 5),
    status VARCHAR(50) NOT NULL CHECK (status IN 
        ('STATUS_UNSPECIFIED', 'STATUS_PENDING', 'STATUS_PROCESSING', 'STATUS_PENDING_REVIEW', 
         'STATUS_UNDER_REVIEW', 'STATUS_ANSWERED', 'STATUS_REJECTED')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    answered_at TIMESTAMPTZ,
    answered_by UUID REFERENCES doctors(id)
);

CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    question_id UUID NOT NULL REFERENCES questions(id),
    doctor_id UUID REFERENCES doctors(id),
    answer_text TEXT NOT NULL,
    ai_draft_answer TEXT,
    ai_confidence FLOAT CHECK (ai_confidence >= 0 AND ai_confidence <= 1),
    ai_references TEXT[],
    review_status VARCHAR(50) NOT NULL CHECK (review_status IN 
        ('DECISION_UNSPECIFIED', 'DECISION_APPROVED', 'DECISION_MODIFIED', 'DECISION_REJECTED')),
    review_comment TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);