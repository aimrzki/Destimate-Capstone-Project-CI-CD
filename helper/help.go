package helper

import (
	"errors"
	"strconv"
)

func ConvertParamToUint(param string) (uint, error) {
	id, err := strconv.ParseUint(param, 10, 32)
	if err != nil {
		return 0, errors.New("Invalid parameter")
	}
	return uint(id), nil
}
