package forms

import (
	"log"
	"regexp"

	"github.com/go-playground/validator/v10"
)

type RepositoryForm struct {
	Name        string `json:"name" binding:"required,max=100,isRepositoryName"`
	Description string `json:"description" binding:"max=300"`
	Access      string `json:"access" binding:"required,isValidAccess"`
}

type UpdateRepositoryForm struct {
	Description  string   `json:"description" binding:"max=300"`
	Content      string   `json:"content"`
	Access       string   `json:"access" binding:"isValidAccess"`
	CustomAccess []string `json:"custom_access"`
	Tags         []string `json:"tags"`
}

var isRepositoryName validator.Func = func(fl validator.FieldLevel) bool {
	name, ok := fl.Field().Interface().(string)
	if ok {
		r, err := regexp.Compile("^[0-9a-z_]+$")
		if err != nil {
			log.Println("Error al compilar el regex", err)
			return false
		}
		return r.MatchString(name)
	}
	return true
}

var isValidAccess validator.Func = func(fl validator.FieldLevel) bool {
	access, ok := fl.Field().Interface().(string)
	if ok {
		return access == "public" || access == "private" || access == "private-group"
	}
	return true
}
