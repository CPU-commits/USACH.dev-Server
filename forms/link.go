package forms

import (
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"github.com/go-playground/validator/v10"
)

type LinkForm struct {
	Type  string `json:"type" binding:"required,isLinkType"`
	Title string `json:"title" binding:"required,max=30"`
	Link  string `json:"link" binding:"required,startswith=http"`
}

var isLinkType validator.Func = func(fl validator.FieldLevel) bool {
	typeLink, ok := fl.Field().Interface().(string)
	if ok {
		linkTypes := []string{
			"drive",
			"github",
			"cloud",
			"pdf",
			"youtube",
			"other",
		}
		match, err := utils.AnyMatch(linkTypes, func(x interface{}) bool {
			tl := x.(string)
			return tl == typeLink
		})
		if !match || err != nil {
			return false
		}
	}
	return true
}
