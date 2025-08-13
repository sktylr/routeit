package handlers

import (
	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
)

func userIdFromRequest(req *routeit.Request) (string, bool) {
	rawUser, _ := req.ContextValue("user")
	user, ok := rawUser.(*dao.User)
	if !ok {
		return "", false
	}
	return user.Id, true
}
