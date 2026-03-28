// Package fieldmap aims to aid develops who often have to map multiple data-sources with different or messy datastructures into uniform structs / formats reliably.
package fieldmap

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type TransformFunc func(any) (any, error)

type Rule struct {
	From      string
	To        string
	Required  bool
	Transform TransformFunc
}

type Config struct {
	Rules []Rule
}

type FieldError struct {
	Field string
	Err   error
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %v", e.Field, e.Err)
}

type MultiError []FieldError

func (m MultiError) Error() string {
	var parts []string
	for _, err := range m {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "; ")
}

func Apply(cfg Config, input map[string]any, out any) error {
	var errs MultiError

	for _, rule := range cfg.Rules {
		val, ok := getByPath(input, rule.From)
		if !ok {
			if rule.Required {
				errs = append(errs, FieldError{
					Field: rule.To,
					Err:   fmt.Errorf("required field missing from path %q", rule.From),
				})
			}
			continue
		}

		var err error
		if rule.Transform != nil {
			val, err = rule.Transform(val)
			if err != nil {
				errs = append(errs, FieldError{
					Field: rule.To,
					Err:   fmt.Errorf("transform failed: %w", err),
				})
				continue
			}
		}

		if err := setField(out, rule.To, val); err != nil {
			errs = append(errs, FieldError{
				Field: rule.To,
				Err:   err,
			})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func getByPath(input map[string]any, path string) (any, bool) {
	parts := strings.Split(path, ".")
	var current any = input

	for _, part := range parts {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func setField(out any, fieldName string, value any) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("out must be a pointer to struct")
	}

	field := rv.Elem().FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("unknown field %q", fieldName)
	}
	if !field.CanSet() {
		return fmt.Errorf("cannot set field %q", fieldName)
	}

	val := reflect.ValueOf(value)

	if val.Type().AssignableTo(field.Type()) {
		field.Set(val)
		return nil
	}

	if val.Type().ConvertibleTo(field.Type()) {
		field.Set(val.Convert(field.Type()))
		return nil
	}

	return fmt.Errorf("cannot assign %T to field %q of type %s", value, fieldName, field.Type())
}

func TrimLower() TransformFunc {
	return func(v any) (any, error) {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", v)
		}
		return strings.ToLower(strings.TrimSpace(s)), nil
	}
}

func StringToBool() TransformFunc {
	return func(v any) (any, error) {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", v)
		}

		switch strings.ToLower(strings.TrimSpace(s)) {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no":
			return false, nil
		default:
			return nil, fmt.Errorf("invalid boolean value %q", s)
		}
	}
}

func ParseTime(layout string) TransformFunc {
	return func(v any) (any, error) {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", v)
		}
		return time.Parse(layout, strings.TrimSpace(s))
	}
}
