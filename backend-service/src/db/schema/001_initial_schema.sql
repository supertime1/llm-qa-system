-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Reference tables for types
CREATE TABLE ref_gender (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ref_chat_session_status (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ref_departments (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ref_biometric_types (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    unit_type VARCHAR(20) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE ref_medical_condition_status (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE ref_chat_message_type (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Prompt template management
CREATE TABLE ref_prompt_templates (
    version VARCHAR(50) PRIMARY KEY,  -- e.g., "v1.0", "v2.1"
    template TEXT NOT NULL,           -- The actual template with placeholders
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Base users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Patient-specific information
CREATE TABLE patients (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    age INTEGER NOT NULL CHECK (age >= 0 AND age <= 150),
    gender VARCHAR(20) NOT NULL REFERENCES ref_gender(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Doctor-specific information
CREATE TABLE doctors (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    department_id VARCHAR(50) REFERENCES ref_departments(id),
    specialization TEXT[],
    years_of_experience INTEGER CHECK (years_of_experience >= 0),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Medical history for patients
CREATE TABLE medical_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(user_id),
    condition TEXT NOT NULL,
    diagnosed_date TIMESTAMPTZ,
    status_id VARCHAR(50) NOT NULL REFERENCES ref_medical_condition_status(id),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Biometric data for patients
CREATE TABLE biometric_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(user_id),
    type_id VARCHAR(50) NOT NULL REFERENCES ref_biometric_types(id),
    value NUMERIC NOT NULL,
    measured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Chat sessions
CREATE TABLE chat_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(user_id),
    status VARCHAR(50) NOT NULL REFERENCES ref_chat_session_status(id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    closed_at TIMESTAMPTZ
);

-- Chat messages
CREATE TABLE chat_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_session_id UUID NOT NULL REFERENCES chat_sessions(id),
    sender_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    message_type VARCHAR(50) NOT NULL REFERENCES ref_chat_message_type(id),
    parent_message_id UUID REFERENCES chat_messages(id), 
    -- Patient Message (id: 123)
    -- └── AI Draft (parent_message_id: 123)
    --     └── Doctor Review (parent_message_id: AI Draft's id)
    metadata JSONB,  -- For storing additional data like confidence scores, references, etc.
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- AI interactions table
CREATE TABLE ai_interactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_message_id UUID NOT NULL REFERENCES chat_messages(id),  -- The AI draft message
    prompt_template_version VARCHAR(50) NOT NULL REFERENCES ref_prompt_templates(version),
    prompt_components JSONB NOT NULL,  -- Store only the dynamic parts
    -- Example structure:
    -- {
    --   "instruction": "specific instruction for this case",
    --   "question": "patient's question",
    --   "context_snapshot": {
    --     "biometric_data_ids": ["uuid1", "uuid2"],  -- References instead of full data
    --     "medical_history_ids": ["uuid1", "uuid2"],
    --     "recent_message_ids": ["uuid1", "uuid2"]
    --   }
    -- }
    ai_response TEXT NOT NULL,  -- The raw AI response
    confidence_score FLOAT,
    references TEXT[],
    review_status VARCHAR(50),  -- approved, rejected, modified
    review_comment TEXT,
    modified_content TEXT,      -- If doctor modified the response
    reviewed_at TIMESTAMPTZ,
    reviewed_by UUID REFERENCES doctors(user_id),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);


-- Insert initial reference data
INSERT INTO ref_gender (id, name, description) VALUES
    ('GENDER_MALE', 'Male', 'Male gender'),
    ('GENDER_FEMALE', 'Female', 'Female gender'),
    ('GENDER_OTHER', 'Other', 'Other gender'),
    ('GENDER_PREFER_NOT_TO_SAY', 'Prefer Not to Say', 'Prefer not to say');

INSERT INTO ref_chat_session_status (id, name, description) VALUES
    ('CHAT_SESSION_STATUS_OPEN', 'Open', 'Chat session is open'),
    ('CHAT_SESSION_STATUS_CLOSED', 'Closed', 'Chat session is closed');

INSERT INTO ref_departments (id, name, description) VALUES
    ('DEPT_GENERAL_MEDICINE', 'General Medicine', 'General medical practice'),
    ('DEPT_CARDIOLOGY', 'Cardiology', 'Heart and cardiovascular system'),
    ('DEPT_PEDIATRICS', 'Pediatrics', 'Medical care for children'),
    ('DEPT_DERMATOLOGY', 'Dermatology', 'Skin related conditions');

INSERT INTO ref_biometric_types (id, name, unit_type, description) VALUES
    ('BLOOD_PRESSURE', 'Blood Pressure', 'mmHg', 'Systolic and diastolic blood pressure'),
    ('HEART_RATE', 'Heart Rate', 'bpm', 'Heart beats per minute'),
    ('TEMPERATURE', 'Body Temperature', '°C', 'Body temperature'),
    ('BLOOD_GLUCOSE', 'Blood Glucose', 'mg/dL', 'Blood sugar level'),
    ('WEIGHT', 'Weight', 'kg', 'Body weight'),
    ('HEIGHT', 'Height', 'cm', 'Body height'),
    ('BMI', 'Body Mass Index', 'kg/m²', 'Body mass index'),
    ('OXYGEN_SATURATION', 'Oxygen Saturation', '%', 'Blood oxygen level'),
    ('RESPIRATORY_RATE', 'Respiratory Rate', 'bpm', 'Breaths per minute'),
    ('STEPS', 'Step Count', 'steps', 'Number of steps taken'); -- Easy to add new types!

INSERT INTO ref_medical_condition_status (id, name, description) VALUES
    ('ACTIVE', 'Active', 'Current medical condition'),
    ('RESOLVED', 'Resolved', 'Past medical condition'),
    ('CHRONIC', 'Chronic', 'Ongoing long-term condition');

INSERT INTO ref_chat_message_type (id, name, description) VALUES
    ('PATIENT_MESSAGE', 'Patient Message', 'Message sent by the patient'),
    ('DOCTOR_MESSAGE', 'Doctor Message', 'Message sent by the doctor'),
    ('AI_DRAFT', 'AI Draft', 'Draft message generated by the AI'),
    ('AI_DRAFT_APPROVED', 'AI Draft Approved', 'Draft message approved by the doctor'),
    ('AI_DRAFT_MODIFIED', 'AI Draft Modified', 'Draft message modified by the doctor'),
    ('AI_DRAFT_REJECTED', 'AI Draft Rejected', 'Draft message rejected by the doctor'),
    ('SYSTEM_MESSAGE', 'System Message', 'System generated message');

-- Example insert
INSERT INTO ref_prompt_templates (version, template, description) VALUES
    ('v1.0', 
     'You are a medical AI assistant. Please provide a clear, accurate, and empathetic response to the patient''s question. Base your response on the provided medical context and history. Include relevant medical information and recommendations.',
     'Initial prompt template for medical responses');

-- Indexes for better query performance
CREATE INDEX idx_chat_messages_session ON chat_messages(chat_session_id);
CREATE INDEX idx_chat_messages_parent ON chat_messages(parent_message_id);
CREATE INDEX idx_chat_messages_created ON chat_messages(created_at);
CREATE INDEX idx_chat_messages_type ON chat_messages(message_type);
CREATE INDEX idx_medical_history_patient ON medical_history(patient_id);
CREATE INDEX idx_biometric_data_patient ON biometric_data(patient_id);
CREATE INDEX idx_biometric_data_measured_at ON biometric_data(measured_at);
CREATE INDEX idx_ai_interactions_chat_message ON ai_interactions(chat_message_id);