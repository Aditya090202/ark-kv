package resp

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseSimpleString(t *testing.T) {
	r := strings.NewReader("+OK\r\n")
	p := NewParser(r)
	v, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeSimpleString || v.String != "OK" {
		t.Fatalf("expected +OK, got %+v", v)
	}
}

func TestParseError(t *testing.T) {
	r := strings.NewReader("-ERR unknown command\r\n")
	p := NewParser(r)
	v, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeError || v.String != "ERR unknown command" {
		t.Fatalf("expected error, got %+v", v)
	}
}

func TestParseInteger(t *testing.T) {
	r := strings.NewReader(":42\r\n")
	p := NewParser(r)
	v, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeInteger || v.Int != 42 {
		t.Fatalf("expected :42, got %+v", v)
	}
}

func TestParseBulkString(t *testing.T) {
	r := strings.NewReader("$5\r\nhello\r\n")
	p := NewParser(r)
	v, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeBulkString || v.String != "hello" {
		t.Fatalf("expected $5\\r\\nhello, got %+v", v)
	}
}

func TestParseNullBulkString(t *testing.T) {
	r := strings.NewReader("$-1\r\n")
	p := NewParser(r)
	v, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeBulkString || !v.Null {
		t.Fatal("expected null bulk string")
	}
}

func TestParseArray(t *testing.T) {
	r := strings.NewReader("*2\r\n$3\r\nGET\r\n$4\r\ntest\r\n")
	p := NewParser(r)
	v, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}
	if v.Type != TypeArray || len(v.Array) != 2 {
		t.Fatalf("expected array of 2, got %+v", v)
	}
	if v.Array[0].String != "GET" || v.Array[1].String != "test" {
		t.Fatalf("expected [GET, test], got %+v", v)
	}
}

func TestWriteSimpleString(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, SimpleString("OK"))
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "+OK\r\n" {
		t.Fatalf("expected '+OK\\r\\n', got '%s'", buf.String())
	}
}

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Error("ERR bad"))
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "-ERR bad\r\n" {
		t.Fatalf("expected '-ERR bad\\r\\n', got '%s'", buf.String())
	}
}

func TestWriteInteger(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, Integer(42))
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != ":42\r\n" {
		t.Fatalf("expected ':42\\r\\n', got '%s'", buf.String())
	}
}

func TestWriteBulkString(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, BulkString("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "$5\r\nhello\r\n" {
		t.Fatalf("expected '$5\\r\\nhello\\r\\n', got '%s'", buf.String())
	}
}

func TestWriteNullBulk(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, NullBulk())
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "$-1\r\n" {
		t.Fatalf("expected '$-1\\r\\n', got '%s'", buf.String())
	}
}

func TestWriteArray(t *testing.T) {
	items := []Value{
		BulkString("SET"),
		BulkString("key"),
		BulkString("value"),
	}
	var buf bytes.Buffer
	err := Write(&buf, Array(items))
	if err != nil {
		t.Fatal(err)
	}
	expected := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	if buf.String() != expected {
		t.Fatalf("expected %q, got %q", expected, buf.String())
	}
}

func TestRoundTrip(t *testing.T) {
	// SET key value
	input := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"
	r := strings.NewReader(input)
	p := NewParser(r)
	v, err := p.Read()
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	err = Write(&buf, v)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != input {
		t.Fatalf("round-trip failed: expected %q, got %q", input, buf.String())
	}
}