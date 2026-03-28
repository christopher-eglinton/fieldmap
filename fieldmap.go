// Package fieldmap provides utilities for mapping data from external sources
// (e.g. APIs, CSVs) into strongly-typed Go structs using declarative rules.
//
// It is designed for scenarios where input data is inconsistent, nested, or
// differs across multiple sources, and needs to be normalized into a uniform format.
package fieldmap

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

// TransformFunc defines a function that transforms a value during mapping.
// It receives an input value and returns a transformed value or an error.
type TransformFunc func(any) (any, error)

// Rule defines a mapping rule from a source path to a destination struct field.
//
// From specifies the dot-path in the input map (e.g. "employee.givenName").
// To specifies the destination struct field name.
// Required indicates whether the field must exist in the input.
// Transform optionally applies a transformation to the value before assignment.
type Rule struct {
	From      string
	To        string
	Required  bool
	Transform TransformFunc
}

// Config contains the set of mapping rules to apply.
type Config struct {
	Rules []Rule
}

// FieldError represents an error associated with a specific destination field.
type FieldError struct {
	Field string
	Err   error
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s: %v", e.Field, e.Err)
}

// MultiError aggregates multiple field errors into a single error.
type MultiError []FieldError

// Error implements the error interface by joining all field errors.
func (m MultiError) Error() string {
	var parts []string
	for _, err := range m {
		parts = append(parts, err.Error())
	}
	return strings.Join(parts, "; ")
}

// Apply executes the mapping defined in cfg, reading values from input
// and assigning them to the struct pointed to by out.
//
// input must be a map[string]any representing the source data.
// out must be a pointer to a struct.
//
// For each rule:
// - The value is retrieved using the From path.
// - If Required is true and the value is missing, an error is recorded.
// - If a Transform is defined, it is applied.
// - The resulting value is assigned to the struct field specified by To.
//
// If one or more errors occur, a MultiError is returned.
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

// getByPath retrieves a value from a nested map using a dot-separated path.
//
// Example:
// path: "employee.givenName"
// input: map[string]any{"employee": {"givenName": "John"}}
//
// Returns the value and true if found, otherwise nil and false.
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

// setField assigns a value to a struct field by name using reflection.
//
// out must be a pointer to a struct.
// fieldName must match an exported struct field.
//
// The function attempts:
// - direct assignment if types match
// - conversion if types are convertible
//
// Returns an error if the field does not exist, cannot be set, or types are incompatible.
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

// TrimLower returns a TransformFunc that trims whitespace and converts a string to lowercase.
func TrimLower() TransformFunc {
	return func(v any) (any, error) {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", v)
		}
		return strings.ToLower(strings.TrimSpace(s)), nil
	}
}

// StringToBool returns a TransformFunc that converts common string values to a boolean.
//
// Accepted truthy values: "true", "1", "yes"
// Accepted falsy values:  "false", "0", "no"
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

// ParseTime returns a TransformFunc that parses a string into time.Time
// using the provided layout (e.g. "2006-01-02").
func ParseTime(layout string) TransformFunc {
	return func(v any) (any, error) {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("expected string, got %T", v)
		}
		return time.Parse(layout, strings.TrimSpace(s))
	}
}
