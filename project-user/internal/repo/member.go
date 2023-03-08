package repo

import (
	"context"
	"github.com/Kirov7/project-user/internal/data/member"
	database "github.com/Kirov7/project-user/internal/database"
)

type MemberRepo interface {
	GetMemberByEmail(ctx context.Context, email string) (bool, error)
	GetMemberByAccount(ctx context.Context, account string) (bool, error)
	GetMemberByMobile(ctx context.Context, mobile string) (bool, error)
	FindMember(ctx context.Context, account, pwd string) (*member.Member, error)
	SaveMember(conn database.DbConn, ctx context.Context, mem *member.Member) error
	FindMemberById(ctx context.Context, id int64) (*member.Member, error)
	FindMemberByIds(background context.Context, ids []int64) ([]*member.Member, error)
}
