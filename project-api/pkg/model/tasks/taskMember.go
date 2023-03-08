package tasks

type TaskMember struct {
	Id                int64  `json:"id"`
	Name              string `json:"name"`
	Avatar            string `json:"avatar"`
	Code              string `json:"Code"`
	IsExecutor        int    `json:"is_executor"`
	IsOwner           int    `json:"is_owner"`
	MemberAccountCode string `json:"member_account_code"`
}
