package tests

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"url-shortener/internal/lib/random"
	"url-shortener/internal/storage"

	gofakeit "github.com/brianvoe/gofakeit/v6"
	he "github.com/gavv/httpexpect/v2"
)

var baseAddr string

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "urlshortener-test")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	mainPath, _ := filepath.Abs("../cmd/url-shortener/main.go")

	binName := "server"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(tempDir, binName)

	buildCmd := exec.Command("go", "build", "-o", binPath, mainPath)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		fmt.Printf("Build failed: %s\nOutput: %s\n", err, out)
		os.Exit(1)
	}

	port := getFreePort()
	host := fmt.Sprintf("127.0.0.1:%d", port)
	baseAddr = "http://" + host

	configDir := filepath.Join(tempDir, "config")
	if err := os.Mkdir(configDir, 0755); err != nil {
		panic(err)
	}
	configPath := filepath.Join(configDir, "local.yaml")

	storagePath := filepath.Join(tempDir, "storage.db")

	configContent := fmt.Sprintf(`
env: "local"
storage_path: "%s"
http_server:
  address: "%s"
  timeout: 4s
  idle_timeout: 30s
  user: "admin"
  password: "password123"
`, escapePath(storagePath), host)

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		panic(err)
	}

	cmd := exec.Command(binPath)
	cmd.Env = append(os.Environ(), "CONFIG_PATH="+configPath)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	if err := waitForServer(baseAddr); err != nil {
		fmt.Printf("Server failed to start: %v\n", err)
		cmd.Process.Kill()
		os.Exit(1)
	}

	exitCode := m.Run()

	killServer(cmd)

	os.Exit(exitCode)
}

func getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 8082
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 8082
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func waitForServer(url string) error {
	for i := 0; i < 50; i++ {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout")
}

func killServer(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	if runtime.GOOS == "windows" {
		cmd.Process.Kill()
	} else {
		cmd.Process.Signal(syscall.SIGTERM)
	}
}

func escapePath(p string) string {
	if runtime.GOOS == "windows" {
		return filepath.ToSlash(p)
	}
	return p
}

func TestURLShortener_HappyPath(t *testing.T) {
	e := he.Default(t, baseAddr)

	e.POST("/url").
		WithJSON(storage.Request{
			URL:   gofakeit.URL(),
			Alias: random.GenerateRandomString(10),
		}).
		WithBasicAuth("admin", "password123").
		Expect().
		Status(200).
		JSON().Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirectDelete(t *testing.T) {
	testCases := []struct {
		name        string
		url         string
		alias       string
		error       string
		checkErrStr bool
	}{
		{
			name:  "Valid URL",
			url:   "https://google.com",
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:        "Invalid URL",
			url:         "invalid_url",
			alias:       gofakeit.Word(),
			checkErrStr: false,
			error:       "validation failed",
		},
		{
			name:  "Empty Alias",
			url:   "https://google.com",
			alias: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := he.WithConfig(he.Config{
				BaseURL:  baseAddr,
				Reporter: he.NewAssertReporter(t),
				Client: &http.Client{
					CheckRedirect: func(req *http.Request, via []*http.Request) error {
						return http.ErrUseLastResponse
					},
				},
			})

			resp := e.POST("/url").
				WithJSON(storage.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth("admin", "password123").
				Expect().Status(http.StatusOK).
				JSON().Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().NotEmpty()
				return
			}

			alias := tc.alias
			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()
				alias = resp.Value("alias").String().Raw()
			}

			testRedirect(e, alias, tc.url)

			e.DELETE("/"+alias).
				WithBasicAuth("admin", "password123").
				Expect().Status(http.StatusOK).
				JSON().Object().
				Value("status").String().IsEqual("OK")

			testRedirectNotFound(e, alias)
		})
	}
}

func testRedirect(e *he.Expect, alias string, urlToRedirect string) {
	e.GET("/"+alias).
		WithBasicAuth("admin", "password123").
		Expect().
		Status(http.StatusFound).
		Header("Location").IsEqual(urlToRedirect)
}

func testRedirectNotFound(e *he.Expect, alias string) {
	e.GET("/"+alias).
		WithBasicAuth("admin", "password123").
		Expect().
		Status(http.StatusNotFound)
}
