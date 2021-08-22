package echo_binder

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"reflect"
	"strings"
)

// Handler is ...
type Handler func(interface{}) (interface{}, error)

// Values is map that contains Handler for each value from tag
type Values map[string]Handler

// Binder is main structure that implement echo.Binder interface
type Binder struct {
	tags map[string]tag
}

// tag is subsidiary structure for keeping information about
type tag struct {
	values Values
	sep    string
}

// Set method note new tag with handlers
func (b *Binder) Set(tagName, sep string, values Values) {
	if b.tags == nil {
		b.tags = make(map[string]tag)
	}

	b.tags[tagName] = tag{
		values: values,
		sep:    sep,
	}
}

func (b *Binder) Bind(i interface{}, c echo.Context) error {
	db := new(echo.DefaultBinder)
	if err := db.Bind(i, c); err != nil {
		return err
	}

	if err := b.work(i); err != nil {
		return err
	}

	return nil
}

func (b *Binder) work(i interface{}) error {
	r := reflect.ValueOf(i).Elem()

	for j := 0; j < r.NumField(); j++ {
		field := r.Field(j)

		if !field.CanSet() {
			continue
		}

		if field.Kind() == reflect.Struct {
			if err := b.work(field.Addr().Interface()); err != nil {
				return err
			}
			continue
		}

		tags := r.Type().Field(j).Tag

		for tagKey, tag := range b.tags {
			sTag, isTag := tags.Lookup(tagKey)

			if !isTag {
				continue
			}

			values := strings.Split(sTag, tag.sep)

			for _, value := range values {
				handler, isValue := tag.values[strings.TrimSpace(value)]

				if !isValue {
					return fmt.Errorf("binder_error: unknown '%s' tag_value from '%s' tag_key", value, tagKey)
				}

				newI, err := handler(field.Interface())

				if err != nil {
					return err
				}

				field.Set(reflect.ValueOf(newI))
			}
		}
	}

	return nil
}
