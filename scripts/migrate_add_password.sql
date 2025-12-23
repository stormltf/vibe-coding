-- 迁移脚本：为已有数据库添加 password 字段
-- 执行方式: mysql -u root -p test < scripts/migrate_add_password.sql

USE test;

-- 检查并添加 password 字段
SET @column_exists = (
    SELECT COUNT(*)
    FROM information_schema.COLUMNS
    WHERE TABLE_SCHEMA = 'test'
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'password'
);

SET @sql = IF(@column_exists = 0,
    'ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL DEFAULT \'\' AFTER email',
    'SELECT "password column already exists"'
);

PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 为已有用户设置默认密码 (password123 的 bcrypt hash)
-- 注意：生产环境应该要求用户重置密码
UPDATE users
SET password = '$2a$10$N9qo8uLOickgx2ZMRZoMye1QV3Jg6O6k3lm0uI8U4dRH7E5KmFMeq'
WHERE password = '' OR password IS NULL;

SELECT 'Migration completed successfully' AS status;
