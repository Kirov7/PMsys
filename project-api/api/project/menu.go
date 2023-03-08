package project

import (
	"context"
	"github.com/Kirov7/project-api/pkg/model/menu"
	common "github.com/Kirov7/project-common"
	"github.com/Kirov7/project-common/errs"
	menuRpc "github.com/Kirov7/project-grpc/project/menu"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"time"
)

type HandlerMenu struct {
}

func NewMenu() *HandlerMenu {
	return &HandlerMenu{}
}

func (d *HandlerMenu) menuList(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := MenuServiceClient.MenuList(ctx, &menuRpc.MenuRequest{})
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	var list []*menu.Menu
	copier.Copy(&list, res.List)
	if list == nil {
		list = []*menu.Menu{}
	}
	c.JSON(http.StatusOK, result.Success(list))
}
