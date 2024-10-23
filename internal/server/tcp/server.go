package tcp

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"faraway/internal/usecases"
	"fmt"
	"net"
	"strings"
	"time"
)

type Server struct {
	cfg          *Config
	powUsecase   usecases.PowUsecase
	quoteUsecase usecases.QuoteUsecase
	logger       Logger
}

type Config struct {
	Address    string
	KeepAlive  time.Duration
	Deadline   time.Duration
	BufferSize int
}

type Logger interface {
	Error(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
}

type Challenge struct {
	Difficulty uint32
	Challenge  []byte
}

func NewServer(cfg *Config, powUsecase usecases.PowUsecase, quoteUsecase usecases.QuoteUsecase, logger Logger) *Server {
	return &Server{
		cfg:          cfg,
		powUsecase:   powUsecase,
		quoteUsecase: quoteUsecase,
		logger:       logger,
	}
}

func (s *Server) Run(ctx context.Context) error {
	lc := net.ListenConfig{
		KeepAlive: s.cfg.KeepAlive,
	}

	listener, err := lc.Listen(ctx, "tcp", s.cfg.Address)
	if err != nil {
		return NewConnectionError("Run", err, "failed to start listener")
	}
	defer listener.Close()

	s.logger.Info("server started", "address", s.cfg.Address)

	return s.serve(ctx, listener)
}

func (s *Server) serve(ctx context.Context, listener net.Listener) error {
	for {
		select {
		case <-ctx.Done():
			return NewConnectionError("serve", ErrServerShutdown, "context cancelled")
		default:
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					s.logger.Debug("listener closed")
					return nil
				}
				s.logger.Error("accept failed", "error", err)
				continue
			}
			go s.handleConnection(conn)
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			s.logger.Error("connection close failed",
				"error", NewConnectionError("handleConnection", err, "cleanup failed"))
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Deadline)
	defer cancel()

	if err := conn.SetDeadline(time.Now().Add(s.cfg.Deadline)); err != nil {
		s.logger.Error("set deadline failed",
			"error", NewConnectionError("handleConnection", err, "setting timeout failed"))
		return
	}

	session := &Session{
		conn:    conn,
		reader:  bufio.NewReader(conn),
		writer:  bufio.NewWriter(conn),
		server:  s,
		context: ctx,
	}

	if err := session.Handle(); err != nil {
		s.handleError(session.writer, err)
	}
}

type Session struct {
	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	server  *Server
	context context.Context
}

// All magic happens here
func (s *Session) Handle() error {
	// Step 1: Send challenge
	challenge, err := s.sendChallenge()
	if err != nil {
		return fmt.Errorf("failed to send challenge: %w", err)
	}

	// Step 2: Read solution
	solution, err := s.readSolution()
	if err != nil {
		return fmt.Errorf("failed to read solution: %w", err)
	}

	// Step 3: Validate and respond
	err = s.validateAndRespond(challenge, solution)
	if err != nil {
		return fmt.Errorf("failed to validate and respond: %w", err)
	}

	return nil
}

func (s *Session) sendChallenge() ([]byte, error) {
	pow, err := s.server.powUsecase.GenerateChallenge()
	if err != nil {
		return nil, NewConnectionError("sendChallenge", ErrChallengeFailed, "generation failed")
	}

	// Send challenge difficulty because we can increase the difficulty on demand (on demand not implemented yet)
	if err := binary.Write(s.writer, binary.BigEndian, pow.Difficulty); err != nil {
		return nil, NewConnectionError("sendChallenge", ErrChallengeDelivery, "write length failed")
	}

	// Send challenge length
	length := int32(len(pow.Challenge))
	if err := binary.Write(s.writer, binary.BigEndian, length); err != nil {
		return nil, NewConnectionError("sendChallenge", ErrChallengeDelivery, "write length failed")
	}

	// Send challenge data
	errCh := make(chan error, 1)
	go func() {
		_, err := s.writer.Write(pow.Challenge)
		if err == nil {
			err = s.writer.Flush()
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, NewConnectionError("sendChallenge", ErrChallengeDelivery, "write data failed")
		}
	case <-s.context.Done():
		return nil, NewConnectionError("sendChallenge", ErrWriteTimeout, "context deadline exceeded")
	}

	return pow.Challenge, nil
}

func (s *Session) readSolution() ([]byte, error) {
	resultCh := make(chan struct {
		solution []byte
		err      error
	}, 1)

	go func() {
		line, err := s.reader.ReadString('\n')
		if err != nil {
			resultCh <- struct {
				solution []byte
				err      error
			}{nil, NewConnectionError("readSolution", err, "read failed")}
			return
		}

		solution, err := parseSolution(line)
		resultCh <- struct {
			solution []byte
			err      error
		}{solution, err}
	}()

	select {
	case result := <-resultCh:
		return result.solution, result.err
	case <-s.context.Done():
		return nil, NewConnectionError("readSolution", ErrReadTimeout, "context deadline exceeded")
	}
}

func (s *Session) validateAndRespond(challenge, solution []byte) error {
	if !s.server.powUsecase.ValidateSolution(challenge, solution) {
		return NewConnectionError("validateAndRespond", ErrInvalidSolution, "validation failed")
	}

	quote := s.server.quoteUsecase.GetRandomQuote()
	response := formatSuccessResponse(quote)

	errCh := make(chan error, 1)
	go func() {
		_, err := s.writer.WriteString(response)
		if err == nil {
			err = s.writer.Flush()
		}
		errCh <- err
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return NewConnectionError("validateAndRespond", err, "write response failed")
		}
	case <-s.context.Done():
		return NewConnectionError("validateAndRespond", ErrWriteTimeout, "context deadline exceeded")
	}

	return nil
}

func (s *Server) handleError(writer *bufio.Writer, err error) {
	response := ToErrorResponse(err)
	s.logger.Error("client error",
		"code", response.Code,
		"message", response.Message,
		"error", err)

	if err := sendErrorResponse(writer, response); err != nil {
		s.logger.Error("failed to send error response", "error", err)
	}
}

// Helper functions

func parseSolution(line string) ([]byte, error) {
	return []byte(strings.TrimSpace(line)), nil
}

func formatSuccessResponse(quote string) string {
	return fmt.Sprintf("SUCCESS:%s\n", quote)
}

func sendErrorResponse(writer *bufio.Writer, response ErrorResponse) error {
	_, err := writer.WriteString(fmt.Sprintf("ERROR:%s:%s\n", response.Code, response.Message))
	if err != nil {
		return err
	}
	return writer.Flush()
}
