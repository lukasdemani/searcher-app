
CREATE TABLE IF NOT EXISTS urls (
    id INT PRIMARY KEY AUTO_INCREMENT,
    url VARCHAR(2048) NOT NULL,
    url_hash VARCHAR(255) NOT NULL UNIQUE,
    title VARCHAR(500),
    html_version VARCHAR(50),
    h1_count INT DEFAULT 0,
    h2_count INT DEFAULT 0,
    h3_count INT DEFAULT 0,
    h4_count INT DEFAULT 0,
    h5_count INT DEFAULT 0,
    h6_count INT DEFAULT 0,
    internal_links_count INT DEFAULT 0,
    external_links_count INT DEFAULT 0,
    broken_links_count INT DEFAULT 0,
    has_login_form BOOLEAN DEFAULT FALSE,
    status ENUM('queued', 'processing', 'completed', 'error') DEFAULT 'queued',
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    INDEX idx_status (status),
    INDEX idx_created_at (created_at),
    INDEX idx_url_hash (url_hash)
);

CREATE TABLE IF NOT EXISTS broken_links (
    id INT PRIMARY KEY AUTO_INCREMENT,
    url_id INT NOT NULL,
    link_url VARCHAR(1000) NOT NULL,
    status_code INT,
    error_message TEXT,
    
    FOREIGN KEY (url_id) REFERENCES urls(id) ON DELETE CASCADE,
    INDEX idx_url_id (url_id)
);

 