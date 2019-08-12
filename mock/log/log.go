package log

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
	"unicode/utf8"
)

const (
	errorKey          = "LOG_NEW_ERROR"
	timeFormat        = "2006-01-02T15:04:05-0700"
	floatFormat       = 'f'
	termCtxMaxPadding = 40
	calldepth         = 3
)

type Logger interface {
	Test() Logger
	New(kv ...interface{}) Logger

	Printf(format string, params ...interface{})
	Println(format string, params ...interface{})
	Debug(msg string, ctx ...interface{})
	Info(msg string, ctx ...interface{})
	Error(msg string, ctx ...interface{})
}

type LoggerImpl struct {
	ctx []interface{}
	log *log.Logger
}

var (
	root             *LoggerImpl
	fieldPaddingLock sync.RWMutex
	fieldPadding     = make(map[string]int)
)

func init() {
	root = &LoggerImpl{
		ctx: []interface{}{},
		log: log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *LoggerImpl) Test() Logger {
	return &LoggerImpl{
		ctx: []interface{}{},
		log: log.New(os.Stdout, "TEST*", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func (l *LoggerImpl) New(kv ...interface{}) Logger {
	return &LoggerImpl{
		ctx: newContext(l.ctx, kv),
		log: l.log,
	}
}

func (l *LoggerImpl) Printf(format string, params ...interface{}) {
	l.log.Output(calldepth, fmt.Sprintf(format, params...))
}

func (l *LoggerImpl) Println(format string, params ...interface{}) {
	l.log.Output(calldepth, fmt.Sprintf(format, params...))
}

func (l *LoggerImpl) Debug(msg string, ctx ...interface{}) {
	l.log.Output(calldepth, logfmt("Debug", msg, newContext(l.ctx, ctx), 0, false))
}

func (l *LoggerImpl) Info(msg string, ctx ...interface{}) {
	l.log.Output(calldepth, logfmt("Info", msg, newContext(l.ctx, ctx), 0, false))
}

func (l *LoggerImpl) Error(msg string, ctx ...interface{}) {
	l.log.Output(calldepth, logfmt("Error", msg, newContext(l.ctx, ctx), 0, false))
}

func Test() Logger {
	return &LoggerImpl{
		ctx: []interface{}{},
		log: log.New(os.Stdout, "TEST*", log.Ldate|log.Ltime|log.Lshortfile),
	}
}

func New(kv ...interface{}) Logger {
	return &LoggerImpl{
		ctx: newContext(root.ctx, kv),
		log: root.log,
	}
}

func Debug(msg string, ctx ...interface{}) {
	root.log.Output(calldepth, logfmt("Debug", msg, newContext(root.ctx, ctx), 0, false))
}

func Info(msg string, ctx ...interface{}) {
	root.log.Output(calldepth, logfmt("Info", msg, newContext(root.ctx, ctx), 0, false))
}

func Error(msg string, ctx ...interface{}) {
	root.log.Output(calldepth, logfmt("Error", msg, newContext(root.ctx, ctx), 0, false))
}

//-----------------------------------------------------------------------------

func newContext(prefix []interface{}, suffix []interface{}) []interface{} {
	normalizedSuffix := normalize(suffix)
	newCtx := make([]interface{}, len(prefix)+len(normalizedSuffix))
	n := copy(newCtx, prefix)
	copy(newCtx[n:], normalizedSuffix)
	return newCtx
}

func normalize(ctx []interface{}) []interface{} {
	// if the caller passed a Ctx object, then expand it
	if len(ctx) == 1 {
		if ctxMap, ok := ctx[0].(Ctx); ok {
			ctx = ctxMap.toArray()
		}
	}

	// ctx needs to be even because it's a series of key/value pairs
	// no one wants to check for errors on logging functions,
	// so instead of erroring on bad input, we'll just make sure
	// that things are the right length and users can fix bugs
	// when they see the output looks wrong
	if len(ctx)%2 != 0 {
		ctx = append(ctx, nil, errorKey, "Normalized odd number of arguments by adding nil")
	}

	return ctx
}

// Ctx is a map of key/value pairs to pass as context to a log function
// Use this only if you really need greater safety around the arguments you pass
// to the logging functions.
type Ctx map[string]interface{}

func (c Ctx) toArray() []interface{} {
	arr := make([]interface{}, len(c)*2)

	i := 0
	for k, v := range c {
		arr[i] = k
		arr[i+1] = v
		i += 2
	}

	return arr
}

func logfmt(lvl, msg string, ctx []interface{}, color int, term bool) string {
	buf := &bytes.Buffer{}
	buf.WriteString(lvl)
	buf.WriteString(": msg=\"")
	buf.WriteString(msg)
	for i := 0; i < len(ctx); i += 2 {
		if i != 0 {
			buf.WriteByte(' ')
		} else {
			buf.WriteString("\" ")
		}

		k, ok := ctx[i].(string)
		v := formatLogfmtValue(ctx[i+1], term)
		if !ok {
			k, v = errorKey, formatLogfmtValue(k, term)
		}

		// XXX: we should probably check that all of your key bytes aren't invalid
		fieldPaddingLock.RLock()
		padding := fieldPadding[k]
		fieldPaddingLock.RUnlock()

		length := utf8.RuneCountInString(v)
		if padding < length && length <= termCtxMaxPadding {
			padding = length

			fieldPaddingLock.Lock()
			fieldPadding[k] = padding
			fieldPaddingLock.Unlock()
		}
		if color > 0 {
			fmt.Fprintf(buf, "\x1b[%dm%s\x1b[0m=", color, k)
		} else {
			buf.WriteString(k)
			buf.WriteByte('=')
		}
		buf.WriteString(v)
		if i < len(ctx)-2 && padding > length {
			buf.Write(bytes.Repeat([]byte{' '}, padding-length))
		}
	}
	return buf.String()
}

// TerminalStringer is an analogous interface to the stdlib stringer, allowing
// own types to have custom shortened serialization formats when printed to the
// screen.
type TerminalStringer interface {
	TerminalString() string
}

// formatValue formats a value for serialization
func formatLogfmtValue(value interface{}, term bool) string {
	if value == nil {
		return "nil"
	}

	if t, ok := value.(time.Time); ok {
		// Performance optimization: No need for escaping since the provided
		// timeFormat doesn't have any escape characters, and escaping is
		// expensive.
		return t.Format(timeFormat)
	}
	if term {
		if s, ok := value.(TerminalStringer); ok {
			// Custom terminal stringer provided, use that
			return escapeString(s.TerminalString())
		}
	}
	value = formatShared(value)
	switch v := value.(type) {
	case bool:
		return strconv.FormatBool(v)
	case float32:
		return strconv.FormatFloat(float64(v), floatFormat, 3, 64)
	case float64:
		return strconv.FormatFloat(v, floatFormat, 3, 64)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", value)
	case string:
		return escapeString(v)
	default:
		return escapeString(fmt.Sprintf("%+v", value))
	}
}

var stringBufPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func escapeString(s string) string {
	needsQuotes := false
	needsEscape := false
	for _, r := range s {
		if r <= ' ' || r == '=' || r == '"' {
			needsQuotes = true
		}
		if r == '\\' || r == '"' || r == '\n' || r == '\r' || r == '\t' {
			needsEscape = true
		}
	}
	if !needsEscape && !needsQuotes {
		return s
	}
	e := stringBufPool.Get().(*bytes.Buffer)
	e.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\\', '"':
			e.WriteByte('\\')
			e.WriteByte(byte(r))
		case '\n':
			e.WriteString("\\n")
		case '\r':
			e.WriteString("\\r")
		case '\t':
			e.WriteString("\\t")
		default:
			e.WriteRune(r)
		}
	}
	e.WriteByte('"')
	var ret string
	if needsQuotes {
		ret = e.String()
	} else {
		ret = string(e.Bytes()[1 : e.Len()-1])
	}
	e.Reset()
	stringBufPool.Put(e)
	return ret
}

func formatShared(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case time.Time:
		return v.Format(timeFormat)

	case error:
		return v.Error()

	case fmt.Stringer:
		return v.String()

	default:
		return v
	}
}
