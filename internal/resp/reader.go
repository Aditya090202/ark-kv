// Package resp implements the Redis Serialization Protocol (RESP).
// Supports: Simple String, Error, Integer, Bulk String, Array, Null.
package resp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// --- Data types ---

const (
	TypeSimpleString = '+'
	TypeError        = '-'
	TypeInteger      = ':'
	TypeBulkString   = '$'
	TypeArray        = '*'
	TypeNull         = '_' // RESP3 null, also accepted as $-1 / *-1
)

// Value represents a parsed RESP value.
type Value struct {
	Type   byte
	String string // for SimpleString, Error, BulkString
	Int    int64  // for Integer
	Array  []Value
	Null   bool
}

// Ok returns a simple OK response.
func Ok() Value {
	return Value{Type: TypeSimpleString, String: "OK"}
}

// NullBulk returns a null bulk string.
func NullBulk() Value {
	return Value{Type: TypeBulkString, Null: true}
}

// Error returns an error value.
func Error(msg string) Value {
	return Value{Type: TypeError, String: msg}
}

// Integer returns an integer value.
func Integer(n int64) Value {
	return Value{Type: TypeInteger, Int: n}
}

// BulkString returns a bulk string value.
func BulkString(s string) Value {
	return Value{Type: TypeBulkString, String: s}
}

// SimpleString returns a simple string value.
func SimpleString(s string) Value {
	return Value{Type: TypeSimpleString, String: s}
}

// Array returns an array value.
func Array(items []Value) Value {
	return Value{Type: TypeArray, Array: items}
}

// --- Parser ---

type Parser struct {
	reader *bufio.Reader
}

func NewParser(r io.Reader) *Parser {
	return &Parser{reader: bufio.NewReader(r)}
}

// Read parses one RESP value from the stream.
func (p *Parser) Read() (Value, error) {
	b, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}
	switch b {
	case TypeSimpleString:
		return p.readSimpleString()
	case TypeError:
		v, err := p.readLine()
		if err != nil {
			return Value{}, err
		}
		return Value{Type: TypeError, String: v}, nil
	case TypeInteger:
		line, err := p.readLine()
		if err != nil {
			return Value{}, err
		}
		n, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return Value{}, fmt.Errorf("invalid integer: %s", line)
		}
		return Value{Type: TypeInteger, Int: n}, nil
	case TypeBulkString:
		return p.readBulkString()
	case TypeArray:
		return p.readArray()
	case TypeNull:
		// RESP3 null — consume the trailing \r\n
		p.readLine()
		return Value{Type: TypeNull, Null: true}, nil
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %c", b)
	}
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", errors.New("invalid RESP line: missing CRLF")
	}
	return line[:len(line)-2], nil
}

func (p *Parser) readSimpleString() (Value, error) {
	s, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: TypeSimpleString, String: s}, nil
}

func (p *Parser) readBulkString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	n, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid bulk string length: %s", line)
	}
	if n == -1 {
		return Value{Type: TypeBulkString, Null: true}, nil
	}
	buf := make([]byte, n+2) // content + \r\n
	if _, err := io.ReadFull(p.reader, buf); err != nil {
		return Value{}, err
	}
	return Value{Type: TypeBulkString, String: string(buf[:n])}, nil
}

func (p *Parser) readArray() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}
	n, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid array length: %s", line)
	}
	if n == -1 {
		return Value{Type: TypeArray, Null: true}, nil
	}
	arr := make([]Value, n)
	for i := 0; i < n; i++ {
		v, err := p.Read()
		if err != nil {
			return Value{}, err
		}
		arr[i] = v
	}
	return Value{Type: TypeArray, Array: arr}, nil
}

// --- Writer ---

// Write serializes a Value to RESP format.
func Write(w io.Writer, v Value) error {
	var buf bytes.Buffer
	if err := writeValue(&buf, v); err != nil {
		return err
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func writeValue(buf *bytes.Buffer, v Value) error {
	switch v.Type {
	case TypeSimpleString:
		buf.WriteByte(TypeSimpleString)
		buf.WriteString(v.String)
		buf.WriteString("\r\n")
	case TypeError:
		buf.WriteByte(TypeError)
		buf.WriteString(v.String)
		buf.WriteString("\r\n")
	case TypeInteger:
		buf.WriteByte(TypeInteger)
		buf.WriteString(strconv.FormatInt(v.Int, 10))
		buf.WriteString("\r\n")
	case TypeBulkString:
		if v.Null {
			buf.WriteString("$-1\r\n")
		} else {
			buf.WriteByte(TypeBulkString)
			buf.WriteString(strconv.Itoa(len(v.String)))
			buf.WriteString("\r\n")
			buf.WriteString(v.String)
			buf.WriteString("\r\n")
		}
	case TypeArray:
		if v.Null {
			buf.WriteString("*-1\r\n")
		} else {
			buf.WriteByte(TypeArray)
			buf.WriteString(strconv.Itoa(len(v.Array)))
			buf.WriteString("\r\n")
			for _, item := range v.Array {
				if err := writeValue(buf, item); err != nil {
					return err
				}
			}
		}
	case TypeNull:
		buf.WriteString("_\r\n")
	default:
		return fmt.Errorf("cannot write unknown RESP type: %c", v.Type)
	}
	return nil
}

// WriteError is a convenience for writing an error response.
func WriteError(w io.Writer, msg string) error {
	return Write(w, Value{Type: TypeError, String: msg})
}

// WriteOK is a convenience for writing an OK response.
func WriteOK(w io.Writer) error {
	return Write(w, Ok())
}