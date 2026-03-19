-- Add verifier system tables

-- Bounty verifiers - defines what verifications are required for a bounty
CREATE TABLE IF NOT EXISTS bounty_verifiers (
    id SERIAL PRIMARY KEY,
    bounty_id INTEGER NOT NULL REFERENCES bounties(id) ON DELETE CASCADE,
    verifier_type VARCHAR(50) NOT NULL, -- 'manual', 'file_exists', 'test_suite', 'pr_approved', 'benchmark'
    config JSONB NOT NULL DEFAULT '{}', -- configuration specific to verifier type
    required BOOLEAN NOT NULL DEFAULT true, -- if false, verifier is optional
    weight REAL NOT NULL DEFAULT 1.0, -- weight for scoring (future use)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Verifier results - stores the results of running verifiers
CREATE TABLE IF NOT EXISTS verifier_results (
    id SERIAL PRIMARY KEY,
    bounty_id INTEGER NOT NULL REFERENCES bounties(id) ON DELETE CASCADE,
    verifier_id INTEGER NOT NULL REFERENCES bounty_verifiers(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'passed', 'failed', 'error'
    result_data JSONB DEFAULT '{}', -- detailed results from verifier
    error_message TEXT, -- error message if status = 'error'
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    checked_by VARCHAR(100) -- agent/user/system that ran the check
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_bounty_verifiers_bounty_id ON bounty_verifiers(bounty_id);
CREATE INDEX IF NOT EXISTS idx_verifier_results_bounty_id ON verifier_results(bounty_id);
CREATE INDEX IF NOT EXISTS idx_verifier_results_verifier_id ON verifier_results(verifier_id);

-- Example verifier configurations for common use cases:
-- File exists: {"files": ["README.md", "tests/test_*.py"], "base_path": "/repo"}
-- Test suite: {"command": "npm test", "working_dir": "/repo", "success_codes": [0]}
-- PR approved: {"repo": "owner/repo", "pr_number": 123, "required_approvals": 1}
-- Benchmark: {"metric": "response_time", "threshold": 100, "comparison": "less_than"}
-- Manual: {"instructions": "Check that feature works correctly", "assignee": "user123"}