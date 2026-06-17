package database

import (
	"context"
	"database/sql"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers — NULL type conversions (exported for repository use)
// ─────────────────────────────────────────────────────────────────────────────

func ToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func NullTimeTo(t sql.NullTime) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func ToNullInt32(v *uint) sql.NullInt32 {
	if v == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: int32(*v), Valid: true}
}

func NullInt32ToUint(v sql.NullInt32) *uint {
	if !v.Valid {
		return nil
	}
	u := uint(v.Int32)
	return &u
}

// ─────────────────────────────────────────────────────────────────────────────
// Users
// ─────────────────────────────────────────────────────────────────────────────

type CreateUserParams struct {
	Name     string
	Email    string
	Password string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (sql.Result, error) {
	return q.db.ExecContext(ctx,
		`INSERT INTO users (name, email, password, created_at) VALUES (?, ?, ?, NOW())`,
		arg.Name, arg.Email, arg.Password,
	)
}

const getUserByEmail = `SELECT id, name, email, password, created_at FROM users WHERE email = ? LIMIT 1`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)
	return u, err
}

const getUserByID = `SELECT id, name, email, password, created_at FROM users WHERE id = ? LIMIT 1`

func (q *Queries) GetUserByID(ctx context.Context, id uint32) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByID, id)
	var u User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt)
	return u, err
}

const searchUsers = `
SELECT id, name, email, password, created_at FROM users
WHERE id != ? AND (name LIKE ? OR email LIKE ?) LIMIT 20`

func (q *Queries) SearchUsers(ctx context.Context, excludeID uint32, term string) ([]User, error) {
	rows, err := q.db.QueryContext(ctx, searchUsers, excludeID, term, term)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

type UpdateUserPasswordParams struct {
	Password string
	ID       uint32
}

func (q *Queries) UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error {
	_, err := q.db.ExecContext(ctx, `UPDATE users SET password = ? WHERE id = ?`, arg.Password, arg.ID)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// Todos
// ─────────────────────────────────────────────────────────────────────────────

type CreateTodoParams struct {
	Title        string
	Description  string
	Completed    bool
	DueDate      sql.NullTime
	UserID       uint32
	ParentTodoID sql.NullInt32
}

func (q *Queries) CreateTodo(ctx context.Context, arg CreateTodoParams) (sql.Result, error) {
	return q.db.ExecContext(ctx,
		`INSERT INTO todos (title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())`,
		arg.Title, arg.Description, arg.Completed, arg.DueDate, arg.UserID, arg.ParentTodoID,
	)
}

const todoColumns = `id, title, description, completed, due_date, user_id, parent_todo_id, created_at, updated_at`

func scanTodo(s interface{ Scan(...interface{}) error }) (Todo, error) {
	var t Todo
	err := s.Scan(&t.ID, &t.Title, &t.Description, &t.Completed,
		&t.DueDate, &t.UserID, &t.ParentTodoID, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func queryTodos(rows *sql.Rows, err error) ([]Todo, error) {
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var todos []Todo
	for rows.Next() {
		t, err := scanTodo(rows)
		if err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, rows.Err()
}

func (q *Queries) GetTodoByID(ctx context.Context, id uint32) (Todo, error) {
	return scanTodo(q.db.QueryRowContext(ctx, `SELECT `+todoColumns+` FROM todos WHERE id = ? LIMIT 1`, id))
}

func (q *Queries) GetGroupByID(ctx context.Context, id uint32) (Todo, error) {
	return scanTodo(q.db.QueryRowContext(ctx,
		`SELECT `+todoColumns+` FROM todos WHERE id = ? AND parent_todo_id IS NULL LIMIT 1`, id))
}

func (q *Queries) GetSubtasksByParentID(ctx context.Context, parentID uint32) ([]Todo, error) {
	return queryTodos(q.db.QueryContext(ctx,
		`SELECT `+todoColumns+` FROM todos WHERE parent_todo_id = ? ORDER BY created_at ASC`, parentID))
}

func (q *Queries) GetSubtasksByParentAndUser(ctx context.Context, parentID, userID uint32) ([]Todo, error) {
	return queryTodos(q.db.QueryContext(ctx,
		`SELECT `+todoColumns+` FROM todos WHERE parent_todo_id = ? AND user_id = ? ORDER BY created_at ASC`,
		parentID, userID))
}

// GetGroupsByUserIDParams holds parameters for the dynamic groups listing query.
// Search is the LIKE pattern (e.g. "%term%") — empty string means no filter.
type GetGroupsByUserIDParams struct {
	UserID uint32
	Search string
}

func (q *Queries) GetGroupsByUserID(ctx context.Context, arg GetGroupsByUserIDParams) ([]Todo, error) {
	query := `SELECT ` + todoColumns + ` FROM todos WHERE user_id = ? AND parent_todo_id IS NULL`
	args := []interface{}{arg.UserID}
	if arg.Search != "" {
		query += ` AND (title LIKE ? OR description LIKE ?)`
		args = append(args, arg.Search, arg.Search)
	}
	query += ` ORDER BY created_at DESC`
	return queryTodos(q.db.QueryContext(ctx, query, args...))
}

type UpdateTodoParams struct {
	Title       string
	Description string
	Completed   bool
	DueDate     sql.NullTime
	ID          uint32
}

func (q *Queries) UpdateTodo(ctx context.Context, arg UpdateTodoParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE todos SET title = ?, description = ?, completed = ?, due_date = ?, updated_at = NOW() WHERE id = ?`,
		arg.Title, arg.Description, arg.Completed, arg.DueDate, arg.ID)
	return err
}

func (q *Queries) DeleteTodo(ctx context.Context, id uint32) error {
	_, err := q.db.ExecContext(ctx, `DELETE FROM todos WHERE id = ?`, id)
	return err
}

// ─────────────────────────────────────────────────────────────────────────────
// Group Shares
// ─────────────────────────────────────────────────────────────────────────────

type CreateGroupShareParams struct {
	GroupID          uint32
	OwnerID          uint32
	SharedWithUserID uint32
	Permission       string
}

func (q *Queries) CreateGroupShare(ctx context.Context, arg CreateGroupShareParams) (sql.Result, error) {
	return q.db.ExecContext(ctx,
		`INSERT INTO group_shares (group_id, owner_id, shared_with_user_id, permission, created_at) VALUES (?, ?, ?, ?, NOW())`,
		arg.GroupID, arg.OwnerID, arg.SharedWithUserID, arg.Permission)
}

func (q *Queries) GetGroupShare(ctx context.Context, groupID, sharedWithUserID uint32) (GroupShare, error) {
	row := q.db.QueryRowContext(ctx,
		`SELECT id, group_id, owner_id, shared_with_user_id, permission, created_at
		 FROM group_shares WHERE group_id = ? AND shared_with_user_id = ? LIMIT 1`,
		groupID, sharedWithUserID)
	var gs GroupShare
	err := row.Scan(&gs.ID, &gs.GroupID, &gs.OwnerID, &gs.SharedWithUserID, &gs.Permission, &gs.CreatedAt)
	return gs, err
}

func (q *Queries) GetGroupSharesByGroupID(ctx context.Context, groupID uint32) ([]GetGroupSharesByGroupIDRow, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT gs.id, gs.group_id, gs.owner_id, gs.shared_with_user_id, gs.permission, gs.created_at,
		        u.id, u.name, u.email
		 FROM group_shares gs
		 JOIN users u ON u.id = gs.shared_with_user_id
		 WHERE gs.group_id = ?`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []GetGroupSharesByGroupIDRow
	for rows.Next() {
		var r GetGroupSharesByGroupIDRow
		if err := rows.Scan(&r.ID, &r.GroupID, &r.OwnerID, &r.SharedWithUserID, &r.Permission, &r.CreatedAt,
			&r.SwID, &r.SwName, &r.SwEmail); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (q *Queries) GetSharesByUserID(ctx context.Context, userID uint32) ([]GroupShare, error) {
	rows, err := q.db.QueryContext(ctx,
		`SELECT id, group_id, owner_id, shared_with_user_id, permission, created_at
		 FROM group_shares WHERE shared_with_user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var shares []GroupShare
	for rows.Next() {
		var gs GroupShare
		if err := rows.Scan(&gs.ID, &gs.GroupID, &gs.OwnerID, &gs.SharedWithUserID, &gs.Permission, &gs.CreatedAt); err != nil {
			return nil, err
		}
		shares = append(shares, gs)
	}
	return shares, rows.Err()
}

func (q *Queries) CountGroupMembers(ctx context.Context, groupID uint32) (int64, error) {
	var count int64
	err := q.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM group_shares WHERE group_id = ?`, groupID).Scan(&count)
	return count, err
}

func (q *Queries) DeleteGroupShare(ctx context.Context, groupID, sharedWithUserID uint32) error {
	_, err := q.db.ExecContext(ctx,
		`DELETE FROM group_shares WHERE group_id = ? AND shared_with_user_id = ?`,
		groupID, sharedWithUserID)
	return err
}

type UpdateGroupSharePermissionParams struct {
	Permission       string
	GroupID          uint32
	SharedWithUserID uint32
}

func (q *Queries) UpdateGroupSharePermission(ctx context.Context, arg UpdateGroupSharePermissionParams) error {
	_, err := q.db.ExecContext(ctx,
		`UPDATE group_shares SET permission = ? WHERE group_id = ? AND shared_with_user_id = ?`,
		arg.Permission, arg.GroupID, arg.SharedWithUserID)
	return err
}
