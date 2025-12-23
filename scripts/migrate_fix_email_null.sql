-- 修复 email 字段允许 NULL 的问题
-- 执行前请先备份数据

-- 1. 查看有多少 NULL 邮箱的记录
SELECT COUNT(*) AS null_email_count FROM users WHERE email IS NULL;

-- 2. 处理 NULL 邮箱的记录（根据业务需求选择一种方式）
-- 方式 A: 删除这些记录
-- DELETE FROM users WHERE email IS NULL;

-- 方式 B: 给这些记录生成临时邮箱
-- UPDATE users SET email = CONCAT('user_', id, '@placeholder.com') WHERE email IS NULL;

-- 3. 修改字段约束为 NOT NULL
ALTER TABLE users MODIFY COLUMN email VARCHAR(255) NOT NULL;
