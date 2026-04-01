package axy

import (
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Resident struct {
	parent          *Base
	local           Local
	key             string
	pascalCasedKind string
	snakeCasedKind  string
	Reference
}

func (r *Resident) resident() *Resident {
	return r
}

func (r *Resident) Key() string {
	return r.key
}

func (r *Resident) PascalCasedKind() string {
	if r.pascalCasedKind != "" {
		return r.pascalCasedKind
	}

	if r.local == nil {
		r.pascalCasedKind = "Unknown"
		return r.pascalCasedKind
	}

	t := reflect.TypeOf(r.local)

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	name := t.Name()
	firstLetter, size := utf8.DecodeRuneInString(name)
	r.pascalCasedKind = string(unicode.ToUpper(firstLetter)) + name[size:]
	return r.pascalCasedKind
}

func (r *Resident) SnakeCasedKind() string {
	if r.snakeCasedKind != "" {
		return r.snakeCasedKind
	}

	snakeCasedKind := matchFirstUpper.ReplaceAllString(r.pascalCasedKind, "${1}_${2}")
	snakeCasedKind = matchAllUpper.ReplaceAllString(snakeCasedKind, "${1}_${2}")
	r.snakeCasedKind = strings.ToLower(snakeCasedKind)
	return r.snakeCasedKind
}

func (r *Resident) prepareLoggerArguments(args ...any) []any {
	combinedArgsLength := len(args)
	if r.key != "" {
		combinedArgsLength++
	}
	combinedArgs := make([]any, 0, combinedArgsLength)

	if r.key != "" {
		combinedArgs = append(combinedArgs, r.SnakeCasedKind()+"_key", r.key)
	}

	combinedArgs = append(combinedArgs, args...)
	return combinedArgs
}

func (r *Resident) Debug(message string, args ...any) {
	r.parent.Debug(r.PascalCasedKind()+": "+message, r.prepareLoggerArguments(args...)...)
}

func (r *Resident) Info(message string, args ...any) {
	r.parent.Info(r.PascalCasedKind()+": "+message, r.prepareLoggerArguments(args...)...)
}

func (r *Resident) Warn(message string, args ...any) {
	r.parent.Warn(r.PascalCasedKind()+": "+message, r.prepareLoggerArguments(args...)...)
}

func (r *Resident) Error(message string, args ...any) {
	r.parent.Error(r.PascalCasedKind()+": "+message, r.prepareLoggerArguments(args...)...)
}
