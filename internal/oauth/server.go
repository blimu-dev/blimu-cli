package oauth

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

type CallbackServer struct {
	server   *http.Server
	listener net.Listener
	codeChan chan CallbackResult
	port     int
}

type CallbackResult struct {
	Code  string
	State string
	Error string
}

func NewCallbackServer() (*CallbackServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %w", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	codeChan := make(chan CallbackResult, 1)

	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
	}

	cs := &CallbackServer{
		server:   server,
		listener: listener,
		codeChan: codeChan,
		port:     port,
	}

	mux.HandleFunc("/callback", cs.handleCallback)

	return cs, nil
}

func (cs *CallbackServer) GetRedirectURI() string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", cs.port)
}

func (cs *CallbackServer) Start(ctx context.Context) error {
	go func() {
		if err := cs.server.Serve(cs.listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Callback server error: %v\n", err)
		}
	}()

	return nil
}

func (cs *CallbackServer) WaitForCallback(ctx context.Context) (CallbackResult, error) {
	select {
	case result := <-cs.codeChan:
		return result, nil
	case <-ctx.Done():
		return CallbackResult{}, ctx.Err()
	}
}

func (cs *CallbackServer) Shutdown(ctx context.Context) error {
	return cs.server.Shutdown(ctx)
}

func (cs *CallbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	result := CallbackResult{
		Code:  query.Get("code"),
		State: query.Get("state"),
		Error: query.Get("error"),
	}

	// Send result to channel
	select {
	case cs.codeChan <- result:
	default:
		// Channel is full, ignore
	}

	// Serve success/error page
	if result.Error != "" {
		cs.serveErrorPage(w, result.Error, query.Get("error_description"))
	} else {
		cs.serveSuccessPage(w)
	}
}

func (cs *CallbackServer) serveSuccessPage(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html")
	template := `
<!DOCTYPE html>
<html>
<head>
    <title>Blimu CLI - Authentication Successful</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 40px; text-align: center; }
        .container { max-width: 400px; margin: 0 auto; }
        .success { color: #10b981; font-size: 48px; margin-bottom: 20px; }
        h1 { color: #1f2937; margin-bottom: 10px; }
        p { color: #6b7280; margin-bottom: 30px; }
        .button { background: #3b82f6; color: white; padding: 12px 24px; border: none; border-radius: 8px; text-decoration: none; display: inline-block; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Successful!</h1>
        <p>You have successfully authenticated with Blimu. You can now close this tab and return to your terminal.</p>
        <script>setTimeout(() => window.close(), 3000);</script>
    </div>
</body>
</html>`
	w.Write([]byte(template))
}

func (cs *CallbackServer) serveErrorPage(w http.ResponseWriter, error, description string) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)

	template := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Blimu CLI - Authentication Error</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 40px; text-align: center; }
        .container { max-width: 400px; margin: 0 auto; }
        .error { color: #ef4444; font-size: 48px; margin-bottom: 20px; }
        h1 { color: #1f2937; margin-bottom: 10px; }
        p { color: #6b7280; margin-bottom: 30px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Authentication Error</h1>
        <p><strong>Error:</strong> %s</p>
        <p>%s</p>
        <p>Please close this tab and try again in your terminal.</p>
    </div>
</body>
</html>`, error, description)
	w.Write([]byte(template))
}
