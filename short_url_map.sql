CREATE TABLE IF NOT EXISTS `short_url_map` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `scene_id` VARCHAR(64) NOT NULL DEFAULT '',
  `url_hash` CHAR(32) NOT NULL,
  `long_url` VARCHAR(2048) NOT NULL,
  `short_code` CHAR(6) NOT NULL,
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '1 normal / 0 disabled',
  `expire_at` DATETIME NULL COMMENT 'NULL = never expire',
  `created_by` VARCHAR(64) NOT NULL DEFAULT '',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_scene_hash` (`scene_id`, `url_hash`),
  UNIQUE KEY `uk_short_code` (`short_code`),
  KEY `idx_expire_at` (`expire_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='short/long URL mapping';
