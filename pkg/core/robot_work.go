package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// RobotWorkRegistry contains all the work units registered on a Robot
type RobotWorkRegistry struct {
	sync.RWMutex

	r map[string]*RobotWork
}

const (
	EveryWorkKind = "every"
	AfterWorkKind = "after"
)

// RobotWork and the RobotWork registry represent units of executing computation
// managed at the Robot level. Unlike the utility functions gobot.After and gobot.Every,
// RobotWork units require a context.Context, and can be cancelled externally by calling code.
//
// Usage:
//
//	someWork := myRobot.Every(context.Background(), time.Second * 2, func(){
//		fmt.Println("Here I am doing work")
//	})
//
//	someWork.CallCancelFunc() // Cancel next tick and remove from work registry
//
// goroutines for Every and After are run on their own WaitGroups for synchronization:
//
//	someWork2 := myRobot.Every(context.Background(), time.Second * 2, func(){
//		fmt.Println("Here I am doing more work")
//	})
//
//	somework2.CallCancelFunc()
//
//	// wait for both Every calls to finish
//	robot.WorkEveryWaitGroup().Wait()
type RobotWork struct {
	id         uuid.UUID
	kind       string
	tickCount  int
	ctx        context.Context //nolint:containedctx // done by intention
	cancelFunc context.CancelFunc
	function   func()
	ticker     *time.Ticker
	duration   time.Duration
}

// ID returns the UUID of the RobotWork
func (rw *RobotWork) ID() uuid.UUID {
	return rw.id
}

// CancelFunc returns the context.CancelFunc used to cancel the work
func (rw *RobotWork) CancelFunc() context.CancelFunc {
	return rw.cancelFunc
}

// CallCancelFunc calls the context.CancelFunc used to cancel the work
func (rw *RobotWork) CallCancelFunc() {
	rw.cancelFunc()
}

// Ticker returns the time.Ticker used in an Every so that calling code can sync on the same channel
func (rw *RobotWork) Ticker() *time.Ticker {
	if rw.kind == AfterWorkKind {
		return nil
	}
	return rw.ticker
}

// TickCount returns the number of times the function successfully ran
func (rw *RobotWork) TickCount() int {
	return rw.tickCount
}

// Duration returns the timeout until an After fires or the period of an Every
func (rw *RobotWork) Duration() time.Duration {
	return rw.duration
}

func (rw *RobotWork) String() string {
	format := `ID: %s
Kind: %s
TickCount: %d

`
	return fmt.Sprintf(format, rw.id, rw.kind, rw.tickCount)
}

// WorkRegistry returns the Robot's WorkRegistry
func (r *Robot) WorkRegistry() *RobotWorkRegistry {
	return r.workRegistry
}

// Every calls the given function for every tick of the provided duration.
func (r *Robot) Every(ctx context.Context, d time.Duration, f func()) *RobotWork {
	// Ensure we have a valid context
	if ctx == nil {
		ctx = context.Background()
	}
	
	// Create a combined context that cancels when either the robot or user context is cancelled
	combinedCtx, combinedCancel := context.WithCancel(ctx)
	
	rw := r.workRegistry.registerEvery(combinedCtx, d, f)
	// Override the cancel function to use our combined one
	rw.cancelFunc = combinedCancel
	
	r.WorkEveryWaitGroup.Add(1)
	go func() {
		defer r.WorkEveryWaitGroup.Done()
		defer rw.ticker.Stop()
		defer r.workRegistry.delete(rw.id)
		
		// Also listen to robot's context for shutdown
		mergedCtx, mergedCancel := context.WithCancel(ctx)
		defer mergedCancel()
		
		go func() {
			select {
			case <-r.ctx.Done():
				mergedCancel()
			case <-mergedCtx.Done():
			}
		}()
		
		for {
			select {
			case <-rw.ctx.Done():
				return
			case <-mergedCtx.Done():
				return
			case <-rw.ticker.C:
				// Safe function execution with panic recovery
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Log panic but don't crash the robot
							fmt.Printf("Panic in Every work function: %v\n", r)
						}
					}()
					f()
					rw.tickCount++
				}()
			}
		}
	}()
	return rw
}

// After calls the given function after the provided duration has elapsed
func (r *Robot) After(ctx context.Context, d time.Duration, f func()) *RobotWork {
	// Ensure we have a valid context
	if ctx == nil {
		ctx = context.Background()
	}
	
	// Create a combined context that cancels when either the robot or user context is cancelled
	combinedCtx, combinedCancel := context.WithCancel(ctx)
	
	rw := r.workRegistry.registerAfter(combinedCtx, d, f)
	// Override the cancel function to use our combined one
	rw.cancelFunc = combinedCancel
	
	ch := time.After(d)
	r.WorkAfterWaitGroup.Add(1)
	go func() {
		defer r.WorkAfterWaitGroup.Done()
		defer r.workRegistry.delete(rw.id)
		
		// Also listen to robot's context for shutdown
		mergedCtx, mergedCancel := context.WithCancel(ctx)
		defer mergedCancel()
		
		go func() {
			select {
			case <-r.ctx.Done():
				mergedCancel()
			case <-mergedCtx.Done():
			}
		}()
		
		for {
			select {
			case <-rw.ctx.Done():
				return
			case <-mergedCtx.Done():
				return
			case <-ch:
				// Safe function execution with panic recovery
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Log panic but don't crash the robot
							fmt.Printf("Panic in After work function: %v\n", r)
						}
					}()
					f()
					rw.tickCount++
				}()
				return // After only runs once
			}
		}
	}()
	return rw
}

// Get returns the RobotWork specified by the provided ID. To delete something from the registry, it's
// necessary to call its context.CancelFunc, which will perform a goroutine-safe delete on the underlying
// map.
func (rwr *RobotWorkRegistry) Get(id uuid.UUID) *RobotWork {
	rwr.Lock()
	defer rwr.Unlock()
	return rwr.r[id.String()]
}

// Delete returns the RobotWork specified by the provided ID
func (rwr *RobotWorkRegistry) delete(id uuid.UUID) {
	rwr.Lock()
	defer rwr.Unlock()
	delete(rwr.r, id.String())
}

// registerAfter creates a new unit of RobotWork and sets up its context/cancellation
func (rwr *RobotWorkRegistry) registerAfter(ctx context.Context, d time.Duration, f func()) *RobotWork {
	rwr.Lock()
	defer rwr.Unlock()

	id := uuid.New()
	rw := &RobotWork{
		id:       id,
		kind:     AfterWorkKind,
		function: f,
		duration: d,
	}

	rw.ctx, rw.cancelFunc = context.WithCancel(ctx)
	rwr.r[id.String()] = rw
	return rw
}

// registerEvery creates a new unit of RobotWork and sets up its context/cancellation
func (rwr *RobotWorkRegistry) registerEvery(ctx context.Context, d time.Duration, f func()) *RobotWork {
	rwr.Lock()
	defer rwr.Unlock()

	id := uuid.New()
	rw := &RobotWork{
		id:       id,
		kind:     EveryWorkKind,
		function: f,
		duration: d,
		ticker:   time.NewTicker(d),
	}

	rw.ctx, rw.cancelFunc = context.WithCancel(ctx)

	rwr.r[id.String()] = rw
	return rw
}
