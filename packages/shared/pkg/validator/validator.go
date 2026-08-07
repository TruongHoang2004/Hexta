package validator

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
	log "gitlab.com/ecommercehub1/shared/pkg/logger"
)

func NewValidator() *validator.Validate {
	return validator.New()
}

func RegisterDecimalTypeFunc(validator *validator.Validate) {
	validator.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if valuer, ok := field.Interface().(decimal.Decimal); ok {
			return valuer.String()
		}
		return nil
	}, decimal.Decimal{})
}

func registerValidation(validator *validator.Validate, logger *log.Logger, tag string, fn validator.Func) {
	if err := validator.RegisterValidation(tag, fn); err != nil {
		logger.GetZap().Fatalf("Register custom validation %s failed with error: %s", tag, err.Error())
	}
}

func RegisterStructValidation(validator *validator.Validate, fn validator.StructLevelFunc, in ...interface{}) {
	validator.RegisterStructValidation(fn, in...)
}

func RegisterValidations(validator *validator.Validate, logger *log.Logger) {
	registerValidation(validator, logger, "valid-decimal-range", validDecimalRange)
	registerValidation(validator, logger, "valid-einvoice-vat-rate-name", validEInvoiceVATRateNameValidation(logger))
	registerValidation(validator, logger, "valid-by-regex-pattern", validByRegexPattern)
	registerValidation(validator, logger, "valid-today", validToday)
	registerValidation(validator, logger, "valid-vat-rate", validVatRate)

}

// einvoiceVATRateNameRegex Mẫu tên loại thuế suất.
// 0%: Thuế suất 0%
// 5%: Thuế suất 5%
// 10%: Thuế suất 10%
// KCT: Không chịu thuế GTGT
// KKKNT: Không kê khai, tính nộp thuế GTGT
// KHAC:AB.CD%: Trường hợp khác, với “:AB.CD” là bắt buộc trong trường hợp xác định được giá trị thuế suất. A, B, C, D là các số nguyên từ 0 đến 9.
// Ví dụ: KHAC:AB.CD%
const eInvoiceVATRateNameRegex = `^((?:0|5|(?:10)|(?:KHAC:[0-9]{0,2}\.?[0-9]{0,2}))%)|(KCT|KKKNT)$`

func validEInvoiceVATRateNameValidation(logger *log.Logger) validator.Func {
	return func(fl validator.FieldLevel) bool {
		vatRateName, ok := fl.Field().Interface().(string)
		if !ok {
			return false
		}

		match, err := regexp.MatchString(eInvoiceVATRateNameRegex, vatRateName)
		if err != nil {
			logger.GetZap().Errorf("EInvoiceVATRateName doesn't match with regex: %+v", err)
		}

		return match
	}
}

var validByRegexPattern validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	regexPattern := fl.Param()
	match, err := regexp.MatchString(regexPattern, value)
	if err != nil {
		return false
	}
	return match
}

var validDecimalRange validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	decimalValue, err := decimal.NewFromString(value)
	if err != nil {
		return false
	}
	params := strings.Split(fl.Param(), "_")
	if len(params) != 2 {
		return false
	}
	sizeCompare, _ := strconv.Atoi(params[0])
	dCompare, _ := strconv.Atoi(params[1])
	return validateDecimalRange(decimalValue.Abs(), sizeCompare, dCompare)
}

// Size chỉ định số lượng chữ số tối đa.
// D chỉ định số lượng chữ số tối đa của vị trí bên phải dấu thập phân.
func validateDecimalRange(decimalNumber decimal.Decimal, sizeCompare, dCompare int) bool {
	decimalNumberStrings := strings.Split(decimalNumber.String(), ".")
	d := 0
	if len(decimalNumberStrings) > 1 {
		d = len(decimalNumberStrings[1])
	}
	size := len(decimalNumberStrings[0]) + d
	if size > sizeCompare || d > dCompare {
		return false
	}
	return true
}

var validToday validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}

	date := value.Format("2006-01-02")
	today := time.Now().Format("2006-01-02")
	return date == today
}

var validVatRate validator.Func = func(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(float64)
	if !ok {
		return false
	}
	return value == -1 || value == -2 || value >= 0
}
