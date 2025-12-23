-- 初始化数据库脚本
USE test;

-- 创建用户表（优化索引）
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL DEFAULT 0,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    UNIQUE INDEX idx_email (email),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入测试数据
-- 注意：password 字段存储的是 bcrypt 加密后的密码
-- 下面的测试数据密码均为 "password123" 的 bcrypt hash
INSERT INTO users (name, age, email, password) VALUES
('张三', 25, 'zhangsan@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'),
('李四', 30, 'lisi@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'),
('王五', 28, 'wangwu@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'),
('赵六', 35, 'zhaoliu@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'),
('钱七', 22, 'qianqi@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq');

-- 批量插入更多测试数据用于性能测试
-- 注意：测试数据使用统一的密码 hash (password123)
DELIMITER //
CREATE PROCEDURE InsertTestUsers(IN num INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    WHILE i <= num DO
        INSERT INTO users (name, age, email, password) VALUES
        (CONCAT('user_', i), FLOOR(18 + RAND() * 50), CONCAT('user', i, '@test.com'), '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq');
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;

-- 插入 1000 条测试数据
-- CALL InsertTestUsers(1000);
