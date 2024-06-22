package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	godap "github.com/google/go-dap"
	"github.com/open-policy-agent/opa/ast/location"
	"github.com/open-policy-agent/opa/debug"
	"github.com/open-policy-agent/opa/logging"
	"github.com/spf13/cobra"
	"github.com/styrainc/regal/internal/dap"
)

func init() {
	verboseLogging := false

	debuggerCommand := &cobra.Command{
		Use:   "debug",
		Short: "Run the Regal OPA Debugger",
		Long:  `Start the Regal OPA debugger and listen on stdin/stdout for client editor messages.`,

		RunE: wrapProfiling(func([]string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			logger := dap.NewDebugLogger(logging.New(), logging.Debug)
			if verboseLogging {
				logger.Local.SetLevel(logging.Debug)
			}

			protoManager := dap.NewProtocolManager(logger.Local)
			logger.ProtocolManager = protoManager

			debugParams := []debug.DebuggerOption{
				debug.SetEventHandler(newEventHandler(protoManager)),
				debug.SetLogger(logger),
			}

			debugger := debug.NewDebugger(debugParams...)

			conn := newCmdConn(os.Stdin, os.Stdout)
			s := newState(ctx, protoManager, debugger, logger)
			return protoManager.Start(ctx, conn, s.messageHandler)
		}),
	}

	debuggerCommand.Flags().BoolVarP(&verboseLogging, "verbose", "v", verboseLogging, "Enable verbose logging")

	RootCommand.AddCommand(debuggerCommand)
}

type state struct {
	ctx                context.Context
	protocolManager    *dap.ProtocolManager
	debugger           debug.Debugger
	session            debug.Session
	logger             *dap.DebugLogger
	serverCapabilities *godap.Capabilities
	clientCapabilities *godap.InitializeRequestArguments
}

func newState(ctx context.Context, protocolManager *dap.ProtocolManager, debugger debug.Debugger, logger *dap.DebugLogger) *state {
	return &state{
		ctx:             ctx,
		protocolManager: protocolManager,
		debugger:        debugger,
		logger:          logger,
		serverCapabilities: &godap.Capabilities{
			SupportsBreakpointLocationsRequest:    true,
			SupportsCancelRequest:                 true,
			SupportsConfigurationDoneRequest:      true,
			SupportsSingleThreadExecutionRequests: true,
			SupportSuspendDebuggee:                true,
			SupportTerminateDebuggee:              true,
			SupportsTerminateRequest:              true,
		},
	}
}

func newEventHandler(pm *dap.ProtocolManager) debug.EventHandler {
	return func(e debug.DebugEvent) {
		switch e.Type {
		case debug.ExceptionEventType:
			pm.SendEvent(dap.NewStoppedExceptionEvent(e.Thread, e.Message))
		case debug.StdoutEventType:
			pm.SendEvent(dap.NewOutputEvent("stdout", e.Message))
		case debug.StoppedEventType:
			pm.SendEvent(dap.NewStoppedEvent(e.Message, e.Thread, nil, "", ""))
		case debug.TerminatedEventType:
			pm.SendEvent(dap.NewTerminatedEvent())
		case debug.ThreadEventType:
			pm.SendEvent(dap.NewThreadEvent(e.Thread, e.Message))
		}
	}
}

func (s *state) messageHandler(message godap.Message) (bool, godap.ResponseMessage, error) {
	var resp godap.ResponseMessage
	var err error
	switch request := message.(type) {
	case *godap.AttachRequest:
		resp = dap.NewAttachResponse()
		err = fmt.Errorf("attach not supported")
	case *godap.BreakpointLocationsRequest:
		resp, err = s.breakpointLocations(request)
	case *godap.ConfigurationDoneRequest:
		// FIXME: Is this when we should start eval?
		resp = dap.NewConfigurationDoneResponse()
	case *godap.ContinueRequest:
		resp, err = s.resume(request)
	case *godap.DisconnectRequest:
		return true, dap.NewDisconnectResponse(), nil
	case *godap.EvaluateRequest:
		resp, err = s.evaluate(request)
	case *godap.InitializeRequest:
		resp, err = s.initialize(request)
	case *godap.LaunchRequest:
		resp, err = s.launch(request)
	case *godap.NextRequest:
		resp, err = s.next(request)
	case *godap.ScopesRequest:
		resp, err = s.scopes(request)
	case *godap.SetBreakpointsRequest:
		resp, err = s.setBreakpoints(request)
	case *godap.StackTraceRequest:
		resp, err = s.stackTrace(request)
	case *godap.StepInRequest:
		resp, err = s.stepIn(request)
	case *godap.StepOutRequest:
		resp, err = s.stepOut(request)
	case *godap.TerminateRequest:
		resp, err = s.terminate(request)
	case *godap.ThreadsRequest:
		resp, err = s.threads(request)
	case *godap.VariablesRequest:
		resp, err = s.variables(request)
	default:
		s.logger.Warn("Handler not found for request: %T", message)
		err = fmt.Errorf("handler not found for request: %T", message)
	}
	return false, resp, err
}

func (s *state) initialize(r *godap.InitializeRequest) (*godap.InitializeResponse, error) {
	if args, err := json.Marshal(r.Arguments); err == nil {
		s.logger.Info("Initializing: %s", args)
	} else {
		s.logger.Info("Initializing")
	}

	s.clientCapabilities = &r.Arguments

	return dap.NewInitializeResponse(*s.serverCapabilities), nil
}

type launchProperties struct {
	Command  string `json:"command"`
	LogLevel string `json:"log_level"`
}

func (s *state) launch(r *godap.LaunchRequest) (*godap.LaunchResponse, error) {
	var props launchProperties
	if err := json.Unmarshal(r.Arguments, &props); err != nil {
		return dap.NewLaunchResponse(), fmt.Errorf("invalid launch properties: %v", err)
	}

	if props.LogLevel != "" {
		s.logger.SetLevelFromString(props.LogLevel)
	} else {
		s.logger.SetRemoteEnabled(false)
	}

	s.logger.Info("Launching: %s", props)

	var err error
	switch props.Command {
	case "eval":
		var evalProps debug.LaunchEvalProperties
		if err := json.Unmarshal(r.Arguments, &evalProps); err != nil {
			return dap.NewLaunchResponse(), fmt.Errorf("invalid launch eval properties: %v", err)
		}
		// FIXME: Should we protect this with a mutex?
		s.session, err = s.debugger.LaunchEval(s.ctx, evalProps)
	case "test":
		err = fmt.Errorf("test not supported")
	case "":
		err = fmt.Errorf("missing launch command")
	default:
		err = fmt.Errorf("unsupported launch command: '%s'", props.Command)
	}

	if err == nil {
		s.protocolManager.SendEvent(dap.NewInitializedEvent())
	}

	err = s.session.ResumeAll()

	return dap.NewLaunchResponse(), err
}

func (s *state) evaluate(_ *godap.EvaluateRequest) (*godap.EvaluateResponse, error) {
	return dap.NewEvaluateResponse(""), fmt.Errorf("evaluate not supported")
}

func (s *state) resume(r *godap.ContinueRequest) (*godap.ContinueResponse, error) {
	return dap.NewContinueResponse(), s.session.Resume(debug.ThreadID(r.Arguments.ThreadId))
}

func (s *state) next(r *godap.NextRequest) (*godap.NextResponse, error) {
	return dap.NewNextResponse(), s.session.StepOver(debug.ThreadID(r.Arguments.ThreadId))
}

func (s *state) stepIn(r *godap.StepInRequest) (*godap.StepInResponse, error) {
	return dap.NewStepInResponse(), s.session.StepIn(debug.ThreadID(r.Arguments.ThreadId))
}

func (s *state) stepOut(r *godap.StepOutRequest) (*godap.StepOutResponse, error) {
	return dap.NewStepOutResponse(), s.session.StepOut(debug.ThreadID(r.Arguments.ThreadId))
}

func (s *state) threads(_ *godap.ThreadsRequest) (*godap.ThreadsResponse, error) {
	var threads []godap.Thread
	ts, err := s.session.Threads()
	if err == nil {
		for _, t := range ts {
			threads = append(threads, godap.Thread{Id: int(t.Id()), Name: t.Name()})
		}
	}
	return dap.NewThreadsResponse(threads), err
}

func (s *state) stackTrace(r *godap.StackTraceRequest) (*godap.StackTraceResponse, error) {
	var stackFrames []godap.StackFrame
	fs, err := s.session.StackTrace(debug.ThreadID(r.Arguments.ThreadId))
	if err == nil {
		for _, f := range fs {
			var source *godap.Source
			source, line, col, endLine, endCol := pos(f.Location())
			stackFrames = append(stackFrames, godap.StackFrame{
				Id:               int(f.ID()),
				Name:             f.Name(),
				Source:           source,
				Line:             line,
				Column:           col,
				EndLine:          endLine,
				EndColumn:        endCol,
				PresentationHint: "normal",
			})
		}
	}
	return dap.NewStackTraceResponse(stackFrames), err
}

func pos(loc *location.Location) (source *godap.Source, line, col, endLine, endCol int) {
	if loc == nil {
		return nil, 1, 0, 1, 0
	}

	if loc.File != "" {
		source = &godap.Source{
			Path: loc.File,
		}
	}

	lines := strings.Split(string(loc.Text), "\n")
	line = loc.Row
	col = loc.Col

	// vs-code will select text if multiple lines are present; avoid this
	//endLine = loc.Row + len(lines) - 1
	//endCol = col + len(lines[len(lines)-1])
	endLine = line
	endCol = col + len(lines[0])
	return
}

func (s *state) scopes(r *godap.ScopesRequest) (*godap.ScopesResponse, error) {
	var scopes []godap.Scope
	ss, err := s.session.Scopes(debug.FrameID(r.Arguments.FrameId))
	if err == nil {
		for _, s := range ss {
			var source *godap.Source
			line := 1
			if loc := s.Location(); loc != nil {
				line = loc.Row
				if loc.File != "" {
					source = &godap.Source{
						Path: loc.File,
					}
				}
			}
			scopes = append(scopes, godap.Scope{
				Name:               s.Name(),
				NamedVariables:     s.NamedVariables(),
				VariablesReference: int(s.VariablesReference()),
				Source:             source,
				Line:               line,
			})
		}
	}
	return dap.NewScopesResponse(scopes), err
}

func (s *state) variables(r *godap.VariablesRequest) (*godap.VariablesResponse, error) {
	var variables []godap.Variable
	vs, err := s.session.Variables(debug.VarRef(r.Arguments.VariablesReference))
	if err == nil {
		for _, v := range vs {
			variables = append(variables, godap.Variable{
				Name:               v.Name(),
				Value:              v.Value(),
				Type:               v.Type(),
				VariablesReference: int(v.VariablesReference()),
			})
		}
	}
	return dap.NewVariablesResponse(variables), err
}

func (s *state) breakpointLocations(request *godap.BreakpointLocationsRequest) (*godap.BreakpointLocationsResponse, error) {
	line := request.Arguments.Line
	s.logger.Debug("Breakpoint locations requested for: %s:%d", request.Arguments.Source.Name, line)

	// TODO: Actually assert where breakpoints can be placed.
	return dap.NewBreakpointLocationsResponse([]godap.BreakpointLocation{
		{
			Line:   line,
			Column: 1,
		},
	}), nil
}

func (s *state) setBreakpoints(request *godap.SetBreakpointsRequest) (*godap.SetBreakpointsResponse, error) {
	locations := make([]location.Location, len(request.Arguments.Breakpoints))
	for i, bp := range request.Arguments.Breakpoints {
		locations[i] = location.Location{
			File: request.Arguments.Source.Path,
			Row:  bp.Line,
		}
	}

	var breakpoints []godap.Breakpoint
	bps, err := s.session.SetBreakpoints(locations)
	if err == nil {
		for _, bp := range bps {
			var source *godap.Source
			line := 1
			l := bp.Location()
			line = bp.Location().Row
			if bp.Location().File != "" {
				source = &godap.Source{
					Path: l.File,
				}
			}
			breakpoints = append(breakpoints, godap.Breakpoint{
				Id:       bp.Id(),
				Source:   source,
				Line:     line,
				Verified: true,
			})
		}
	}

	return dap.NewSetBreakpointsResponse(breakpoints), err
}

func (s *state) terminate(_ *godap.TerminateRequest) (*godap.TerminateResponse, error) {
	return dap.NewTerminateResponse(), s.session.Terminate()
}

type cmdConn struct {
	in  io.ReadCloser
	out io.WriteCloser
}

func newCmdConn(in io.ReadCloser, out io.WriteCloser) *cmdConn {
	return &cmdConn{
		in:  in,
		out: out,
	}
}

func (c *cmdConn) Read(p []byte) (n int, err error) {
	return c.in.Read(p)
}

func (c *cmdConn) Write(p []byte) (n int, err error) {
	return c.out.Write(p)
}

func (c *cmdConn) Close() error {
	var errs []error
	if err := c.in.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := c.out.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors: %v", errs)
	}
	return nil
}
