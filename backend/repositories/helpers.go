package repositories

import (
	"database/sql"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers — NULL type conversions
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
