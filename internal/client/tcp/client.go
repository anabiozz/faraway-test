package tcp

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"faraway/internal/usecases"
	"io"
	"net"
	"strings"
	"time"
)

type Client struct {
	cfg           *Config
	solverUsecase usecases.SolverUsecase
	logger        Logger
}

type Config struct {
	ServerAddr     string
	ConnectTimeout time.Duration
	RequestTimeout time.Duration
	RetryAttempts  int
	RetryDelay     time.Duration
	MaxMessageSize int64
	BufferSize     int
}

type Logger interface {
	Error(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type Challenge struct {
	Difficulty uint64
	Data       []byte
}

func NewClient(cfg *Config, solverUsecase usecases.SolverUsecase, logger Logger) *Client {
	return &Client{
		cfg:           cfg,
		solverUsecase: solverUsecase,
		logger:        logger,
	}
}

func (c *Client) Start(ctx context.Context) error {
	var lastErr error
	for attempt := 0; attempt < c.cfg.RetryAttempts; attempt++ {
		if attempt > 0 {
			c.logger.Info("retrying connection",
				"attempt", attempt+1,
				"max_attempts", c.cfg.RetryAttempts)
			time.Sleep(c.cfg.RetryDelay)
		}

		if err := c.executeSession(ctx); err != nil {
			lastErr = NewClientError("Start", err, "session failed")
			c.logger.Error("session error",
				"attempt", attempt+1,
				"error", err)
			continue
		}
		return nil
	}
	return lastErr
}

func (c *Client) executeSession(ctx context.Context) error {
	connectCtx, cancel := context.WithTimeout(ctx, c.cfg.ConnectTimeout)
	defer cancel()

	conn, err := c.connect(connectCtx)
	if err != nil {
		return err
	}
	defer conn.Close()

	session := &ClientSession{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		writer:  bufio.NewWriter(conn),
		client:  c,
		context: ctx,
	}

	return session.Execute()
}

func (c *Client) connect(ctx context.Context) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", c.cfg.ServerAddr)
	if err != nil {
		return nil, NewClientError("connect", err, "connection failed")
	}

	if err := conn.SetDeadline(time.Now().Add(c.cfg.RequestTimeout)); err != nil {
		conn.Close()
		return nil, NewClientError("connect", err, "setting timeout failed")
	}

	return conn, nil
}

type ClientSession struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	client  *Client
	context context.Context
}

// All magic happens here
func (s *ClientSession) Execute() error {
	// Step 1: Receive challenge
	challenge, err := s.receiveChallenge()
	if err != nil {
		return err
	}

	// Step 2: Solve challenge
	solution, err := s.solveChallenge(challenge)
	if err != nil {
		return err
	}

	// Step 3: Send solution and receive response
	return s.sendSolutionAndGetResponse(solution)
}

func (s *ClientSession) receiveChallenge() (*Challenge, error) {
	// Read difficulty
	var difficulty uint64
	if err := binary.Read(s.reader, binary.BigEndian, &difficulty); err != nil {
		return nil, NewClientError("receiveChallenge", err, "reading difficulty failed")
	}

	// Read challenge length
	var length int32
	if err := binary.Read(s.reader, binary.BigEndian, &length); err != nil {
		return nil, NewClientError("receiveChallenge", err, "reading length failed")
	}

	if length <= 0 || length > int32(s.client.cfg.MaxMessageSize) {
		return nil, NewClientError("receiveChallenge", ErrInvalidMessageSize, "invalid challenge size")
	}

	// Read challenge data
	data := make([]byte, length)
	n, err := io.ReadFull(s.reader, data)
	if err != nil {
		if err == io.EOF {
			return nil, NewClientError("receiveChallenge", ErrConnectionClosed, "unexpected EOF")
		}
		return nil, NewClientError("receiveChallenge", err, "reading challenge failed")
	}

	if n != int(length) {
		return nil, NewClientError("receiveChallenge", ErrInvalidMessageSize,
			"challenge size mismatch")
	}

	return &Challenge{
		Difficulty: difficulty,
		Data:       data,
	}, nil
}

func (s *ClientSession) solveChallenge(challenge *Challenge) (string, error) {
	solution := s.client.solverUsecase.FindSolution(challenge.Data, challenge.Difficulty)
	if solution == "" {
		return "", NewClientError("solveChallenge", ErrSolutionNotFound, "no solution found")
	}

	return solution, nil
}

func (s *ClientSession) sendSolutionAndGetResponse(solution string) error {
	errCh := make(chan error, 1)
	go func() {
		_, err := s.writer.WriteString(solution + "\n")
		if err == nil {
			err = s.writer.Flush()
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return NewClientError("sendSolutionAndGetResponse", err, "sending solution failed")
		}
	case <-s.context.Done():
		return NewClientError("sendSolutionAndGetResponse", ErrWriteTimeout, "write timeout")
	}

	responseCh := make(chan struct {
		response string
		err      error
	}, 1)

	go func() {
		response, err := s.reader.ReadString('\n')
		responseCh <- struct {
			response string
			err      error
		}{response, err}
	}()

	select {
	case result := <-responseCh:
		if result.err != nil {
			return NewClientError("sendSolutionAndGetResponse", result.err, "reading response failed")
		}
		return s.handleResponse(strings.TrimSpace(result.response))
	case <-s.context.Done():
		return NewClientError("sendSolutionAndGetResponse", ErrReadTimeout, "read timeout")
	}
}

func (s *ClientSession) handleResponse(response string) error {
	if strings.HasPrefix(response, "SUCCESS:") {
		quote := strings.TrimPrefix(response, "SUCCESS:")
		s.client.logger.Info("received quote", "quote", quote)
		return nil
	}

	if strings.HasPrefix(response, "ERROR:") {
		parts := strings.SplitN(strings.TrimPrefix(response, "ERROR:"), ":", 2)
		if len(parts) != 2 {
			return NewClientError("handleResponse", ErrInvalidProtocol, "invalid error format")
		}
		return NewClientError("handleResponse", errors.New(parts[0]), parts[1])
	}

	return NewClientError("handleResponse", ErrInvalidProtocol, "invalid response format")
}
