-- ============================================================
-- Users
-- ============================================================

-- name: CreateUser :execresult
INSERT INTO users (name, email, password, created_at)
VALUES (?, ?, ?, NOW());

-- name: GetUserByEmail :one
SELECT id, name, email, password, created_at
FROM users
WHERE email = ?
LIMIT 1;

-- name: GetUserByID :one
SELECT id, name, email, password, created_at
FROM users
WHERE id = ?
LIMIT 1;

-- name: SearchUsers :many
SELECT id, name, email, password, created_at
FROM users
WHERE id != ?
  AND (name LIKE ? OR email LIKE ?)
LIMIT 20;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = ?
WHERE id = ?;

-- ============================================================
-- Todos (Groups + Subtasks share this table)
-- ============================================================

-- name: CreateTodo :execresult
INSERT INTO todos (title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW());

-- name: GetTodoByID :one
SELECT id, title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at
FROM todos
WHERE id = ?
LIMIT 1;

-- name: GetGroupsByUserID :many
-- Returns top-level groups (parent_todo_id IS NULL) for a user.
-- Filtering by status/sort is done in Go after this query.
SELECT id, title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at
FROM todos
WHERE user_id = ?
  AND parent_todo_id IS NULL
  AND (? = '' OR title LIKE ? OR description LIKE ?)
ORDER BY created_at DESC;

-- name: GetGroupByID :one
SELECT id, title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at
FROM todos
WHERE id = ?
  AND parent_todo_id IS NULL
LIMIT 1;

-- name: GetSubtasksByParentID :many
SELECT id, title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at
FROM todos
WHERE parent_todo_id = ?
ORDER BY created_at ASC;

-- name: GetSubtasksByParentAndUser :many
SELECT id, title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at
FROM todos
WHERE parent_todo_id = ?
  AND user_id = ?
ORDER BY created_at ASC;

-- name: UpdateTodo :exec
UPDATE todos
SET title = ?, description = ?, completed = ?, due_date = ?, updated_at = NOW()
WHERE id = ?;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = ?;

-- ============================================================
-- Group Shares
-- ============================================================

-- name: CreateGroupShare :execresult
INSERT INTO group_shares (group_id, owner_id, shared_with_user_id, permission, created_at)
VALUES (?, ?, ?, ?, NOW());

-- name: GetGroupShare :one
SELECT id, group_id, owner_id, shared_with_user_id, permission, created_at
FROM group_shares
WHERE group_id = ?
  AND shared_with_user_id = ?
LIMIT 1;

-- name: GetGroupSharesByGroupID :many
SELECT
    gs.id, gs.group_id, gs.owner_id, gs.shared_with_user_id, gs.permission, gs.created_at,
    u.id   AS sw_id,
    u.name AS sw_name,
    u.email AS sw_email
FROM group_shares gs
JOIN users u ON u.id = gs.shared_with_user_id
WHERE gs.group_id = ?;

-- name: GetSharesByUserID :many
SELECT
    gs.id, gs.group_id, gs.owner_id, gs.shared_with_user_id, gs.permission, gs.created_at,
    t.id AS g_id, t.title AS g_title, t.description AS g_description,
    t.completed AS g_completed, t.due_date AS g_due_date,
    t.user_id AS g_user_id, t.parent_todo_id AS g_parent_todo_id,
    t.created_at AS g_created_at, t.updated_at AS g_updated_at,
    u.id AS owner_id_val, u.name AS owner_name, u.email AS owner_email
FROM group_shares gs
JOIN todos t ON t.id = gs.group_id
LEFT JOIN todos sub ON sub.parent_todo_id = t.id
JOIN users u ON u.id = t.user_id
WHERE gs.shared_with_user_id = ?
GROUP BY gs.id, t.id, u.id;

-- name: CountGroupMembers :one
SELECT COUNT(*) FROM group_shares WHERE group_id = ?;

-- name: DeleteGroupShare :exec
DELETE FROM group_shares
WHERE group_id = ? AND shared_with_user_id = ?;

-- name: UpdateGroupSharePermission :exec
UPDATE group_shares
SET permission = ?
WHERE group_id = ? AND shared_with_user_id = ?;
