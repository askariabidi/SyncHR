 -- SyncHR Database Schema
-- A Distributed Web-Based Human Resource Management System
-- Created for Distributed Programming Course - University of Florence

-- ============================================================================
-- TABLE: USERS
-- ============================================================================
-- Stores user information for both employees and HR managers
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL CHECK (role IN ('employee', 'hr_manager')), -- Role-based access
    department VARCHAR(100),
    phone_number VARCHAR(20),
    date_of_joining DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- TABLE: ATTENDANCE
-- ============================================================================
-- Records daily attendance with check-in and check-out timestamps
CREATE TABLE IF NOT EXISTS attendance (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    check_in_time TIMESTAMP NOT NULL,
    check_out_time TIMESTAMP,
    date DATE NOT NULL,
    status VARCHAR(50) CHECK (status IN ('checked_in', 'checked_out', 'absent')), -- Status tracking
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, date) -- Only one check-in per day per user
);

-- ============================================================================
-- TABLE: BREAK_TIME
-- ============================================================================
-- Logs employee break times during the day
CREATE TABLE IF NOT EXISTS break_time (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    break_start_time TIMESTAMP NOT NULL,
    break_end_time TIMESTAMP,
    duration_minutes INT, -- Calculated duration
    date DATE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- TABLE: LEAVE_TYPES
-- ============================================================================
-- Predefined leave categories (sick, casual, earned, etc.)
CREATE TABLE IF NOT EXISTS leave_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    max_days_per_year INT DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- TABLE: LEAVE_BALANCE
-- ============================================================================
-- Tracks remaining leave balance per employee per leave type
CREATE TABLE IF NOT EXISTS leave_balance (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leave_type_id INT NOT NULL REFERENCES leave_types(id) ON DELETE CASCADE,
    balance INT NOT NULL DEFAULT 0, -- Remaining days
    year INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, leave_type_id, year) -- One balance per employee per leave type per year
);

-- ============================================================================
-- TABLE: LEAVE_REQUEST
-- ============================================================================
-- Stores leave requests with approval workflow
CREATE TABLE IF NOT EXISTS leave_request (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leave_type_id INT NOT NULL REFERENCES leave_types(id) ON DELETE CASCADE,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    number_of_days INT NOT NULL,
    reason TEXT,
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')), -- Approval status
    approved_by INT REFERENCES users(id) ON DELETE SET NULL, -- HR manager who approved
    approval_date TIMESTAMP,
    approval_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- TABLE: PUBLIC_HOLIDAYS
-- ============================================================================
-- Stores public and company holidays for the year
CREATE TABLE IF NOT EXISTS public_holidays (
    id SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    holiday_date DATE NOT NULL UNIQUE,
    description TEXT,
    country VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- TABLE: PAYSLIP
-- ============================================================================
-- Stores monthly payslip information for employees
CREATE TABLE IF NOT EXISTS payslip (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    month INT NOT NULL, -- Month (1-12)
    year INT NOT NULL,
    basic_salary DECIMAL(12, 2) NOT NULL,
    allowances DECIMAL(12, 2) DEFAULT 0,
    deductions DECIMAL(12, 2) DEFAULT 0,
    tax DECIMAL(12, 2) DEFAULT 0,
    net_salary DECIMAL(12, 2) NOT NULL,
    working_days INT,
    leave_taken INT DEFAULT 0,
    bonus DECIMAL(12, 2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, month, year) -- One payslip per employee per month
);

-- ============================================================================
-- TABLE: NOTIFICATIONS
-- ============================================================================
-- Stores real-time notifications for WebSocket communication
CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50), -- leave_approved, leave_rejected, payslip_generated, etc.
    related_entity_id INT, -- ID of related leave_request, payslip, etc.
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================
-- Indexes to optimize frequently queried columns
CREATE INDEX idx_attendance_user_date ON attendance(user_id, date);
CREATE INDEX idx_break_time_user_date ON break_time(user_id, date);
CREATE INDEX idx_leave_request_user ON leave_request(user_id);
CREATE INDEX idx_leave_request_status ON leave_request(status);
CREATE INDEX idx_leave_balance_user ON leave_balance(user_id);
CREATE INDEX idx_payslip_user_month_year ON payslip(user_id, month, year);
CREATE INDEX idx_notifications_user ON notifications(user_id);
CREATE INDEX idx_public_holidays_date ON public_holidays(holiday_date);

-- ============================================================================
-- SEED DATA: DEFAULT LEAVE TYPES
-- ============================================================================
-- Insert common leave types
INSERT INTO leave_types (name, description, max_days_per_year) VALUES
('Sick Leave', 'Leave for medical reasons', 10),
('Casual Leave', 'General leave for personal reasons', 12),
('Earned Leave', 'Paid leave earned based on service', 20),
('Maternity Leave', 'Leave for new mothers', 90),
('Paternity Leave', 'Leave for new fathers', 10)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- SEED DATA: SAMPLE PUBLIC HOLIDAYS (Italy 2025)
-- ============================================================================
INSERT INTO public_holidays (name, holiday_date, country, description) VALUES
('New Year''s Day', '2025-01-01', 'Italy', 'Beginning of the calendar year'),
('Epiphany', '2025-01-06', 'Italy', 'Religious observance'),
('Easter Monday', '2025-04-21', 'Italy', 'Day after Easter'),
('Liberation Day', '2025-04-25', 'Italy', 'Italian national holiday'),
('Labour Day', '2025-05-01', 'Italy', 'International Workers'' Day'),
('Republic Day', '2025-06-02', 'Italy', 'Italian national holiday'),
('Assumption of Mary', '2025-08-15', 'Italy', 'Religious observance'),
('All Saints'' Day', '2025-11-01', 'Italy', 'Religious observance'),
('Immaculate Conception', '2025-12-08', 'Italy', 'Religious observance'),
('Christmas Day', '2025-12-25', 'Italy', 'Christian holiday'),
('St. Stephen''s Day', '2025-12-26', 'Italy', 'Day after Christmas')
ON CONFLICT (holiday_date) DO NOTHING;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
