// The package controller contains all the HTTP request handlers
package controller

import (
	"github.com/asiragusa/wschat/response"
	"github.com/kataras/iris"
	"github.com/kataras/iris/context"
)

func sendResponse(ctx context.Context, response response.Response) {
	ctx.StatusCode(response.GetCode())
	if response.GetCode() != iris.StatusNoContent {
		ctx.JSON(response)
	}
}
