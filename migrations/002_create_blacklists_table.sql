CREATE TABLE IF NOT EXISTS `blacklists` (
    `id` bigint unsigned AUTO_INCREMENT PRIMARY KEY,
    `user_id` bigint unsigned NULL COMMENT '按用户封禁，可空',
    `ip` varchar(45) DEFAULT '' COMMENT '封禁IP',
    `reason` varchar(256) NOT NULL COMMENT '封禁原因',
    `blocked_by` bigint unsigned NOT NULL COMMENT '操作管理员ID',
    `blocked_at` datetime(3) NOT NULL COMMENT '封禁时间',
    `expired_at` datetime(3) NULL COMMENT '到期时间，空=永久',
    `is_active` tinyint(1) DEFAULT 1 COMMENT 'true生效中 false已解封',
    `created_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` datetime(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_ip` (`ip`),
    INDEX `idx_expired_at` (`expired_at`),
    INDEX `idx_is_active` (`is_active`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='风控黑名单表';
