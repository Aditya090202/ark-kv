// Package server implements the TCP server and RESP command handlers.
package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Aditya090202/distributed-key-value-store/internal/resp"
	"github.com/Aditya090202/distributed-key-value-store/internal/store"
)

// Server is the miniKV TCP server.
type Server struct {
	store   *store.Store
	clients sync.Map // map[net.Conn]struct{}
	conns   int64    // total connections handled
	cmds    int64    // total commands processed
}

func New() *Server {
	return &Server{
		store: store.New(),
	}
}

// HandleConnection processes a single client connection.
func (s *Server) HandleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.clients.Delete(conn)
	}()

	atomic.AddInt64(&s.conns, 1)
	s.clients.Store(conn, struct{}{})

	parser := resp.NewParser(conn)
	remote := conn.RemoteAddr().String()

	log.Printf("[%s] connected", remote)

	for {
		// Read the next RESP command
		v, err := parser.Read()
		if err != nil {
			// Connection closed or error
			break
		}

		atomic.AddInt64(&s.cmds, 1)

		// Expect an array of bulk strings
		if v.Type != resp.TypeArray {
			resp.WriteError(conn, "ERR expected array")
			continue
		}

		if len(v.Array) == 0 {
			resp.WriteError(conn, "ERR empty command")
			continue
		}

		// First element is the command name
		cmd := strings.ToUpper(v.Array[0].String)
		args := v.Array[1:]

		s.handleCommand(conn, cmd, args)
	}

	log.Printf("[%s] disconnected", remote)
}

func (s *Server) handleCommand(conn net.Conn, cmd string, args []resp.Value) {
	switch cmd {
	case "PING":
		if len(args) > 0 {
			resp.Write(conn, resp.BulkString(args[0].String))
		} else {
			resp.Write(conn, resp.SimpleString("PONG"))
		}

	case "SET":
		if len(args) < 2 {
			resp.WriteError(conn, "ERR wrong number of arguments for 'SET' command")
			return
		}
		key := args[0].String
		val := args[1].String

		// Parse optional EX/PX/EXAT/PXAT
		var ttl time.Duration
		for i := 2; i < len(args); i++ {
			opt := strings.ToUpper(args[i].String)
			switch opt {
			case "EX":
				if i+1 >= len(args) {
					resp.WriteError(conn, "ERR syntax error")
					return
				}
				sec, err := strconv.ParseInt(args[i+1].String, 10, 64)
				if err != nil {
					resp.WriteError(conn, "ERR invalid EX value")
					return
				}
				ttl = time.Duration(sec) * time.Second
				i++
			case "PX":
				if i+1 >= len(args) {
					resp.WriteError(conn, "ERR syntax error")
					return
				}
				ms, err := strconv.ParseInt(args[i+1].String, 10, 64)
				if err != nil {
					resp.WriteError(conn, "ERR invalid PX value")
					return
				}
				ttl = time.Duration(ms) * time.Millisecond
				i++
			}
		}

		s.store.Set(key, val, ttl)
		resp.WriteOK(conn)

	case "GET":
		if len(args) != 1 {
			resp.WriteError(conn, "ERR wrong number of arguments for 'GET' command")
			return
		}
		val, ok := s.store.Get(args[0].String)
		if !ok {
			resp.Write(conn, resp.NullBulk())
		} else {
			resp.Write(conn, resp.BulkString(val))
		}

	case "DEL":
		if len(args) == 0 {
			resp.WriteError(conn, "ERR wrong number of arguments for 'DEL' command")
			return
		}
		keys := make([]string, len(args))
		for i, a := range args {
			keys[i] = a.String
		}
		n := s.store.Del(keys...)
		resp.Write(conn, resp.Integer(int64(n)))

	case "EXISTS":
		if len(args) != 1 {
			resp.WriteError(conn, "ERR wrong number of arguments for 'EXISTS' command")
			return
		}
		n := s.store.Exists(args[0].String)
		resp.Write(conn, resp.Integer(int64(n)))

	case "EXPIRE":
		if len(args) != 2 {
			resp.WriteError(conn, "ERR wrong number of arguments for 'EXPIRE' command")
			return
		}
		sec, err := strconv.ParseInt(args[1].String, 10, 64)
		if err != nil {
			resp.WriteError(conn, "ERR invalid expire time")
			return
		}
		ok := s.store.Expire(args[0].String, time.Duration(sec)*time.Second)
		if ok {
			resp.Write(conn, resp.Integer(1))
		} else {
			resp.Write(conn, resp.Integer(0))
		}

	case "TTL":
		if len(args) != 1 {
			resp.WriteError(conn, "ERR wrong number of arguments for 'TTL' command")
			return
		}
		ttl := s.store.TTL(args[0].String)
		resp.Write(conn, resp.Integer(ttl))

	case "KEYS":
		pattern := "*"
		if len(args) > 0 {
			pattern = args[0].String
		}
		keys := s.store.Keys(pattern)
		items := make([]resp.Value, len(keys))
		for i, k := range keys {
			items[i] = resp.BulkString(k)
		}
		resp.Write(conn, resp.Array(items))

	case "DBSIZE":
		resp.Write(conn, resp.Integer(int64(s.store.Len())))

	case "FLUSHDB":
		s.store.Flush()
		resp.WriteOK(conn)

	case "INFO":
		stats := s.store.Stats()
		lines := []string{
			fmt.Sprintf("# Server"),
			fmt.Sprintf("conns:%d", atomic.LoadInt64(&s.conns)),
			fmt.Sprintf("cmds:%d", atomic.LoadInt64(&s.cmds)),
			fmt.Sprintf("keys:%d", stats["keys"]),
			fmt.Sprintf("expired_keys:%d", stats["expired"]),
		}
		info := strings.Join(lines, "\r\n")
		resp.Write(conn, resp.BulkString(info+"\r\n"))

	case "COMMAND":
		// Redis sends a big array for COMMAND. We send a minimal placeholder.
		resp.Write(conn, resp.Array([]resp.Value{}))

	default:
		resp.WriteError(conn, fmt.Sprintf("ERR unknown command '%s'", cmd))
	}
}

// Shutdown cleanly stops the server.
func (s *Server) Shutdown() {
	s.store.Shutdown()
	s.clients.Range(func(key, _ interface{}) bool {
		key.(net.Conn).Close()
		return true
	})
}