-- ============================================================================
-- Vibe Coding - Database Initialization Script
-- ============================================================================
-- This script initializes the complete database schema for Vibe Coding.
--
-- Usage:
--   mysql -u root -p < scripts/init.sql
--
-- Requirements:
--   - MySQL 8.0+
--   - Root or admin privileges to create database
--
-- Test Account:
--   Email: test@example.com
--   Password: password123
-- ============================================================================

-- ----------------------------------------------------------------------------
-- 1. Create Database
-- ----------------------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `vibe_coding`
    DEFAULT CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE `vibe_coding`;

-- ----------------------------------------------------------------------------
-- 2. Create Users Table
-- ----------------------------------------------------------------------------
-- Primary table for user authentication and profile management
CREATE TABLE IF NOT EXISTS `users` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Primary key',
    `name` VARCHAR(100) NOT NULL COMMENT 'Username for display',
    `age` INT NOT NULL DEFAULT 0 COMMENT 'User age',
    `email` VARCHAR(255) NOT NULL COMMENT 'Email address (unique, for login)',
    `password` VARCHAR(255) NOT NULL DEFAULT '' COMMENT 'Bcrypt hashed password',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    PRIMARY KEY (`id`),
    UNIQUE INDEX `idx_email` (`email`) COMMENT 'Unique email for login',
    INDEX `idx_name` (`name`) COMMENT 'Index for name search',
    INDEX `idx_created_at` (`created_at`) COMMENT 'Index for pagination'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User accounts';

-- ----------------------------------------------------------------------------
-- 3. Create Projects Table
-- ----------------------------------------------------------------------------
-- Stores user's AI-generated web projects
CREATE TABLE IF NOT EXISTS `projects` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Primary key',
    `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'Owner user ID',
    `name` VARCHAR(255) NOT NULL DEFAULT 'New Project' COMMENT 'Project name',
    `html` LONGTEXT COMMENT 'Generated HTML content',
    `css` LONGTEXT COMMENT 'Generated CSS content',
    `messages` LONGTEXT COMMENT 'Chat history in JSON format',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) COMMENT 'Creation timestamp',
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3) COMMENT 'Last update timestamp',
    PRIMARY KEY (`id`),
    INDEX `idx_project_user_id` (`user_id`) COMMENT 'Index for user projects lookup',
    INDEX `idx_project_created_at` (`created_at`) COMMENT 'Index for sorting',
    CONSTRAINT `fk_project_user` FOREIGN KEY (`user_id`)
        REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='User workspace projects';

-- ----------------------------------------------------------------------------
-- 4. Insert Test Data
-- ----------------------------------------------------------------------------
-- Test accounts for development and demo purposes
-- All passwords are bcrypt hash of "password123"
--
-- To generate a new bcrypt hash:
--   node -e "console.log(require('bcrypt').hashSync('your-password', 10))"
--   or use: htpasswd -bnBC 10 "" your-password | tr -d ':\n'

INSERT INTO `users` (`name`, `age`, `email`, `password`) VALUES
('Test User', 25, 'test@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'),
('Demo User', 30, 'demo@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'),
('Admin', 35, 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq')
ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP;

-- ----------------------------------------------------------------------------
-- 5. Create Sample Project (Optional)
-- ----------------------------------------------------------------------------
INSERT INTO `projects` (`user_id`, `name`, `html`, `css`, `messages`)
SELECT
    u.id,
    'Welcome Project',
    '<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome to Vibe Coding</title>
</head>
<body>
    <div class="container">
        <h1>Welcome to Vibe Coding</h1>
        <p>This is your first AI-generated project.</p>
        <p>Try creating something new with natural language!</p>
    </div>
</body>
</html>',
    'body {
    font-family: system-ui, -apple-system, sans-serif;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0;
    color: white;
}

.container {
    text-align: center;
    padding: 2rem;
    background: rgba(255,255,255,0.1);
    border-radius: 1rem;
    backdrop-filter: blur(10px);
}

h1 {
    margin-bottom: 1rem;
}',
    '[]'
FROM `users` u
WHERE u.email = 'test@example.com'
ON DUPLICATE KEY UPDATE `updated_at` = CURRENT_TIMESTAMP(3);

-- ----------------------------------------------------------------------------
-- 6. Stored Procedure for Bulk Test Data (Optional)
-- ----------------------------------------------------------------------------
-- Use this to generate large amounts of test data for performance testing
--
-- Usage: CALL InsertTestUsers(1000);

DROP PROCEDURE IF EXISTS `InsertTestUsers`;

DELIMITER //

CREATE PROCEDURE `InsertTestUsers`(IN num INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE batch_size INT DEFAULT 100;
    DECLARE current_batch INT DEFAULT 0;

    -- Disable autocommit for better performance
    SET autocommit = 0;

    WHILE i <= num DO
        INSERT INTO `users` (`name`, `age`, `email`, `password`) VALUES
        (
            CONCAT('user_', i),
            FLOOR(18 + RAND() * 50),
            CONCAT('user', i, '@test.com'),
            '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'
        );

        SET current_batch = current_batch + 1;

        -- Commit every batch_size records
        IF current_batch >= batch_size THEN
            COMMIT;
            SET current_batch = 0;
        END IF;

        SET i = i + 1;
    END WHILE;

    -- Final commit
    COMMIT;
    SET autocommit = 1;

    SELECT CONCAT('Successfully inserted ', num, ' test users') AS result;
END //

DELIMITER ;

-- ----------------------------------------------------------------------------
-- 7. Verification Queries
-- ----------------------------------------------------------------------------
-- Uncomment these to verify the installation

-- Check table structure
-- DESCRIBE users;
-- DESCRIBE projects;

-- Check test data
-- SELECT id, name, email, created_at FROM users;
-- SELECT id, user_id, name, created_at FROM projects;

-- Check foreign key constraints
-- SELECT TABLE_NAME, CONSTRAINT_NAME, REFERENCED_TABLE_NAME
-- FROM information_schema.KEY_COLUMN_USAGE
-- WHERE TABLE_SCHEMA = 'vibe_coding' AND REFERENCED_TABLE_NAME IS NOT NULL;

SELECT 'Database initialization completed successfully!' AS status;
SELECT COUNT(*) AS user_count FROM users;
SELECT COUNT(*) AS project_count FROM projects;
