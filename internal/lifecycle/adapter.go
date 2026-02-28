package lifecycle

import (
	"context"
	"sync"
)

// Startable represents components with Start(*sync.WaitGroup) / Stop() signatures.
type Startable interface {
	Start(wg *sync.WaitGroup)
	Stop()
}

// NewSimpleComponent creates a Component adapter from a Startable.
func NewSimpleComponent(name string, s Startable) Component {
	return &SimpleComponent{
		ComponentName: name,
		StartFunc:     s.Start,
		StopFunc:      s.Stop,
	}
}

// SimpleComponent adapts components with Start(*sync.WaitGroup) / Stop() signatures.
type SimpleComponent struct {
	ComponentName string
	StartFunc     func(wg *sync.WaitGroup)
	StopFunc      func()
}

func (c *SimpleComponent) Name() string { return c.ComponentName }

func (c *SimpleComponent) Start(_ context.Context, wg *sync.WaitGroup) error {
	c.StartFunc(wg)
	return nil
}

func (c *SimpleComponent) Stop(_ context.Context) error {
	c.StopFunc()
	return nil
}

// NewFuncComponent creates a Component from start/stop functions.
func NewFuncComponent(
	name string,
	startFn func(ctx context.Context, wg *sync.WaitGroup) error,
	stopFn func(ctx context.Context) error,
) *FuncComponent {
	return &FuncComponent{ComponentName: name, StartFunc: startFn, StopFunc: stopFn}
}

// FuncComponent wraps arbitrary start/stop functions.
type FuncComponent struct {
	ComponentName string
	StartFunc     func(ctx context.Context, wg *sync.WaitGroup) error
	StopFunc      func(ctx context.Context) error
}

func (c *FuncComponent) Name() string { return c.ComponentName }

func (c *FuncComponent) Start(ctx context.Context, wg *sync.WaitGroup) error {
	return c.StartFunc(ctx, wg)
}

func (c *FuncComponent) Stop(ctx context.Context) error {
	if c.StopFunc != nil {
		return c.StopFunc(ctx)
	}
	return nil
}

// ErrorComponent adapts components with Start(context.Context) error / Stop() signatures.
type ErrorComponent struct {
	ComponentName string
	StartFunc     func(ctx context.Context) error
	StopFunc      func()
}

func (c *ErrorComponent) Name() string { return c.ComponentName }

func (c *ErrorComponent) Start(ctx context.Context, _ *sync.WaitGroup) error {
	return c.StartFunc(ctx)
}

func (c *ErrorComponent) Stop(_ context.Context) error {
	c.StopFunc()
	return nil
}
