package forms

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

type ReactionForm struct {
	Reaction string `json:"reaction" binding:"required,isReaction"`
}

var isReaction validator.Func = func(fl validator.FieldLevel) bool {
	field, ok := fl.Field().Interface().(string)
	if ok {
		reg := regexp.MustCompile(`^[\p{So}\p{Sk}\p{Sm}\p{Sc}\p{Pd}\p{Zs}\x{1F1E6}-\x{1F1FF}\x{1F300}-\x{1F5FF}\x{1F600}-\x{1F64F}\x{1F680}-\x{1F6FF}\x{1F700}-\x{1F77F}\x{1F780}-\x{1F7FF}\x{1F800}-\x{1F8FF}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}\x{02702}-\x{027B0}\x{02600}-\x{027BF}\x{1F680}-\x{1F6C0}\x{024C2}-\x{1F251}\x{1F300}-\x{1F5FF}]+$`)
		return reg.MatchString(field)
	}
	return true
}
