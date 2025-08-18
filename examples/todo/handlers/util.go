package handlers

import (
	"strconv"

	"github.com/sktylr/routeit"
	"github.com/sktylr/routeit/examples/todo/dao"
)

func userIdFromRequest(req *routeit.Request) string {
	rawUser, _ := req.ContextValue("user")
	user, ok := rawUser.(*dao.User)
	if !ok {
		// Due to the middleware we have in place, this should never happen. To
		// comply with the type signature, we return an empty string, though it
		// could be more accurate to panic as this is highly unexpected.
		return ""
	}
	return user.Id
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
