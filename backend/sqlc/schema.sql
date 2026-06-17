-- SQLC canonical schema for Todo Application
-- Source of truth for all table definitions.

CREATE TABLE IF NOT EXISTS `users` (
    `id`         INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `name`       VARCHAR(100) NOT NULL,
    `email`      VARCHAR(191) NOT NULL,
    `password`   VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS `todos` (
    `id`             INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title`          VARCHAR(255) NOT NULL,
    `description`    TEXT NOT NULL DEFAULT '',
    `completed`      BOOLEAN NOT NULL DEFAULT FALSE,
    `due_date`       TIMESTAMP NULL DEFAULT NULL,
    `user_id`        INT UNSIGNED NOT NULL,
    `parent_todo_id` INT UNSIGNED NULL DEFAULT NULL,
    `created_at`     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at`     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT `fk_todos_users`  FOREIGN KEY (`user_id`)        REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_todos_parent` FOREIGN KEY (`parent_todo_id`) REFERENCES `todos` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX IF NOT EXISTS `idx_todos_user_id`   ON `todos` (`user_id`);
CREATE INDEX IF NOT EXISTS `idx_todos_completed`  ON `todos` (`completed`);
CREATE INDEX IF NOT EXISTS `idx_todos_parent_id`  ON `todos` (`parent_todo_id`);

CREATE TABLE IF NOT EXISTS `group_shares` (
    `id`                   INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `group_id`             INT UNSIGNED NOT NULL,
    `owner_id`             INT UNSIGNED NOT NULL,
    `shared_with_user_id`  INT UNSIGNED NOT NULL,
    `permission`           VARCHAR(50) NOT NULL,
    `created_at`           TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY `idx_group_share_unique` (`group_id`, `shared_with_user_id`),
    CONSTRAINT `fk_group_shares_group`  FOREIGN KEY (`group_id`)            REFERENCES `todos` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_group_shares_owner`  FOREIGN KEY (`owner_id`)            REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_group_shares_target` FOREIGN KEY (`shared_with_user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
