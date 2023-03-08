package domain

import (
	"context"
	loginRpc "github.com/Kirov7/project-grpc/user/login"
	rpc "github.com/Kirov7/project-project/internal/rpc"
	"time"
)

type UserRpcDomain struct {
	lc loginRpc.LoginServiceClient
}

func NewUserRpcDomain() *UserRpcDomain {
	return &UserRpcDomain{
		lc: rpc.LoginServiceClient,
	}
}
func (d *UserRpcDomain) MemberList(mIdList []int64) ([]*loginRpc.MemberMessage, map[int64]*loginRpc.MemberMessage, error) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	messageList, err := d.lc.FindMemInfoByIds(c, &loginRpc.MemIdRequest{MemIds: mIdList})
	mMap := make(map[int64]*loginRpc.MemberMessage)
	for _, v := range messageList.MemberList {
		mMap[v.Id] = v
	}
	return messageList.MemberList, mMap, err
}

func (d *UserRpcDomain) MemberInfo(ctx context.Context, memberCode int64) (*loginRpc.MemberMessage, error) {
	c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	memberMsh, err := d.lc.FindMemInfoById(c, &loginRpc.MemIdRequest{MemId: memberCode})
	return memberMsh, err
}
