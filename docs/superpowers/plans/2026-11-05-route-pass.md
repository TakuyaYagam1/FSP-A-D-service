# Route Pass Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Владелец вручную выпускает постоянный пропуск после готовности маршрута; владелец и два назначенных участника видят и скачивают его на шаге 4.

**Architecture:** Postgres хранит единственный факт выпуска по route_slot_id; оба варианта usecase используют собственную действующую проверку готовности, а право чтения проверяют по идентификаторам владельца и участников. UI получает данные выпуска с сервера и локально воспроизводит PNG; прежний независимый черновик остается в отдельной панели.

**Tech Stack:** Go, PostgreSQL/goose/sqlc, ogen/OpenAPI, Redis, vanilla JavaScript/Canvas, Docker Compose.

**Spec:** [Route pass design](../specs/2026-11-05-route-pass-design.md)

## Global Constraints

- Один выпуск на поставку; время выпуска не меняется при повторном POST.
- GET доступен владельцу, carrier и inspector; POST только владельцу и только при актуальной готовности, кроме повторного запроса уже выпущенного пропуска.
- Данные остаются после 15-минутного TTL, не включают slot_code и не заявляют о двух уникальных подписях.
- Существующие approval writes, code release, checker и задуманные различия vulnerable/patch не изменять.
- Существующие Docker volumes и текущую БД не удалять; patch backend не запускать против рабочей БД.
- Не создавать коммит без отдельного запроса владельца проекта.

## Review Focus

- Повторный POST после истечения TTL возвращает прежний выпуск, а не ошибку или новую дату -> тест Task 2.
- Два одновременных POST создают один выпуск и возвращают одну дату -> интеграционный тест Task 2.
- Чужой или неавторизованный аккаунт не видит даже готовность чужой поставки -> API-тест Task 3.
- Дубли подписи уязвимой версии и подтверждение одной роли исправленной версии дают различную готовность -> usecase-тест Task 2.
- Вкладка участника, открытая после выпуска или после перезагрузки и TTL, восстанавливает доступ без sessionStorage владельца -> browser-тест Task 4.

---

### Task 1: Persistence and schema

**Files (in both `wahrwelt/backend` and `wahrwelt-patch/backend`):**
- Create: `db/migrations/000002_route_passes.sql`, `db/queries/route_pass.sql`, `internal/domain/route_pass.go`, `internal/port/outbound/route_pass.go`, `internal/adapter/secondary/postgres/route_pass.go`
- Generated: `internal/adapter/secondary/postgres/sqlc/*`
- Test: `internal/adapter/secondary/postgres/route_pass_test.go` (isolated DB when available)

**Interfaces:** `domain.RoutePass{SlotID, CaseID uuid.UUID; Kind, Origin, Destination, CarrierUsername, InspectorUsername string; IssuedAt time.Time}`. `outbound.RoutePassRepository` has `GetRoutePass(ctx context.Context, slotID uuid.UUID) (domain.RoutePass, error)` and `CreateRoutePass(ctx context.Context, slotID uuid.UUID) (domain.RoutePass, error)`; queries join case and participant names but never return slot_code. Persistence provides issued_at and immutable route metadata.

- [ ] Add failing repository tests for missing pass, repeated issue and concurrent creation with identical issued_at.
- [ ] Run `go test ./internal/adapter/secondary/postgres/...` in isolated test DB; expect failure from missing table or method.
- [ ] Add additive migration with unique route_slot_id FK, timestamp default now(), sqlc queries and adapter; run `make generate` in each backend.
- [ ] Re-run repository tests and `make generate-check` in both; expect PASS and no drift.

### Task 2: Readiness, access and issuance

**Files (in both backend trees):**
- Modify: `internal/usecase/route_slot.go` to expose `Status(ctx context.Context, actorID, slotID uuid.UUID) (domain.RouteSlotStatus, error)` using the existing version-specific count rules; reuse it from owner list where appropriate.
- Create: `internal/usecase/route_pass.go`, `internal/usecase/route_pass_test.go`
- Modify: `internal/bootstrap/wire.go` and generated `wire_gen.go`

**Interfaces:** `usecase.NewRoutePass(slots *usecase.RouteSlot, cases outbound.CaseRepository, users outbound.UserRepository, passes outbound.RoutePassRepository) *usecase.RoutePass`; `Get(ctx, actorID, slotID) (domain.RoutePassState, error)` and `Issue(ctx, actorID, slotID) (domain.RoutePassState, error)`. `RoutePassState{Ready bool; Pass *RoutePass}`. Status checks actor == owner/carrier/inspector before reading Redis; Issue checks actor == owner, then existing pass, then ready, then inserts idempotently.

- [ ] Write tests: owner/carrier/inspector GET, foreign/zero actor hidden, early POST rejected, two concurrent issue calls, repeated POST after expiry returns identical pass, distinct-role patch rule versus vulnerable count rule.
- [ ] Run `go test ./internal/usecase` in each backend; expect failure from missing symbols.
- [ ] Implement status reuse and issuance; bind repository/usecase in Wire.
- [ ] Run usecase tests, `go vet ./...` and `make generate-check`; expect PASS.

### Task 3: Public read and issue API

**Files (in both backend trees):**
- Modify: `api/openapi.yml`, `internal/adapter/primary/http/v1/router.go`, `internal/bootstrap/wire.go`
- Create: `internal/adapter/primary/http/v1/route_pass.go`, `internal/adapter/primary/http/v1/route_pass_test.go`
- Generated: `internal/adapter/primary/http/api/*`, `internal/bootstrap/wire_gen.go`

**Interfaces:** `GET /api/v1/route-slots/{slot_id}/pass` -> `{ready:boolean, issued:boolean, pass?:{route_slot_id,case_id,kind,origin,destination,carrier_username,inspector_username,issued_at}}`; `POST` same URL, no body, returns the same shape with issued=true. Auth cookie matches existing route-slot endpoints. Foreign/missing -> 404, unauthenticated -> 401, premature issue -> current domain error mapping.

- [ ] Write failing HTTP tests for owner, both participants, foreign, unauthenticated, before-ready, repeated issue, and absence of slot_code in JSON.
- [ ] Run `go test ./internal/adapter/primary/http/v1`; expect missing handlers or route.
- [ ] Extend OpenAPI and v1 handlers, regenerate ogen/sqlc/Wire with `make generate` in each backend.
- [ ] Run HTTP tests and both `make generate-check`; expect PASS.

### Task 4: Step 4 and PNG

**Files (in both frontend trees):**
- Modify: `index.html`, `app.js`, `styles.css`, `Dockerfile`
- Create: `route-pass.js`, `route-pass.test.cjs`
- Preserve: `document-preview.js` and the owner-only unsigned preview flow.

**Interfaces:** `loadRoutePass(slotID)` GETs server state on selected slot or participant link and after approval/issuance; `issueRoutePass()` POSTs only as owner when ready. `buildRoutePassPNG(pass)` produces a PNG download with label `МАРШРУТ ГОТОВ`, case and route identifiers, assigned usernames, issued time, and no code or signature claim. Step 4 remains disabled until ready or issued; issued pass stays available after Redis expiry. The previous `К черновику PNG` action remains in step 3, opening an unnumbered draft pane.

- [ ] Add failing DOM/Canvas tests for locked, ready-unissued owner, waiting participant, issued participant after reload, expired-but-issued, and rendering text without code.
- [ ] Run `node --test <each frontend>/route-pass.test.cjs`; expect failure from absent module/controls.
- [ ] Add the pane and handlers in each frontend, style 390px and desktop without overflow, copy asset into both images.
- [ ] Run Node tests, `node --check` on changed scripts, and `git diff --check`; expect PASS.

### Task 5: Live validation and safe handoff

**Files:** No production source changes; adjust tests only for demonstrated failures.

- [ ] Run `go test ./...`, `go vet ./...`, `make generate-check` in both backends; run all checker tests from the repository root.
- [ ] Build both backend/frontend images; deploy only main compose with its existing named volumes and check all health statuses.
- [ ] Browser test main UI with owner plus two existing participant accounts: issue once, reload in both sessions, download identical pass metadata after approval TTL mocked or expired; unauthorized account receives 404; old draft still says `НЕ СОГЛАСОВАНО`.
- [ ] Verify mobile 390px, desktop, no console errors; do not reset Docker or migrate the patch backend on the live PostgreSQL volume.
- [ ] Review final diff and preserve user-owned files; report skipped checks and the explicit fact that a client-rendered PNG is not a cryptographic signature.
