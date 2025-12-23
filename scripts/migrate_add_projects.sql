-- Migration: Add projects table
-- Run this script to add project persistence support

-- Create projects table
CREATE TABLE IF NOT EXISTS `projects` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `name` VARCHAR(255) NOT NULL DEFAULT 'New Project',
    `html` LONGTEXT,
    `css` LONGTEXT,
    `messages` LONGTEXT COMMENT 'JSON format chat history',
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    INDEX `idx_project_user_id` (`user_id`),
    INDEX `idx_project_created_at` (`created_at`),
    CONSTRAINT `fk_project_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add comment to table
ALTER TABLE `projects` COMMENT = 'User workspace projects for vibe coding';
