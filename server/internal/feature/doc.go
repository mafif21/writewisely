// Package feature is the parent of the business modules (package-by-feature).
//
// Each module lives in its own subpackage, internal/feature/<name>/, and owns
// its full stack:
//
//	handler.go     Fiber handlers + swaggo annotations; HTTP only, no business logic
//	service.go     business logic and orchestration
//	repository.go  Postgres access; the only place SQL lives
//	dto.go         API request/response structs (json tags, used in swagger annotations)
//	model.go       the domain entity
//	*_test.go      unit tests, colocated with the code they test
//
// DTOs are always separate from the domain model and mapped explicitly; never
// expose a model or DB row to the client. A persistence struct (DAO) is added
// in repository.go only when the DB shape differs from the model, e.g. jsonb
// or nullable columns; simple entities scan straight into the model.
//
// Feature dependencies are one-way (e.g. submission -> evaluation, never the
// reverse). Interfaces are defined in the consuming package; adapters under
// internal/adapter import the feature to implement them.
//
// No modules exist yet; submission and evaluation arrive in later tasks.
package feature
