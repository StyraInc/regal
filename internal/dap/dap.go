// Copyright 2024 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package dap

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"sync"

	"github.com/google/go-dap"
	godap "github.com/google/go-dap"
	"github.com/open-policy-agent/opa/debug"
	"github.com/open-policy-agent/opa/logging"
)

type MessageHandler func(request godap.Message) (bool, godap.ResponseMessage, error)

type ProtocolManager struct {
	//handle  messageHandler
	inChan  chan godap.Message
	outChan chan godap.Message
	logger  logging.Logger
	seq     int
	seqLock sync.Mutex
}

func NewProtocolManager(logger logging.Logger) *ProtocolManager {
	return &ProtocolManager{
		//handle:  handler,
		inChan:  make(chan godap.Message),
		outChan: make(chan godap.Message),
		logger:  logger,
	}
}

func (pm *ProtocolManager) Start(ctx context.Context, conn io.ReadWriteCloser, handle MessageHandler) error {
	reader := bufio.NewReader(conn)
	done := make(chan error)

	go func() {
		for resp := range pm.outChan {
			if pm.logger.GetLevel() == logging.Debug {
				if respData, _ := json.Marshal(resp); respData != nil {
					pm.logger.Debug("Sending %T\n%s", resp, respData)
				} else {
					pm.logger.Debug("Sending %T", resp)
				}
			}
			if err := godap.WriteProtocolMessage(conn, resp); err != nil {
				done <- err
				return
			}
		}
	}()

	go func() {
		for {
			pm.logger.Debug("Waiting for message...")
			req, err := godap.ReadProtocolMessage(reader)
			if err != nil {
				done <- err
				return
			}

			if pm.logger.GetLevel() == logging.Debug {
				if reqData, _ := json.Marshal(req); reqData != nil {
					pm.logger.Debug("Received %T\n%s", req, reqData)
				} else {
					pm.logger.Debug("Received %T", req)
				}
			}

			stop, resp, err := handle(req)
			if err != nil {
				pm.logger.Warn("Error handling request: %v", err)
			}
			pm.SendResponse(resp, req, err)
			if stop {
				done <- err
				return
			}
		}
	}()

	for {
		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (pm *ProtocolManager) SendEvent(e godap.EventMessage) {
	e.GetEvent().Seq = pm.nextSeq()
	pm.outChan <- e
}

func (pm *ProtocolManager) SendResponse(resp godap.ResponseMessage, req godap.Message, err error) {
	if resp == nil {
		return
	}

	if r := resp.GetResponse(); r != nil {
		r.Success = err == nil
		if err != nil {
			r.Message = err.Error()
		}
		r.Seq = pm.nextSeq()
		if req != nil {
			r.RequestSeq = req.GetSeq()
		}
	}
	pm.outChan <- resp
}

func (pm *ProtocolManager) Close() {
	close(pm.outChan)
	close(pm.inChan)
}

func (pm *ProtocolManager) nextSeq() int {
	if pm == nil {
		return 0
	}
	pm.seqLock.Lock()
	defer pm.seqLock.Unlock()
	pm.seq++
	return pm.seq
}

func NewContinueResponse() *godap.ContinueResponse {
	return &godap.ContinueResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "continue",
			Success: true,
		},
	}
}

func NewNextResponse() *godap.NextResponse {
	return &godap.NextResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "next",
			Success: true,
		},
	}
}

func NewStepInResponse() *godap.StepInResponse {
	return &godap.StepInResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "stepIn",
			Success: true,
		},
	}
}

func NewStepOutResponse() *godap.StepOutResponse {
	return &godap.StepOutResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "stepOut",
			Success: true,
		},
	}
}

func NewInitializeResponse(capabilities godap.Capabilities) *godap.InitializeResponse {
	return &godap.InitializeResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				//Seq:  pm.nextSeq(),
				Type: "response",
			},
			Command: "initialize",
			//RequestSeq: r.GetSeq(),
			Success: true,
		},
		Body: capabilities,
	}
}

func NewAttachResponse() *godap.AttachResponse {
	return &godap.AttachResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "attach",
			Success: true,
		},
	}
}

func NewBreakpointLocationsResponse(breakpoints []godap.BreakpointLocation) *godap.BreakpointLocationsResponse {
	return &godap.BreakpointLocationsResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "breakpointLocations",
			Success: true,
		},
		Body: godap.BreakpointLocationsResponseBody{
			Breakpoints: breakpoints,
		},
	}
}

func NewSetBreakpointsResponse(breakpoints []godap.Breakpoint) *godap.SetBreakpointsResponse {
	return &godap.SetBreakpointsResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "setBreakpoints",
			Success: true,
		},
		Body: godap.SetBreakpointsResponseBody{
			Breakpoints: breakpoints,
		},
	}
}

func NewConfigurationDoneResponse() *godap.ConfigurationDoneResponse {
	return &godap.ConfigurationDoneResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "configurationDone",
			Success: true,
		},
	}
}

func NewDisconnectResponse() *godap.DisconnectResponse {
	return &godap.DisconnectResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "disconnect",
			Success: true,
		},
	}
}

func NewEvaluateResponse(value string) *godap.EvaluateResponse {
	return &godap.EvaluateResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "evaluate",
			Success: true,
		},
		Body: godap.EvaluateResponseBody{
			Result: value,
		},
	}
}

func NewLaunchResponse() *godap.LaunchResponse {
	return &godap.LaunchResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "launch",
			Success: true,
		},
	}
}

func NewScopesResponse(scopes []godap.Scope) *godap.ScopesResponse {
	return &godap.ScopesResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "scopes",
			Success: true,
		},
		Body: dap.ScopesResponseBody{
			Scopes: scopes,
		},
	}
}

func NewStackTraceResponse(stack []godap.StackFrame) *godap.StackTraceResponse {
	return &godap.StackTraceResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "stackTrace",
			Success: true,
		},
		Body: godap.StackTraceResponseBody{
			StackFrames: stack,
			TotalFrames: len(stack),
		},
	}
}

func NewTerminateResponse() *godap.TerminateResponse {
	return &godap.TerminateResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "terminate",
			Success: true,
		},
	}
}

func NewThreadsResponse(threads []godap.Thread) *godap.ThreadsResponse {
	return &godap.ThreadsResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "threads",
			Success: true,
		},
		Body: godap.ThreadsResponseBody{
			Threads: threads,
		},
	}
}

func NewVariablesResponse(variables []godap.Variable) *godap.VariablesResponse {
	return &godap.VariablesResponse{
		Response: godap.Response{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "response",
			},
			Command: "variables",
			Success: true,
		},
		Body: godap.VariablesResponseBody{
			Variables: variables,
		},
	}
}

// Events

func NewInitializedEvent() *godap.InitializedEvent {
	return &godap.InitializedEvent{
		Event: godap.Event{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "event",
			},
			Event: "initialized",
		},
	}
}

func NewOutputEvent(category string, output string) *godap.OutputEvent {
	return &godap.OutputEvent{
		Event: godap.Event{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "event",
			},
			Event: "output",
		},
		Body: godap.OutputEventBody{
			Output:   output,
			Category: category,
		},
	}
}

func NewThreadEvent(threadId debug.ThreadID, reason string) *godap.ThreadEvent {
	return &godap.ThreadEvent{
		Event: godap.Event{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "event",
			},
			Event: "thread",
		},
		Body: godap.ThreadEventBody{
			Reason:   reason,
			ThreadId: int(threadId),
		},
	}
}

func NewTerminatedEvent() *godap.TerminatedEvent {
	return &godap.TerminatedEvent{
		Event: godap.Event{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "event",
			},
			Event: "terminated",
		},
	}
}

func NewStoppedEntryEvent(threadId debug.ThreadID) *godap.StoppedEvent {
	return NewStoppedEvent("entry", threadId, nil, "", "")
}

func NewStoppedExceptionEvent(threadId debug.ThreadID, text string) *godap.StoppedEvent {
	return NewStoppedEvent("exception", threadId, nil, "", text)
}

func NewStoppedResultEvent(threadId debug.ThreadID) *godap.StoppedEvent {
	return NewStoppedEvent("result", threadId, nil, "", "")
}

func NewStoppedBreakpointEvent(threadId debug.ThreadID, bp *godap.Breakpoint) *godap.StoppedEvent {
	return NewStoppedEvent("breakpoint", threadId, []int{bp.Id}, "", "")
}

func NewStoppedEvent(reason string, threadId debug.ThreadID, bps []int, description string, text string) *godap.StoppedEvent {
	return &godap.StoppedEvent{
		Event: godap.Event{
			ProtocolMessage: godap.ProtocolMessage{
				Type: "event",
			},
			Event: "stopped",
		},
		Body: godap.StoppedEventBody{
			Reason:            reason,
			ThreadId:          int(threadId),
			Text:              text,
			AllThreadsStopped: true,
			HitBreakpointIds:  bps,
			PreserveFocusHint: false,
		},
	}
}
