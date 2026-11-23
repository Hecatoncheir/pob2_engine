package pobwindow

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadOpenBuild(t *testing.T) {
	state := t.TempDir()
	installation := t.TempDir()
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	endpoint := connection{Port: uint16(listener.Addr().(*net.TCPAddr).Port), Token: strings.Repeat("a", 64), InstallDir: installation}
	writeConnection(t, state, "config.json", endpoint)
	writeConnection(t, state, "endpoint.json", endpoint)
	served := make(chan error, 1)
	go func() {
		served <- serveOpenBuild(listener, endpoint.Token)
	}()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	client, err := Dial(ctx, state, installation)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	raw, err := client.Call(ctx, "exportBuild", nil)
	var exported struct {
		XML string `json:"xml"`
	}
	if err != nil || json.Unmarshal(raw, &exported) != nil || exported.XML != "<PathOfBuilding2/>" {
		t.Fatalf("экспорт окна: %s, %v", raw, err)
	}
	if err := <-served; err != nil {
		t.Fatal(err)
	}
}

func TestChangedBridgeTokenRequiresRestart(t *testing.T) {
	state := t.TempDir()
	endpoint := connection{Port: 37281, Token: strings.Repeat("a", 64), InstallDir: t.TempDir()}
	writeConnection(t, state, "endpoint.json", endpoint)
	endpoint.Token = strings.Repeat("b", 64)
	writeConnection(t, state, "config.json", endpoint)

	_, err := Dial(t.Context(), state, "")
	bridgeError, ok := err.(*BridgeError)
	if !ok || bridgeError.Code != "BRIDGE_RESTART_REQUIRED" {
		t.Fatalf("изменившийся токен моста: %v", err)
	}
}

func writeConnection(t *testing.T, directory, name string, value connection) {
	t.Helper()
	content, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, name), content, 0600); err != nil {
		t.Fatal(err)
	}
}

func serveOpenBuild(listener net.Listener, token string) error {
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	reader := bufio.NewScanner(conn)
	if err := bridgeAnswer(conn, reader, token, "initialize", 0, `{"mode":"app"}`); err != nil {
		return err
	}
	return bridgeAnswer(conn, reader, token, "exportBuild", 7, `{"xml":"<PathOfBuilding2/>"}`)
}

func bridgeAnswer(conn net.Conn, reader *bufio.Scanner, token, method string, revision int, result string) error {
	if !reader.Scan() {
		return fmt.Errorf("PoB не получил %s", method)
	}
	var value request
	if err := json.Unmarshal(reader.Bytes(), &value); err != nil {
		return err
	}
	if value.Protocol != 1 || value.Token != token || value.Method != method || value.Params["revision"] != float64(revision) {
		return fmt.Errorf("неверная команда моста %s", method)
	}
	return json.NewEncoder(conn).Encode(response{Protocol: 1, ID: value.ID, Revision: 7, Result: json.RawMessage(result)})
}
