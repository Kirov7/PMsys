package data

import (
	"github.com/Kirov7/project-common/encrypts"
	"github.com/Kirov7/project-common/tms"
	"github.com/Kirov7/project-project/pkg/model"
)

type Project struct {
	Id                 int64
	Cover              string
	Name               string
	Description        string
	AccessControlType  int
	WhiteList          string
	Sort               int
	Deleted            int
	TemplateCode       int
	Schedule           float64
	CreateTime         int64
	OrganizationCode   int64
	DeletedTime        string
	Private            int
	Prefix             string
	OpenPrefix         int
	Archive            int
	ArchiveTime        int64
	OpenBeginTime      int
	OpenTaskPrivate    int
	TaskBoardTheme     string
	BeginTime          int64
	EndTime            int64
	AutoUpdateSchedule int
}

func (*Project) TableName() string {
	return "project"
}

func ToProjectMap(list []*Project) map[int64]*Project {
	m := make(map[int64]*Project, len(list))
	for _, v := range list {
		m[v.Id] = v
	}
	return m
}

type ProjectMember struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64 `gorm:"column:member_code"`
	JoinTime    int64
	IsOwner     int64
	Authorize   string
}

func (*ProjectMember) TableName() string {
	return "project_member"
}

type ProjectAndMember struct {
	Project
	ProjectCode int64
	MemberCode  int64 `gorm:"column:member_code"`
	JoinTime    int64
	IsOwner     int64
	Authorize   string
	OwnerName   string
	Collected   int
}

func (m *Project) GetAccessControlType() string {
	if m.AccessControlType == 0 {
		return "open"
	}
	if m.AccessControlType == 1 {
		return "private"
	}
	if m.AccessControlType == 2 {
		return "custom"
	}
	return ""
}

func ToMap(orgs []*ProjectAndMember) map[int64]*ProjectAndMember {
	m := make(map[int64]*ProjectAndMember)
	for _, v := range orgs {
		m[v.Id] = v
	}
	return m
}

type ProjectCollection struct {
	Id          int64
	ProjectCode int64
	MemberCode  int64
	CreateTime  int64
}

func (*ProjectCollection) TableName() string {
	return "project_collection"
}

type ProjectTemplate struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       int64
	OrganizationCode int64
	Cover            string
	MemberCode       int64
	IsSystem         int
}

func (*ProjectTemplate) TableName() string {
	return "project_template"
}

type ProjectTemplateAll struct {
	Id               int
	Name             string
	Description      string
	Sort             int
	CreateTime       string
	OrganizationCode string
	Cover            string
	MemberCode       string
	IsSystem         int
	TaskStages       []*TaskStagesOnlyName
	Code             string
}

func (pt *ProjectTemplate) Convert(taskStages []*TaskStagesOnlyName) *ProjectTemplateAll {
	organizationCode, _ := encrypts.EncryptInt64(pt.OrganizationCode, model.AESKey)
	memberCode, _ := encrypts.EncryptInt64(pt.MemberCode, model.AESKey)
	code, _ := encrypts.EncryptInt64(int64(pt.Id), model.AESKey)
	pta := &ProjectTemplateAll{
		Id:               pt.Id,
		Name:             pt.Name,
		Description:      pt.Description,
		Sort:             pt.Sort,
		CreateTime:       tms.FormatByMill(pt.CreateTime),
		OrganizationCode: organizationCode,
		Cover:            pt.Cover,
		MemberCode:       memberCode,
		IsSystem:         pt.IsSystem,
		TaskStages:       taskStages,
		Code:             code,
	}
	return pta
}

func ToProjectTemplateIds(pts []ProjectTemplate) []int {
	var ids []int
	for _, v := range pts {
		ids = append(ids, v.Id)
	}
	return ids
}
