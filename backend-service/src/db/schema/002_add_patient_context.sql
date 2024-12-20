-- Add demographic columns to existing patients table
ALTER TABLE patients
    ADD COLUMN age INTEGER NOT NULL CHECK (age >= 0 AND age <= 150),
    ADD COLUMN gender VARCHAR(20) NOT NULL CHECK (gender IN ('MALE', 'FEMALE', 'OTHER', 'PREFER_NOT_TO_SAY'));

-- Create medical history table
CREATE TABLE medical_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(id),
    condition TEXT NOT NULL,
    diagnosed_date TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL CHECK (status IN ('ACTIVE', 'RESOLVED', 'CHRONIC')),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Create biometric data table
CREATE TABLE biometric_data (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    patient_id UUID NOT NULL REFERENCES patients(id),
    type VARCHAR(50) NOT NULL CHECK (type IN (
        'BLOOD_PRESSURE', 'HEART_RATE', 'TEMPERATURE', 
        'BLOOD_GLUCOSE', 'WEIGHT', 'HEIGHT', 'BMI', 
        'OXYGEN_SATURATION', 'RESPIRATORY_RATE'
    )),
    value TEXT NOT NULL,
    unit VARCHAR(20) NOT NULL,
    measured_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for better query performance
CREATE INDEX idx_medical_history_patient ON medical_history(patient_id);
CREATE INDEX idx_biometric_data_patient ON biometric_data(patient_id);
CREATE INDEX idx_biometric_data_measured_at ON biometric_data(measured_at);