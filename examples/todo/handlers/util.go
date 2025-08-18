package handlers

import (
	"strconv"

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

func queryParamOrDefault(param string, qs *routeit.QueryParams, def int) (int, error) {
	val, hasVal, err := qs.Only(param)
	if err != nil {
		return 0, err
	}

	valNum := def
	if hasVal {
		valNum, err = strconv.Atoi(val)
		if err != nil || valNum <= 0 {
			return 0, routeit.ErrBadRequest().
				WithMessagef("%#q is not a valid %s number", param, val).
				WithCause(err)
		}
	}
	return valNum, nil
}
