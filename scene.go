package tgb

import "fmt"

type Scene struct {
	name   string
	router *Router

	enter Handler
	leave Handler
}

func NewScene(name string, opts ...RouterOption) *Scene {
	return &Scene{
		name:   name,
		router: NewRouter(opts...),
	}
}

func (s *Scene) Use(middlewares ...Middleware) {
	s.router.Use(middlewares...)
}

func (s *Scene) On(filter Filter, handler Handler, middlewares ...Middleware) {
	s.router.On(filter, handler, middlewares...)
}

func (s *Scene) Sub(filter Filter, middlewares ...Middleware) *SubRouter {
	return s.router.Sub(filter, middlewares...)
}

func (s *Scene) Enter(handler Handler) {
	s.enter = handler
}

func (s *Scene) Leave(handler Handler) {
	s.leave = handler
}

func (s *Scene) Handle(ctx *Context) error {
	return s.router.Handle(ctx)
}

type SceneManager struct {
	scenes map[string]*Scene
	getter func(*Context) string
	setter func(*Context, string)
}

func NewSceneManager(getter func(*Context) string, setter func(*Context, string)) *SceneManager {
	return &SceneManager{
		scenes: make(map[string]*Scene),
		getter: getter,
		setter: setter,
	}
}

func (sm *SceneManager) Scene(name string, opts ...RouterOption) *Scene {
	if s, ok := sm.scenes[name]; ok {
		return s
	}

	s := NewScene(name, opts...)
	sm.scenes[name] = s
	return s
}

type SceneControl struct {
	manager *SceneManager
	ctx     *Context
}

func (sc *SceneControl) Enter(name string) error {
	newScene, ok := sc.manager.scenes[name]
	if !ok {
		return fmt.Errorf("scene %s not found", name)
	}

	currentSceneName := sc.manager.getter(sc.ctx)
	if currentSceneName != "" {
		if oldScene, ok := sc.manager.scenes[currentSceneName]; ok && oldScene.leave != nil {
			if err := oldScene.leave(sc.ctx); err != nil {
				return fmt.Errorf("failed to leave scene %s: %w", currentSceneName, err)
			}
		}
	}

	sc.manager.setter(sc.ctx, name)

	if newScene.enter != nil {
		if err := newScene.enter(sc.ctx); err != nil {
			return fmt.Errorf("failed to enter scene %s: %w", name, err)
		}
	}

	return nil
}

func (sc *SceneControl) Leave() error {
	currentSceneName := sc.manager.getter(sc.ctx)
	if currentSceneName == "" {
		return nil
	}

	if oldScene, ok := sc.manager.scenes[currentSceneName]; ok && oldScene.leave != nil {
		if err := oldScene.leave(sc.ctx); err != nil {
			return fmt.Errorf("failed to leave scene %s: %w", currentSceneName, err)
		}
	}

	sc.manager.setter(sc.ctx, "")
	return nil
}

func (sc *SceneControl) Current() string {
	return sc.manager.getter(sc.ctx)
}

type sceneControlKey struct{}

func (sm *SceneManager) Middleware() Middleware {
	return func(next Handler) Handler {
		return func(ctx *Context) error {
			control := &SceneControl{
				manager: sm,
				ctx:     ctx,
			}
			ctx.Set(sceneControlKey{}, control)

			currentSceneName := sm.getter(ctx)
			if currentSceneName != "" {
				if scene, ok := sm.scenes[currentSceneName]; ok {
					return scene.Handle(ctx)
				}
			}

			return next(ctx)
		}
	}
}
