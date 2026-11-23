package pobwindow

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

const maxMessage = 8 << 20

type connection struct {
	Port       uint16 `json:"port"`
	Token      string `json:"token"`
	InstallDir string `json:"installDir"`
}

type BridgeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BridgeError) Error() string { return e.Code + ": " + e.Message }

type Client struct {
	conn      net.Conn
	reader    *bufio.Scanner
	directory string
	token     string
	sequence  int
	revision  int
	status    json.RawMessage
}

func StateDirectory() (string, error) {
	root := os.Getenv("XDG_DATA_HOME")
	if runtime.GOOS == "windows" {
		root = os.Getenv("LOCALAPPDATA")
	}
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		root = filepath.Join(home, ".local", "share")
	}
	return filepath.Abs(filepath.Join(root, "pob-poe2-mcp", "bridge"))
}

func Dial(ctx context.Context, state, directory string) (*Client, error) {
	endpoint, err := readEndpoint(state, directory)
	if err != nil {
		return nil, err
	}
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp4", "127.0.0.1:"+strconv.Itoa(int(endpoint.Port)))
	if err != nil {
		return nil, &BridgeError{"BRIDGE_NOT_RUNNING", "Не удалось подключиться к окну PoB"}
	}
	reader := bufio.NewScanner(conn)
	reader.Buffer(nil, maxMessage)
	client := &Client{conn: conn, reader: reader, directory: endpoint.InstallDir, token: endpoint.Token}
	status, err := client.Call(ctx, "initialize", nil)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	client.status = status
	return client, nil
}

func readEndpoint(state, directory string) (connection, error) {
	var endpoint, config connection
	if err := readConnection(filepath.Join(state, "endpoint.json"), &endpoint); err != nil {
		return endpoint, &BridgeError{"BRIDGE_NOT_RUNNING", "Окно PoB не опубликовало адрес моста"}
	}
	if endpoint.Port < 1024 || len(endpoint.Token) < 32 || !filepath.IsAbs(endpoint.InstallDir) {
		return endpoint, &BridgeError{"INVALID_BRIDGE_ENDPOINT", "Некорректный адрес моста PoB"}
	}
	if err := readConnection(filepath.Join(state, "config.json"), &config); err != nil || config.Token != endpoint.Token {
		return endpoint, &BridgeError{"BRIDGE_RESTART_REQUIRED", "Перезапустите PoB, чтобы окно применило настройки моста"}
	}
	if directory != "" && !sameDirectory(directory, endpoint.InstallDir) {
		return endpoint, &BridgeError{"APP_PATH_MISMATCH", "Мост запущен из другого каталога PoB"}
	}
	return endpoint, nil
}

func readConnection(path string, value *connection) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	content, err := io.ReadAll(io.LimitReader(file, 64<<10))
	if err != nil {
		return err
	}
	return json.Unmarshal(content, value)
}

func sameDirectory(left, right string) bool {
	a, err := os.Stat(left)
	if err != nil {
		return false
	}
	b, err := os.Stat(right)
	return err == nil && a.IsDir() && b.IsDir() && os.SameFile(a, b)
}

func (c *Client) Close() error            { return c.conn.Close() }
func (c *Client) Status() json.RawMessage { return c.status }

type request struct {
	Protocol int            `json:"protocol"`
	ID       string         `json:"id"`
	Method   string         `json:"method"`
	Params   map[string]any `json:"params"`
	Token    string         `json:"token"`
}

type response struct {
	Protocol int             `json:"protocol"`
	ID       string          `json:"id"`
	Revision int             `json:"revision"`
	Result   json.RawMessage `json:"result"`
	Error    *BridgeError    `json:"error"`
}

func (c *Client) Call(ctx context.Context, method string, arguments map[string]any) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	deadline, _ := ctx.Deadline()
	if err := c.conn.SetDeadline(deadline); err != nil {
		return nil, err
	}
	stop := context.AfterFunc(ctx, func() { _ = c.conn.Close() })
	defer stop()
	reply, err := c.exchange(method, arguments)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, err
	}
	c.revision = reply.Revision
	if reply.Error != nil {
		return nil, reply.Error
	}
	return reply.Result, nil
}

func (c *Client) exchange(method string, arguments map[string]any) (response, error) {
	c.sequence++
	params := make(map[string]any, len(arguments)+1)
	for key, value := range arguments {
		params[key] = value
	}
	params["revision"] = c.revision
	value := request{Protocol: 1, ID: strconv.Itoa(c.sequence), Method: method, Params: params, Token: c.token}
	if err := c.send(value); err != nil {
		return response{}, err
	}
	return c.receive(value.ID)
}

func (c *Client) send(value request) error {
	line, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if len(line) >= maxMessage {
		return errors.New("запрос к мосту PoB превышает допустимый размер")
	}
	_, err = c.conn.Write(append(line, '\n'))
	return err
}

func (c *Client) receive(id string) (response, error) {
	var reply response
	if !c.reader.Scan() {
		return reply, &BridgeError{"ENGINE_STOPPED", "Окно PoB завершило обмен с мостом"}
	}
	if err := json.Unmarshal(c.reader.Bytes(), &reply); err != nil {
		return reply, fmt.Errorf("неверный ответ моста PoB: %w", err)
	}
	if reply.Protocol != 1 || reply.ID != id || reply.Revision < 0 || reply.Result == nil && reply.Error == nil {
		return reply, &BridgeError{"INVALID_RESPONSE", "Нарушена последовательность ответов PoB"}
	}
	return reply, nil
}
