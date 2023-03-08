package repo

import (
	"context"
	"github.com/Kirov7/project-user/internal/data/organization"
	database "github.com/Kirov7/project-user/internal/database"
)

type OrganizationRepo interface {
	SaveOrganization(conn database.DbConn, ctx context.Context, org *organization.Organization) error
	FindOrganizationByMemId(ctx context.Context, memId int64) ([]organization.Organization, error)
}
