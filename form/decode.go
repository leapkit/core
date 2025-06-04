// Package decode provides form decoding for nested, pointer-based, and embedded structs.
package form

import (
	"cmp"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

func Decode(r *http.Request, dst any) error {
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			return err
		}
	} else if err := r.ParseForm(); err != nil {
		return err
	}

	if len(r.Form) == 0 {
		r.Form = r.URL.Query()
	}

	t := reflect.TypeOf(dst)
	v := reflect.ValueOf(dst)

	if t.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("dst must be pointer to struct")
	}

	return decodeForm(v.Elem(), r.Form, "")
}

func decodeForm(dst reflect.Value, form map[string][]string, prefix string) error {
	t := dst.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		fieldVal := dst.Field(i)

		tagName := field.Tag.Get("form")

		if field.PkgPath != "" || tagName == "-" || !fieldVal.CanSet() {
			continue
		}

		fName := cmp.Or(tagName, field.Name)
		fType := field.Type
		fKind := fType.Kind()
		isAnon := field.Anonymous
		isPtr := fType.Kind() == reflect.Ptr

		if prefix != "" && !isAnon {
			fName = prefix + "." + fName
		}

		if isPtr {
			fType = fType.Elem()
			fKind = fType.Kind()

			if _, ok := form[fName]; !ok && fKind != reflect.Struct {
				continue
			}

			if fieldVal.IsNil() {
				fieldVal.Set(reflect.New(fType))
			}

			fieldVal = fieldVal.Elem()
		}

		switch fKind {
		case reflect.Struct:
			subPrefix := fName
			if isAnon {
				subPrefix = prefix
			}

			if dec, ok := customDecoders[fType]; ok {
				values, ok := form[fName]
				if !ok || len(values) == 0 {
					continue
				}

				v, err := dec(values[0])
				if err != nil {
					return fmt.Errorf("cannot convert to %q: %w", fType, err)
				}

				fieldVal.Set(reflect.ValueOf(v))

				return nil
			}

			if err := decodeForm(fieldVal, form, subPrefix); err != nil {
				return err
			}
		case reflect.Slice:
			elemType := fType.Elem()
			isPtr := elemType.Kind() == reflect.Ptr
			baseType := elemType
			if isPtr {
				baseType = elemType.Elem()
			}

			if baseType.Kind() == reflect.Struct {
				slice := reflect.MakeSlice(fType, 0, 0)
				index := 0
				for {
					subKey := fmt.Sprintf("%s[%d]", fName, index)
					var found bool
					for k := range form {
						if strings.HasPrefix(k, subKey+".") {
							found = true
							break
						}
					}
					if !found {
						break
					}

					el := reflect.New(baseType).Elem()
					if err := decodeForm(el, form, subKey); err != nil {
						return err
					}

					if isPtr {
						slice = reflect.Append(slice, el.Addr())
					} else {
						slice = reflect.Append(slice, el)
					}
					index++
				}
				fieldVal.Set(slice)
				continue
			}

			values, ok := form[fName]
			if !ok {
				continue
			}

			slice := reflect.MakeSlice(fType, len(values), len(values))
			for i, val := range values {
				el := slice.Index(i)
				if isPtr {
					ptr := reflect.New(baseType).Elem()
					if err := decodeField(ptr, val); err != nil {
						return err
					}
					ref := reflect.New(baseType)
					ref.Elem().Set(ptr)
					el.Set(ref)
				} else {
					if err := decodeField(el, val); err != nil {
						return err
					}
				}
			}
			fieldVal.Set(slice)
		default:
			if values, ok := form[fName]; ok && len(values) > 0 {
				if err := decodeField(fieldVal, values[0]); err != nil {
					return fmt.Errorf("failed to set %s: %w", fName, err)
				}
			}
		}
	}
	return nil
}

func decodeField(fieldValue reflect.Value, value string) error {
	if dec, ok := customDecoders[fieldValue.Type()]; ok {
		v, err := dec(value)
		if err != nil {
			return fmt.Errorf("cannot convert to %q: %w", fieldValue.Type(), err)
		}

		fieldValue.Set(reflect.ValueOf(v))

		return nil
	}

	if dec, ok := builtInDecoders[fieldValue.Kind()]; ok {
		v, err := dec(value)
		if err != nil {
			return fmt.Errorf("cannot convert to %q: %w", fieldValue.Kind(), err)
		}

		fieldValue.Set(reflect.ValueOf(v))
		return nil
	}
	return nil
}
