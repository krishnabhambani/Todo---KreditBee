CREATE TABLE IF NOT EXISTS `todos` (
    `id` INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(255) NOT NULL,
    `description` TEXT,
    `completed` BOOLEAN DEFAULT FALSE,
    `due_date` TIMESTAMP DEFAULT NULL,
    `user_id` INT UNSIGNED NOT NULL,
    `parent_todo_id` INT UNSIGNED DEFAULT NULL,
    `created_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT `fk_todos_users` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
    CONSTRAINT `fk_todos_parent` FOREIGN KEY (`parent_todo_id`) REFERENCES `todos` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX `idx_todos_user_id` ON `todos` (`user_id`);
CREATE INDEX `idx_todos_completed` ON `todos` (`completed`);
CREATE INDEX `idx_todos_parent_id` ON `todos` (`parent_todo_id`);

