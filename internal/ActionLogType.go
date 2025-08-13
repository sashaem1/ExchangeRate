package internal

import (
	"fmt"
	"strings"
)

var BaseActionLogType = map[string]struct{}{
	"PAIR": {},
	"DATE": {},
}

type ActionLogType struct {
	Action string
}

func NewActionLogType(action string) (ActionLogType, error) {
	op := "internal.logDbType.NewActionLogType"
	action = strings.ToUpper(strings.TrimSpace(action))

	if len(action) != 4 {
		err := fmt.Sprintf("Название действия должно состоять из 4 символов")
		return ActionLogType{}, fmt.Errorf("%s: %s", op, err)
	}

	if _, ok := BaseActionLogType[action]; !ok {
		err := fmt.Sprintf("Отсутствует такое действие в системе: %s", action)
		return ActionLogType{}, fmt.Errorf("%s: %s", op, err)
	}

	return ActionLogType{Action: action}, nil
}
