-- Create tenant_attachments table for storing document attachments during tenant check-in
CREATE TABLE tenant_attachments (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    blob BYTEA NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_file_size CHECK (file_size <= 10485760), -- 10MB limit
    CONSTRAINT unique_tenant_attachment UNIQUE (tenant_id, file_name)
);

-- Create lease_attachments table for storing document attachments during lease operations
CREATE TABLE lease_attachments (
    id SERIAL PRIMARY KEY,
    lease_id INTEGER NOT NULL REFERENCES leases(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100) NOT NULL,
    file_size BIGINT NOT NULL,
    blob BYTEA NOT NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT chk_file_size CHECK (file_size <= 10485760), -- 10MB limit
    CONSTRAINT unique_lease_attachment UNIQUE (lease_id, file_name)
);

-- Create index on tenant_id for faster lookups
CREATE INDEX idx_tenant_attachments_tenant_id ON tenant_attachments(tenant_id);

-- Create index on lease_id for faster lookups
CREATE INDEX idx_lease_attachments_lease_id ON lease_attachments(lease_id);
