package axy

import (
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Local struct {
	parent          *Base
	resident        Resident
	key             string
	pascalCasedKind string
	snakeCasedKind  string
}

func (l *Local) local() *Local {
	return l
}

func (l *Local) Key() string {
	return l.key
}

func (l *Local) PascalCasedKind() string {
	if l.pascalCasedKind != "" {
		return l.pascalCasedKind
	}

	if l.resident == nil {
		l.pascalCasedKind = "Unknown"
		return l.pascalCasedKind
	}

	t := reflect.TypeOf(l.resident)

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	name := t.Name()
	firstLetter, size := utf8.DecodeRuneInString(name)
	l.pascalCasedKind = string(unicode.ToUpper(firstLetter)) + name[size:]
	return l.pascalCasedKind
}

func (l *Local) SnakeCasedKind() string {
	if l.snakeCasedKind != "" {
		return l.snakeCasedKind
	}

	snakeCasedKind := matchFirstUpper.ReplaceAllString(l.pascalCasedKind, "${1}_${2}")
	snakeCasedKind = matchAllUpper.ReplaceAllString(snakeCasedKind, "${1}_${2}")
	l.snakeCasedKind = strings.ToLower(snakeCasedKind)
	return l.snakeCasedKind
}

func (l *Local) prepareLoggerArguments(args ...any) []any {
	combinedArgsLength := len(args)
	if l.key != "" {
		combinedArgsLength++
	}
	combinedArgs := make([]any, 0, combinedArgsLength)

	if l.key != "" {
		combinedArgs = append(combinedArgs, l.SnakeCasedKind()+"_key", l.key)
	}

	combinedArgs = append(combinedArgs, args...)
	return combinedArgs
}

func (l *Local) Debug(message string, args ...any) {
	l.parent.Debug(l.PascalCasedKind()+": "+message, l.prepareLoggerArguments(args...)...)
}

func (l *Local) Info(message string, args ...any) {
	l.parent.Info(l.PascalCasedKind()+": "+message, l.prepareLoggerArguments(args...)...)
}

func (l *Local) Warn(message string, args ...any) {
	l.parent.Warn(l.PascalCasedKind()+": "+message, l.prepareLoggerArguments(args...)...)
}

func (l *Local) Error(message string, args ...any) {
	l.parent.Error(l.PascalCasedKind()+": "+message, l.prepareLoggerArguments(args...)...)
}
