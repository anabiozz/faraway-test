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
	"sync"
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
	Data []byte
	Type string
}

func NewClient(
	cfg *Config,
	solverUsecase usecases.SolverUsecase,
	logger Logger,
) *Client {
	return &Client{
		cfg:           cfg,
		solverUsecase: solverUsecase,
		logger:        logger,
	}
}

func (c *Client) Start(ctx context.Context) error {
	var lastErr error
	var wg sync.WaitGroup
	const maxConnections = 10

	for attempt := 0; attempt < maxConnections; attempt++ {
		time.Sleep(3 * time.Second)

		wg.Add(1)
		go func(attempt int) {
			defer wg.Done()

			if attempt > 0 {
				c.logger.Info("retrying connection",
					"attempt", attempt+1,
					"max_attempts", maxConnections)
				time.Sleep(c.cfg.RetryDelay)
			}

			if err := c.executeSession(ctx); err != nil {
				lastErr = NewClientError("Start", err, "session failed")
				c.logger.Error("session error",
					"attempt", attempt+1,
					"error", err)
				return
			}
		}(attempt)
	}

	// Wait for all connection attempts to complete
	wg.Wait()
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
	return s.sendSolutionAndGetResponse(challenge.Type, solution)
}

func (s *ClientSession) receiveChallenge() (*Challenge, error) {
	// Read challenge type
	var challengeType byte
	if err := binary.Read(s.reader, binary.BigEndian, &challengeType); err != nil {
		return nil, NewClientError("receiveChallenge", err, "reading challengeType failed")
	}

	if challengeType != 0x00 && challengeType != 0x01 {
		return nil, NewClientError("receiveChallenge", ErrInvalidChallengeType, "invalid challenge type")
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

	challengeTypeStr := ""
	if challengeType == 0x00 {
		challengeTypeStr = "CPU"
	} else if challengeType == 0x01 {
		challengeTypeStr = "Memory"
	}

	return &Challenge{
		Data: data,
		Type: challengeTypeStr,
	}, nil
}

func (s *ClientSession) solveChallenge(challenge *Challenge) (string, error) {
	if challenge.Type == "CPU" {
		solution := s.client.solverUsecase.FindCPUBoundSolution(challenge.Data)
		if solution == "" {
			return "", NewClientError("solveChallenge", ErrSolutionNotFound, "no solution found for CPU-bound challenge")
		}
		return solution, nil
	} else if challenge.Type == "Memory" {
		solution, err := s.client.solverUsecase.FindMemoryBoundSolution(challenge.Data)
		if err != nil {
			return "", NewClientError("solveChallenge", err, "no solution found for Memory-bound challenge")
		}
		return solution, nil
	} else {
		return "", NewClientError("solveChallenge", ErrInvalidChallengeType, "invalid challenge type")
	}
}

func (s *ClientSession) sendSolutionAndGetResponse(challengeType, solution string) error {
	errCh := make(chan error, 1)

	go func() {
		// Send challenge type
		if _, err := s.writer.WriteString(challengeType + "\n"); err != nil {
			errCh <- NewClientError("sendChallengeTypeAndSolution", err, "sending challenge type failed")
			return
		}

		// Send solution
		if _, err := s.writer.WriteString(solution + "\n"); err != nil {
			errCh <- NewClientError("sendChallengeTypeAndSolution", err, "sending solution failed")
			return
		}

		// Flush the writer
		if err := s.writer.Flush(); err != nil {
			errCh <- NewClientError("sendChallengeTypeAndSolution", err, "flush failed")
			return
		}

		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-s.context.Done():
		return NewClientError("sendChallengeTypeAndSolution", ErrWriteTimeout, "write timeout")
	}

	// Read the server response
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
			return NewClientError("sendChallengeTypeAndSolution", result.err, "reading response failed")
		}
		return s.handleResponse(strings.TrimSpace(result.response))
	case <-s.context.Done():
		return NewClientError("sendChallengeTypeAndSolution", ErrReadTimeout, "read timeout")
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
