CREATE DATABASE IF NOT EXISTS clusterd;

USE clusterd;

CREATE TABLE IF NOT EXISTS jobs (
    id INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    ref_id VARCHAR(255) NOT NULL,
    runner VARCHAR(255),
    cmd VARCHAR(1024) NOT NULL,
    create_time TIMESTAMP NOT NULL,
    start_time TIMESTAMP,
    last_seen_time TIMESTAMP
);

CREATE TABLE IF NOT EXISTS job_archives (
    id INT NOT NULL PRIMARY KEY,
    ref_id VARCHAR(255) NOT NULL,
    runner VARCHAR(255),
    exit_code INT,
    cmd VARCHAR(1024) NOT NULL,
    create_time TIMESTAMP NOT NULL,
    start_time TIMESTAMP,
    end_time TIMESTAMP NOT NULL
);
