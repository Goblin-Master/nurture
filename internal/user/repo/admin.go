package repo

import (
	"context"
	"nurture/internal/user/repo/dao"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type AdminUserRow struct {
	UserID   string
	Username string
	Avatar   string
}

func toAdminUserRow(row dao.AdminListUsersRow) AdminUserRow {
	return AdminUserRow{
		UserID:   row.UserID,
		Username: row.Username,
		Avatar:   row.Avatar,
	}
}

func (ur *UserRepo) AdminListUsers(ctx context.Context, keyword string, page, pageSize int) ([]AdminUserRow, bool, error) {
	limit := int32(pageSize + 1)
	offset := int32((page - 1) * pageSize)
	rows, err := ur.userDao.AdminListUsers(ctx, dao.AdminListUsersParams{
		Column1: keyword,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		ur.log.Error(err)
		return nil, false, ErrDefault
	}
	hasMore := false
	if len(rows) > pageSize {
		hasMore = true
		rows = rows[:pageSize]
	}
	ret := make([]AdminUserRow, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, toAdminUserRow(row))
	}
	return ret, hasMore, nil
}

func (ur *UserRepo) AdminUpdateUserRole(ctx context.Context, userID string, role int16) error {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return err
	}
	n, err := ur.userDao.AdminUpdateUserRole(ctx, dao.AdminUpdateUserRoleParams{
		UserID: uid,
		Role:   role,
		Utime:  time.Now().UnixMilli(),
	})
	if err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	if n == 0 {
		return ErrUserNotExist
	}
	return nil
}
