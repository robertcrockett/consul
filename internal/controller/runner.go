// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/go-hclog"
	"golang.org/x/sync/errgroup"

	"github.com/hashicorp/consul/agent/consul/controller/queue"
	"github.com/hashicorp/consul/internal/controller/cache"
	"github.com/hashicorp/consul/internal/resource"
	"github.com/hashicorp/consul/internal/storage"
	"github.com/hashicorp/consul/proto-public/pbresource"
)

// Runtime contains the dependencies required by reconcilers.
type Runtime struct {
	Client pbresource.ResourceServiceClient
	Logger hclog.Logger
	Cache  cache.ReadOnlyCache
}

// controllerRunner contains the actual implementation of running a controller
// including creating watches, calling the reconciler, handling retries, etc.
type controllerRunner struct {
	ctrl   *Controller
	client pbresource.ResourceServiceClient
	logger hclog.Logger
	cache  cache.Cache
}

func newControllerRunner(c *Controller, client pbresource.ResourceServiceClient, defaultLogger hclog.Logger) *controllerRunner {
	return &controllerRunner{
		ctrl:   c,
		client: client,
		logger: c.buildLogger(defaultLogger),
		// Do not build the cache here. If we build/set it when the controller runs
		// then if a controller is restarted it will invalidate the previous cache automatically.
	}
}

func (c *controllerRunner) run(ctx context.Context) error {
	c.logger.Debug("controller running")
	defer c.logger.Debug("controller stopping")

	c.cache = c.ctrl.buildCache()

	group, groupCtx := errgroup.WithContext(ctx)
	recQueue := runQueue[Request](groupCtx, c.ctrl)

	// Managed Type Events → Reconciliation Queue
	group.Go(func() error {
		return c.watch(groupCtx, c.ctrl.managedTypeWatch.watchedType, func(res *pbresource.Resource) {
			recQueue.Add(Request{ID: res.Id})
		})
	})

	for _, w := range c.ctrl.watches {
		mapQueue := runQueue[mapperRequest](groupCtx, c.ctrl)
		watcher := w

		// Watched Type Events → Mapper Queue
		group.Go(func() error {
			return c.watch(groupCtx, watcher.watchedType, func(res *pbresource.Resource) {
				mapQueue.Add(mapperRequest{res: res})
			})
		})

		// Mapper Queue → Mapper → Reconciliation Queue
		group.Go(func() error {
			return c.runMapper(groupCtx, watcher, mapQueue, recQueue, func(ctx context.Context, runtime Runtime, itemType queue.ItemType) ([]Request, error) {
				return watcher.mapper(ctx, runtime, itemType.(mapperRequest).res)
			})
		})
	}

	for _, cw := range c.ctrl.customWatches {
		customMapQueue := runQueue[Event](groupCtx, c.ctrl)
		watcher := cw
		// Custom Events → Mapper Queue
		group.Go(func() error {
			return watcher.source.Watch(groupCtx, func(e Event) {
				customMapQueue.Add(e)
			})
		})

		// Mapper Queue → Mapper → Reconciliation Queue
		group.Go(func() error {
			return c.runCustomMapper(groupCtx, watcher, customMapQueue, recQueue, func(ctx context.Context, runtime Runtime, itemType queue.ItemType) ([]Request, error) {
				return watcher.mapper(ctx, runtime, itemType.(Event))
			})
		})
	}

	// Reconciliation Queue → Reconciler
	group.Go(func() error {
		return c.runReconciler(groupCtx, recQueue)
	})

	return group.Wait()
}

func runQueue[T queue.ItemType](ctx context.Context, ctrl *Controller) queue.WorkQueue[T] {
	base, max := ctrl.backoff()
	return queue.RunWorkQueue[T](ctx, base, max)
}

func (c *controllerRunner) watch(ctx context.Context, typ *pbresource.Type, add func(*pbresource.Resource)) error {
	wl, err := c.client.WatchList(ctx, &pbresource.WatchListRequest{
		Type: typ,
		Tenancy: &pbresource.Tenancy{
			Partition: storage.Wildcard,
			PeerName:  storage.Wildcard,
			Namespace: storage.Wildcard,
		},
	})
	if err != nil {
		c.logger.Error("failed to create watch", "error", err)
		return err
	}

	for {
		event, err := wl.Recv()
		if err != nil {
			c.logger.Warn("error received from watch", "error", err)
			return err
		}

		// Keep the cache up to date. There main reason to do this here is
		// to ensure that any mapper/reconciliation queue deduping wont
		// hide events from being observed and updating the cache state.
		// Therefore we should do this before any queueing.
		switch event.Operation {
		case pbresource.WatchEvent_OPERATION_UPSERT:
			c.cache.Insert(event.Resource)
		case pbresource.WatchEvent_OPERATION_DELETE:
			c.cache.Delete(event.Resource)
		}

		add(event.Resource)
	}
}

func (c *controllerRunner) runMapper(
	ctx context.Context,
	w *watch,
	from queue.WorkQueue[mapperRequest],
	to queue.WorkQueue[Request],
	mapper func(ctx context.Context, runtime Runtime, itemType queue.ItemType) ([]Request, error),
) error {
	logger := c.logger.With("watched_resource_type", resource.ToGVK(w.watchedType))

	for {
		item, shutdown := from.Get()
		if shutdown {
			return nil
		}

		if err := c.doMap(ctx, mapper, to, item, logger); err != nil {
			from.AddRateLimited(item)
			from.Done(item)
			continue
		}

		from.Forget(item)
		from.Done(item)
	}
}

func (c *controllerRunner) runCustomMapper(
	ctx context.Context,
	cw customWatch,
	from queue.WorkQueue[Event],
	to queue.WorkQueue[Request],
	mapper func(ctx context.Context, runtime Runtime, itemType queue.ItemType) ([]Request, error),
) error {
	logger := c.logger.With("watched_event", cw.source)

	for {
		item, shutdown := from.Get()
		if shutdown {
			return nil
		}

		if err := c.doMap(ctx, mapper, to, item, logger); err != nil {
			from.AddRateLimited(item)
			from.Done(item)
			continue
		}

		from.Forget(item)
		from.Done(item)
	}
}

func (c *controllerRunner) doMap(ctx context.Context, mapper func(ctx context.Context, runtime Runtime, itemType queue.ItemType) ([]Request, error), to queue.WorkQueue[Request], item queue.ItemType, logger hclog.Logger) error {
	var reqs []Request
	if err := c.handlePanic(func() error {
		var err error
		reqs, err = mapper(ctx, c.runtime(logger.With("map-request-key", item.Key())), item)
		return err
	}); err != nil {
		return err
	}

	for _, r := range reqs {
		if !resource.EqualType(r.ID.Type, c.ctrl.managedTypeWatch.watchedType) {
			logger.Error("dependency mapper returned request for a resource of the wrong type",
				"type_expected", resource.ToGVK(c.ctrl.managedTypeWatch.watchedType),
				"type_got", resource.ToGVK(r.ID.Type),
			)
			continue
		}
		to.Add(r)
	}
	return nil
}

func (c *controllerRunner) runReconciler(ctx context.Context, queue queue.WorkQueue[Request]) error {
	for {
		req, shutdown := queue.Get()
		if shutdown {
			return nil
		}

		c.logger.Trace("handling request", "request", req)
		err := c.handlePanic(func() error {
			return c.ctrl.reconciler.Reconcile(ctx, c.runtime(c.logger.With("resource-id", req.ID.String())), req)
		})
		if err == nil {
			queue.Forget(req)
		} else {
			var requeueAfter RequeueAfterError
			if errors.As(err, &requeueAfter) {
				queue.Forget(req)
				queue.AddAfter(req, time.Duration(requeueAfter))
			} else {
				queue.AddRateLimited(req)
			}
		}
		queue.Done(req)
	}
}

func (c *controllerRunner) handlePanic(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stack := hclog.Stacktrace()
			c.logger.Error("controller panic",
				"panic", r,
				"stack", stack,
			)
			err = fmt.Errorf("panic [recovered]: %v", r)
			return
		}
	}()

	return fn()
}

func (c *controllerRunner) runtime(logger hclog.Logger) Runtime {
	return Runtime{
		Client: c.client,
		Logger: logger,
		Cache:  c.cache,
	}
}

type mapperRequest struct{ res *pbresource.Resource }

// Key satisfies the queue.ItemType interface. It returns a string which will be
// used to de-duplicate requests in the queue.
func (i mapperRequest) Key() string {
	return fmt.Sprintf(
		"type=%q,part=%q,peer=%q,ns=%q,name=%q,uid=%q",
		resource.ToGVK(i.res.Id.Type),
		i.res.Id.Tenancy.Partition,
		i.res.Id.Tenancy.PeerName,
		i.res.Id.Tenancy.Namespace,
		i.res.Id.Name,
		i.res.Id.Uid,
	)
}
