-- 初始化数据库脚本
USE test;

-- 创建用户表（优化索引）
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    age INT NOT NULL DEFAULT 0,
    email VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    UNIQUE INDEX idx_email (email),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 插入测试数据
INSERT INTO users (name, age, email) VALUES
('张三', 25, 'zhangsan@example.com'),
('李四', 30, 'lisi@example.com'),
('王五', 28, 'wangwu@example.com'),
('赵六', 35, 'zhaoliu@example.com'),
('钱七', 22, 'qianqi@example.com');

-- 批量插入更多测试数据用于性能测试
DELIMITER //
CREATE PROCEDURE InsertTestUsers(IN num INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    WHILE i <= num DO
        INSERT INTO users (name, age, email) VALUES
        (CONCAT('user_', i), FLOOR(18 + RAND() * 50), CONCAT('user', i, '@test.com'));
        SET i = i + 1;
    END WHILE;
END //
DELIMITER ;

-- 插入 1000 条测试数据
-- CALL InsertTestUsers(1000);
