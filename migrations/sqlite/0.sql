CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ref_id INTEGER NOT NULL,
    category INTEGER NOT NULL,
    metadata VARCHAR NOT NULL,
    runner VARCHAR(255),
    create_time TIMESTAMP NOT NULL,
    schedule_time TIMESTAMP NOT NULL,
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
    create_time TIMESTAMP NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP NOT NULL
);

