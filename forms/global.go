package forms

import (
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var isMongoId validator.Func = func(fl validator.FieldLevel) bool {
	field, ok := fl.Field().Interface().(string)
	if ok {
		_, err := primitive.ObjectIDFromHex(field)
		return err == nil
	}
	return true
}
