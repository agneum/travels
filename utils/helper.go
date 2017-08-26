package utils

import (
	"strconv"

	routing "github.com/qiangxue/fasthttp-routing"
)

func ResponseWithJSON(ctx *routing.Context, json []byte, code int) {
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(code)
	ctx.SetBody(json)
}

func ParseIdParameter(parameter interface{}) (id uint64, err error) {
	stringID, ok := parameter.(string)
	if !ok {
		return
	}

	id, err = strconv.ParseUint(stringID, 10, 32)
	if err != nil {
		return
	}

	return id, nil
}
