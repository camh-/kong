package kong

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

// A Resolver resolves a Flag value from an external source.
type Resolver interface {
	// Validate configuration against Application.
	//
	// This can be used to validate that all provided configuration is valid within  this application.
	Validate(app *Application) error

	// Resolve the value for a Flag.
	Resolve(context *Context, parent *Path, flag *Flag) (interface{}, error)
}

// ResolverFunc is a convenience type for non-validating Resolvers.
type ResolverFunc func(context *Context, parent *Path, flag *Flag) (interface{}, error)

var _ Resolver = ResolverFunc(nil)

func (r ResolverFunc) Resolve(context *Context, parent *Path, flag *Flag) (interface{}, error) { // nolint: golint
	return r(context, parent, flag)
}
func (r ResolverFunc) Validate(app *Application) error { return nil } //  nolint: golint

// DefaultsResolver resolves values from the `default` tag on a flag.
//
// It is installed by default. Use ClearResolvers() to disable this.
func DefaultsResolver() Resolver {
	return ResolverFunc(func(context *Context, parent *Path, flag *Flag) (interface{}, error) {
		if flag.Tag.Default == "" {
			return nil, nil
		}
		return flag.Tag.Default, nil
	})
}

// EnvarResolver resolves values from environment variables.
//
// It is installed by default. Use ClearResolvers() to disable this.
func EnvarResolver() Resolver {
	return ResolverFunc(func(context *Context, parent *Path, flag *Flag) (interface{}, error) {
		if flag.Tag.Env == "" {
			return nil, nil
		}
		envar := os.Getenv(flag.Tag.Env)
		if envar != "" {
			return envar, nil
		}
		return nil, nil
	})
}

// JSON returns a Resolver that retrieves values from a JSON source.
//
// Hyphens in flag names are replaced with underscores.
func JSON(r io.Reader) (Resolver, error) {
	values := map[string]interface{}{}
	err := json.NewDecoder(r).Decode(&values)
	if err != nil {
		return nil, err
	}
	var f ResolverFunc = func(context *Context, parent *Path, flag *Flag) (interface{}, error) {
		name := strings.Replace(flag.Name, "-", "_", -1)
		raw, ok := values[name]
		if !ok {
			return nil, nil
		}
		return raw, nil
	}

	return f, nil
}
