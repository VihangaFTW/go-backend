package api

import (
	"github.com/VihangaFTW/Go-Backend/db/util"
	"github.com/go-playground/validator/v10"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	// ok variable will be true if the type assertion to string was successful
	//? checking whether the field value to be validated is a string
	if currency, ok := fl.Field().Interface().(string); ok {
		// check if the string is a supported currency
		return util.IsSupportedCurrency(currency)
	}
	
	

	return false
}


