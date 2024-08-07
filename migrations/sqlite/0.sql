CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ref_id INTEGER NOT NULL,
    category INTEGER NOT NULL,
    metadata VARCHAR NOT NULL,
    runner VARCHAR(255),
    create_time TIMESTAMP NOT NULL,
    schedule_time TIMESTAMP,
    start_time TIMESTAMP,
    last_seen_time TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_archives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ref_id VARCHAR(255) NOT NULL,
    runner VARCHAR(255),
    exit_code INT,
    category INTEGER NOT NULL,
    metadata VARCHAR NOT NULL,
    create_time TIMESTAMP NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS record_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    params VARCHAR(4096) NOT NULL,
    create_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS record_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INT,
    domain_name VARCHAR(1024) NOT NULL,
    app_name VARCHAR(1024),
    stream_name VARCHAR(1024),
    create_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS record_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INT,
    domain_name VARCHAR(1024) NOT NULL,
    app_name VARCHAR(1024) NOT NULL,
    stream_name VARCHAR(1024) NOT NULL,
    stream_type INT,
    source_url VARCHAR(1024),
    store_path VARCHAR(1024),
    create_time TIMESTAMP NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS record_cb_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    description VARCHAR(4096),
    callback_key VARCHAR(4096),
    begin_url VARCHAR(4096),
    end_url VARCHAR(4096),
    record_url VARCHAR(4096),
    record_status_url VARCHAR(4096),
    porn_censorship_url VARCHAR(4096),
    stream_mix_url VARCHAR(4096),
    push_exception_url VARCHAR(4096),
    audio_audit_url VARCHAR(4096),
    snapshot_url VARCHAR(4096),
    create_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS record_cb_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INT,
    domain_name VARCHAR(1024) NOT NULL,
    app_name VARCHAR(1024),
    create_time TIMESTAMP NOT NULL
);

