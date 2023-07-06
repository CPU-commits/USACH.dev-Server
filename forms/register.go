package forms

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func Init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("isRepositoryName", isRepositoryName)
		v.RegisterValidation("isValidAccess", isValidAccess)
		v.RegisterValidation("isLinkType", isLinkType)
		v.RegisterValidation("isMongoId", isMongoId)
		v.RegisterValidation("isReaction", isReaction)
	}
}
