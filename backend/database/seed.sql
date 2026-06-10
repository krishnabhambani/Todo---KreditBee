-- Sample Seed Data for MySQL Database

-- Passwords are set to "password123" (hashed via bcrypt)
-- Bcrypt Hash: $2a$10$jmq15Jfjlg/6aWnDB/WKjuALwYGZVkv6vBdSh0Qo9fd83SHof5K5q

INSERT INTO `users` (`id`, `name`, `email`, `password`, `created_at`) VALUES
(1, 'John Doe', 'john@example.com', '$2a$10$jmq15Jfjlg/6aWnDB/WKjuALwYGZVkv6vBdSh0Qo9fd83SHof5K5q', NOW()),
(2, 'Jane Smith', 'jane@example.com', '$2a$10$jmq15Jfjlg/6aWnDB/WKjuALwYGZVkv6vBdSh0Qo9fd83SHof5K5q', NOW())
ON DUPLICATE KEY UPDATE `id`=`id`;

INSERT INTO `todos` (`id`, `title`, `description`, `completed`, `user_id`, `created_at`, `updated_at`) VALUES
(1, 'Buy Groceries', 'Pick up milk, organic eggs, avocados, and whole wheat bread.', false, 1, NOW(), NOW()),
(2, 'Submit Project Proposal', 'Complete the final draft of the Q3 project proposal and send to team lead.', true, 1, NOW(), NOW()),
(3, 'Workout Session', '45 minutes cardiovascular training and core exercises.', false, 1, NOW(), NOW()),
(4, 'Doctor Appointment', 'Routine checkup scheduled for Thursday morning.', false, 2, NOW(), NOW()),
(5, 'Read Tech Article', 'Read up on GORM optimizations and MySQL indexes.', true, 2, NOW(), NOW())
ON DUPLICATE KEY UPDATE `id`=`id`;
