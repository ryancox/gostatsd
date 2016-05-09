package statsd

import (
	"bytes"
	"errors"
	"math"
	"strconv"

	"github.com/atlassian/gostatsd/types"
)

type lexer struct {
	input         []byte
	len           uint32
	start         uint32
	pos           uint32
	eventTitleLen uint32
	eventTextLen  uint32
	m             *types.Metric
	e             *types.Event
	tags          types.Tags
	namespace     string
	err           error
	sampling      float64
}

// assumes we don't have \x00 bytes in input.
const eof byte = 0

var (
	errMissingKeySep         = errors.New("missing key separator")
	errEmptyKey              = errors.New("key zero len")
	errMissingValueSep       = errors.New("missing value separator")
	errInvalidType           = errors.New("invalid type")
	errInvalidFormat         = errors.New("invalid format")
	errInvalidSamplingOrTags = errors.New("invalid sampling or tags")
	errInvalidAttributes     = errors.New("invalid event attributes")
	errOverflow              = errors.New("overflow")
	errNotEnoughData         = errors.New("not enough data")
	errNaN                   = errors.New("invalid value NaN")
)

var escapedNewline = []byte("\\n")
var newline = []byte("\n")

var priorityNormal = []byte("normal")
var priorityLow = []byte("low")

var alertInfo = []byte("info")
var alertError = []byte("error")
var alertWarning = []byte("warning")
var alertSuccess = []byte("success")

func (l *lexer) next() byte {
	if l.pos >= l.len {
		return eof
	}
	b := l.input[l.pos]
	l.pos++
	return b
}

func (l *lexer) run(input []byte, namespace string) (*types.Metric, *types.Event, error) {
	l.input = input
	l.namespace = namespace
	l.len = uint32(len(l.input))
	l.sampling = float64(1)

	for state := lexSpecial; state != nil; {
		state = state(l)
	}
	if l.err != nil {
		return nil, nil, l.err
	}
	if l.m != nil {
		if l.m.Type != types.SET {
			v, err := strconv.ParseFloat(l.m.StringValue, 64)
			if err != nil {
				return nil, nil, err
			}
			if math.IsNaN(v) {
				return nil, nil, errNaN
			}
			l.m.Value = v
			l.m.StringValue = ""
		}
		if l.m.Type == types.COUNTER {
			l.m.Value = l.m.Value / l.sampling
		}
		l.m.Tags = l.tags
	} else {
		l.e.Tags = l.tags
	}
	return l.m, l.e, nil
}

type stateFn func(*lexer) stateFn

// check the first byte for special DataDog type.
func lexSpecial(l *lexer) stateFn {
	switch b := l.next(); b {
	case '_':
		return lexDataDogSpecial
	case eof:
		l.err = errInvalidType
		return nil
	default:
		l.pos--
		l.m = new(types.Metric)
		return lexKeySep
	}
}

// lex until we find the colon separator between key and value.
func lexKeySep(l *lexer) stateFn {
	for {
		switch b := l.next(); b {
		case '/':
			l.input[l.pos-1] = '-'
		case ' ', '\t':
			l.input[l.pos-1] = '_'
		case ':':
			return lexKey
		case eof:
			l.err = errMissingKeySep
			return nil
		case '.', '-', '_':
			continue
		default:
			r := rune(b)
			if (97 <= r && 122 >= r) || (65 <= r && 90 >= r) || (48 <= r && 57 >= r) {
				continue
			}
			l.input = append(l.input[0:l.pos-1], l.input[l.pos:]...)
			l.len--
			l.pos--
		}
	}
}

// lex DataDog special type.
func lexDataDogSpecial(l *lexer) stateFn {
	switch b := l.next(); b {
	// _e{title.length,text.length}:title|text|d:date_happened|h:hostname|p:priority|t:alert_type|#tag1,tag2
	case 'e':
		l.e = new(types.Event)
		return lexAssert('{',
			lexUint32(&l.eventTitleLen,
				lexAssert(',',
					lexUint32(&l.eventTextLen,
						lexAssert('}', lexAssert(':', lexEventBody))))))
	default:
		l.err = errInvalidType
		return nil
	}
}

func lexEventBody(l *lexer) stateFn {
	if l.len-l.pos < l.eventTitleLen+1+l.eventTextLen {
		l.err = errNotEnoughData
		return nil
	}
	if l.input[l.pos+l.eventTitleLen] != '|' {
		l.err = errInvalidFormat
		return nil
	}
	l.e.Title = string(l.input[l.pos : l.pos+l.eventTitleLen])
	l.pos += l.eventTitleLen + 1
	l.e.Text = string(bytes.Replace(l.input[l.pos:l.pos+l.eventTextLen], escapedNewline, newline, -1))
	l.pos += l.eventTextLen
	return lexEventAttributes
}

func lexEventAttributes(l *lexer) stateFn {
	switch b := l.next(); b {
	case '|':
		return lexEventAttribute
	case eof:
	default:
		l.err = errInvalidAttributes
	}
	return nil
}

func lexEventAttribute(l *lexer) stateFn {
	// d:date_happened|h:hostname|p:priority|t:alert_type|#tag1,tag2
	switch b := l.next(); b {
	case 'd':
		return lexAssert(':', lexUint(func(l *lexer, value uint64) stateFn {
			if value > math.MaxInt64 {
				l.err = errOverflow
				return nil
			}
			l.e.DateHappened = int64(value)
			return lexEventAttributes
		}))
	case 'h':
		return lexAssert(':', lexUntil('|', func(l *lexer, data []byte) stateFn {
			l.e.Hostname = string(data)
			return lexEventAttributes
		}))
	case 'p':
		return lexAssert(':', lexUntil('|', func(l *lexer, data []byte) stateFn {
			if bytes.Equal(data, priorityLow) {
				l.e.Priority = types.PriLow
			} else if bytes.Equal(data, priorityNormal) {
				// Normal is default
			} else {
				l.err = errInvalidAttributes
				return nil
			}
			return lexEventAttributes
		}))
	case 't':
		return lexAssert(':', lexUntil('|', func(l *lexer, data []byte) stateFn {
			if bytes.Equal(data, alertError) {
				l.e.AlertType = types.AlertError
			} else if bytes.Equal(data, alertWarning) {
				l.e.AlertType = types.AlertWarning
			} else if bytes.Equal(data, alertSuccess) {
				l.e.AlertType = types.AlertSuccess
			} else if bytes.Equal(data, alertInfo) {
				// Info is default
			} else {
				l.err = errInvalidAttributes
				return nil
			}
			return lexEventAttributes
		}))
	case '#':
		return lexTags
	case eof:
	default:
		l.err = errInvalidAttributes
	}
	return nil
}

func lexUint32(target *uint32, next stateFn) stateFn {
	return lexUint(func(l *lexer, value uint64) stateFn {
		if value > math.MaxUint32 {
			l.err = errOverflow
			return nil
		}
		*target = uint32(value)
		return next
	})
}

func lexUint(handler func(*lexer, uint64) stateFn) stateFn {
	return func(l *lexer) stateFn {
		var value uint64
		start := l.pos
	loop:
		for {
			switch b := l.next(); {
			case '0' <= b && b <= '9':
				n := value*10 + uint64(b-'0')
				if n < value {
					l.err = errOverflow
					return nil
				}
				value = n
			case b == eof:
				break loop
			default:
				l.pos--
				break loop
			}
		}
		if start == l.pos {
			l.err = errInvalidFormat
			return nil
		}
		return handler(l, value)
	}
}

// lexAssert returns a function that checks if the next byte matches the provided byte and returns next in that case.
func lexAssert(nextByte byte, next stateFn) stateFn {
	return func(l *lexer) stateFn {
		switch b := l.next(); b {
		case nextByte:
			return next
		default:
			l.err = errInvalidFormat
			return nil
		}
	}
}

func lexUntil(stop byte, handler func(*lexer, []byte) stateFn) stateFn {
	return func(l *lexer) stateFn {
		start := l.pos
		p := bytes.IndexByte(l.input[l.pos:], stop)
		switch p {
		case -1:
			l.pos = l.len
		default:
			l.pos += uint32(p)
		}
		return handler(l, l.input[start:l.pos])
	}
}

// lex the key.
func lexKey(l *lexer) stateFn {
	if l.start == l.pos-1 {
		l.err = errEmptyKey
		return nil
	}
	l.m.Name = string(l.input[l.start : l.pos-1])
	if l.namespace != "" {
		l.m.Name = l.namespace + "." + l.m.Name
	}
	l.start = l.pos
	return lexValueSep
}

// lex until we find the pipe separator between value and modifier.
func lexValueSep(l *lexer) stateFn {
	for {
		// cheap check here. ParseFloat will do it.
		switch b := l.next(); b {
		case '|':
			return lexValue
		case eof:
			l.err = errMissingValueSep
			return nil
		}
	}
}

// lex the value.
func lexValue(l *lexer) stateFn {
	l.m.StringValue = string(l.input[l.start : l.pos-1])
	l.start = l.pos
	return lexType
}

// lex the type.
func lexType(l *lexer) stateFn {
	b := l.next()
	switch b {
	case 'c':
		l.m.Type = types.COUNTER
		l.start = l.pos
		return lexTypeSep
	case 'g':
		l.m.Type = types.GAUGE
		l.start = l.pos
		return lexTypeSep
	case 'm':
		if b := l.next(); b != 's' {
			l.err = errInvalidType
			return nil
		}
		l.start = l.pos
		l.m.Type = types.TIMER
		return lexTypeSep
	case 's':
		l.m.Type = types.SET
		l.start = l.pos
		return lexTypeSep
	default:
		l.err = errInvalidType
		return nil

	}
}

// lex the possible separator between type and sampling rate.
func lexTypeSep(l *lexer) stateFn {
	b := l.next()
	switch b {
	case eof:
		return nil
	case '|':
		l.start = l.pos
		return lexSampleRateOrTags
	}
	l.err = errInvalidType
	return nil
}

// lex the sample rate or the tags.
func lexSampleRateOrTags(l *lexer) stateFn {
	b := l.next()
	switch b {
	case '@':
		l.start = l.pos
		for {
			switch b := l.next(); b {
			case '|':
				return lexSampleRate
			case eof:
				l.pos++
				return lexSampleRate
			}
		}
	case '#':
		return lexTags
	default:
		l.err = errInvalidSamplingOrTags
		return nil
	}
}

// lex the sample rate.
func lexSampleRate(l *lexer) stateFn {
	v, err := strconv.ParseFloat(string(l.input[l.start:l.pos-1]), 64)
	if err != nil {
		l.err = err
		return nil
	}
	l.sampling = v
	if l.pos >= l.len {
		return nil
	}
	return lexAssert('#', lexTags)
}

// lex the tags.
func lexTags(l *lexer) stateFn {
	l.start = l.pos
	for {
		switch b := l.next(); b {
		case ',':
			l.tags = append(l.tags, string(l.input[l.start:l.pos-1]))
			l.start = l.pos
		case eof:
			l.pos++
			l.tags = append(l.tags, string(l.input[l.start:l.pos-1]))
			return nil
		case '.', ':', '-', '_':
			continue
		case '/':
			l.input[l.pos-1] = '-'
		case ' ', '\t':
			l.input[l.pos-1] = '_'
		default:
			r := rune(b)
			if (97 <= r && 122 >= r) || (48 <= r && 57 >= r) {
				continue
			}
			if 65 <= r && 90 >= r {
				l.input[l.pos-1] = byte(r + 32)
				continue
			}
			l.input = append(l.input[0:l.pos-1], l.input[l.pos:]...)
			l.len--
			l.pos--
		}
	}
}
