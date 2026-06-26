CREATE TABLE IF NOT EXISTS `users` (
    `id` bigint unsigned AUTO_INCREMENT PRIMARY KEY,
    `username` varchar(32) NOT NULL UNIQUE COMMENT '用户名',
    `email` varchar(128) NOT NULL UNIQUE COMMENT '邮箱',
    `password` varchar(256) NOT NULL COMMENT 'bcrypt hash',
    `nickname` varchar(64) DEFAULT '' COMMENT '昵称',
    `avatar` varchar(256) DEFAULT '' COMMENT '头像URL',
    `gender` tinyint DEFAULT 0 COMMENT '0未知 1男 2女',
    `birthday` datetime(3) NULL COMMENT '生日',
    `status` tinyint DEFAULT 1 COMMENT '1正常 0禁用',
    `last_login_at` datetime(3) NULL COMMENT '最后登录时间',
    `last_login_ip` varchar(45) DEFAULT '' COMMENT '最后登录IP',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX `idx_gender` (`gender`),
    INDEX `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';
