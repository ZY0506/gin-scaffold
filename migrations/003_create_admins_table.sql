-- 管理员账号表
CREATE TABLE IF NOT EXISTS `admins` (
    `id` bigint unsigned AUTO_INCREMENT PRIMARY KEY,
    `username` varchar(32) NOT NULL UNIQUE COMMENT '管理员用户名',
    `password` varchar(256) NOT NULL COMMENT 'bcrypt hash',
    `nickname` varchar(64) DEFAULT '' COMMENT '昵称',
    `avatar` varchar(256) DEFAULT '' COMMENT '头像URL',
    `status` tinyint DEFAULT 1 COMMENT '1正常 0禁用',
    `last_login_at` datetime(3) NULL COMMENT '最后登录时间',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='管理员表';

-- 管理员操作日志表
CREATE TABLE IF NOT EXISTS `operation_logs` (
    `id` bigint unsigned AUTO_INCREMENT PRIMARY KEY,
    `admin_id` bigint unsigned NOT NULL COMMENT '操作管理员ID',
    `method` varchar(10) NOT NULL COMMENT '请求方法',
    `path` varchar(512) NOT NULL COMMENT '请求路径',
    `status_code` int DEFAULT 0 COMMENT 'HTTP状态码',
    `client_ip` varchar(45) DEFAULT '' COMMENT '客户端IP',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    INDEX `idx_admin_id` (`admin_id`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志表';
