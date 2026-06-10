package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Post("/api/v1/auth/signup", app.signupHandler)
		r.Post("/api/v1/auth/login", app.loginHandler)
	})

	r.Group(func(r chi.Router) {
		r.Use(app.requireAuth)

		r.Route("/api/v1/todos", func(r chi.Router) {
			r.Get("/", app.listTodosHandler)
			r.Post("/", app.createTodoHandler)

			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", app.getTodoHandler)
				r.Put("/", app.updateTodoHandler)
				r.Delete("/", app.deleteTodoHandler)
				r.Patch("/complete", app.completeTodoHandler)
			})

		})

		// r.Route("/api/v1/projects", func(r chi.Router) {
		// 	r.Use(app.requireAuth)
		// 	r.Get("/profile", app.getProfileHandler)
		// 	r.Put("/profile", app.updateProfileHandler)
		// })
	})

	return r
}
