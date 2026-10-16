CREATE TABLE IF NOT EXISTS `click_events` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `short_code` CHAR(6) NOT NULL,
  `scene_id` VARCHAR(64) NOT NULL DEFAULT '',
  `visitor_hash` CHAR(64) NOT NULL,
  `clicked_at` DATETIME(3) NOT NULL COMMENT 'UTC',
  PRIMARY KEY (`id`),
  KEY `idx_clicked_at` (`clicked_at`),
  KEY `idx_code_clicked` (`short_code`, `clicked_at`),
  KEY `idx_scene_clicked` (`scene_id`, `clicked_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='click events for analytics (UTC)';
