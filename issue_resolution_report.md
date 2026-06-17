# Issue Resolution Report

## Issue 1: Multiple API calls on Dashboard
### Problem
The Dashboard page was making redundant and overlapping network requests for groups and shared groups, causing performance degradation and UI flicker.
### Root Cause
Missing or improperly configured dependency arrays (`[]`) within React `useEffect` hooks, triggering re-renders that invoked API calls repeatedly.
### Solution Implemented
Optimized `useEffect` hooks and implemented structured loading states to prevent redundant data fetching.
### Technical Changes
Updated `DashboardPage.jsx` to correctly utilize dependencies and wrap fetch logic in conditional blocks that evaluate existing data presence and loading booleans.
### Outcome
Network traffic drastically reduced, leading to a smooth, flicker-free dashboard load experience.

---

## Issue 2: Current Date Validation Bug
### Problem
Users were able to select deadlines in the past when creating or editing tasks.
### Root Cause
The HTML5 `<input type="date">` component lacked dynamic validation attributes enforcing a minimum acceptable date.
### Solution Implemented
Dynamically injected the current date into the `min` attribute of the date picker.
### Technical Changes
Added `min={new Date().toISOString().split('T')[0]}` to date inputs in `TodoFormPage.jsx` and `EditGroupModal.jsx`.
### Outcome
Invalid temporal states are strictly blocked on the client side before they can be transmitted to the server.

---

## Issue 3: Missing Edit Group Feature
### Problem
Users had no ability to modify a group's title, description, or due date after initial creation.
### Root Cause
The feature was scoped out of the initial MVP, lacking both a UI interface and frontend-backend wiring.
### Solution Implemented
Developed an Edit Group modal overlay and connected it to the REST API.
### Technical Changes
Created `EditGroupModal.jsx`, implemented state management to pre-fill existing data, and wired the form submission to `PUT /api/groups/:id`.
### Outcome
Users possess full CRUD capabilities over parent group metadata.

---

## Issue 4: Missing Collaborator Role Update Feature
### Problem
Group owners could not alter a collaborator's access level (e.g., from VIEW to EDIT) once shared.
### Root Cause
The UI lacked a control mechanism to issue role update commands to the backend.
### Solution Implemented
Introduced an inline role-management dropdown visible only to group owners.
### Technical Changes
Modified `GroupDetailsPage.jsx` to render a `<select>` element beside collaborator names. Emits a `PATCH /api/groups/:id/share/:userId/role` request upon change.
### Outcome
Dynamic, seamless permission management enabling granular access control.

---

## Issue 5: Profile Page displaying User ID
### Problem
The application exposed internal database integer IDs on the User Profile screen.
### Root Cause
Direct deserialization of the raw user payload to the UI without a dedicated presentation filter.
### Solution Implemented
Sanitized the profile view to only display relevant human-readable identifiers.
### Technical Changes
Refactored `ProfilePage.jsx` state rendering to explicitly map and display only the `Name`, `Email`, and `CreatedAt` fields.
### Outcome
Enhanced UI cleanliness and masked internal database increment keys from end users.

---

## Issue 6: Group → Subtask Hierarchy Implementation
### Problem
The original list was flat, offering no organizational capability for complex projects.
### Root Cause
The relational schema did not support self-referencing relationships.
### Solution Implemented
Implemented a strict 1-level deep Parent (Group) -> Child (Subtask) hierarchy.
### Technical Changes
Added a nullable `parent_todo_id` foreign key to the `todos` table. Abstracted the controllers so that groups (`POST /groups`) and subtasks (`POST /groups/:id/tasks`) map cleanly to this relationship.
### Outcome
Logical separation of high-level goals and actionable items.

---

## Issue 7: Group Sharing Implementation
### Problem
Application operated entirely in single-player mode; users could not collaborate.
### Root Cause
Lack of a join table defining access rights between disparate users and groups.
### Solution Implemented
Introduced a permission-based sharing architecture (VIEW/EDIT).
### Technical Changes
Created the `group_shares` table. Implemented full CRUD in the `ShareController` (`POST /groups/:id/share`, `DELETE /groups/:id/share/:userId`, etc.). Added deep permission verification inside `todo_service.go` via `GetPermission()`.
### Outcome
Multi-tenant collaboration became the core feature of the application.

---

## Issue 8: Deadlines and Progress Tracking
### Problem
Users lacked temporal awareness and completion visibility over their projects.
### Root Cause
Missing datetime tracking and aggregate mathematical calculations.
### Solution Implemented
Added deadline tracking and dynamic health-status calculation.
### Technical Changes
Added `due_date` to `todos`. Implemented `CalculateGroupHealth()` in the backend which dynamically aggregates `TotalSubtasks`, `CompletedSubtasks`, computes `Progress` percentages, and resolves a `HealthStatus` (Active, Overdue, Due Today).
### Outcome
Highly visual, actionable UI indicators that drive user urgency and organization.

---

## Issue 9: Secure Password Handling
### Problem
Authentication accepted weak passwords and stored them without rigorous compliance.
### Root Cause
Lack of regular expression evaluation and policy enforcement on registration.
### Solution Implemented
Enforced strict entropy rules and bcrypt-based cryptographic storage.
### Technical Changes
Updated `auth_service.go` to reject passwords failing a regex requiring: min 8 characters, 1 uppercase, 1 lowercase, 1 number, and 1 special character. Standardized on `golang.org/x/crypto/bcrypt`.
### Outcome
Enterprise-grade security against brute-force and dictionary attacks.

---

## Issue 10: Config Dependency Injection
### Problem
Environment variables (`os.Getenv`) were accessed globally across random packages, making the app fragile and untestable.
### Root Cause
Anti-pattern reliance on global system state.
### Solution Implemented
Abstracted configuration into a highly testable interface.
### Technical Changes
Created the `config` package providing a `Config` interface (backed by Viper). Injected down from `main.go` into services.
### Outcome
A 12-factor compliant configuration layer supporting easy test mocking.

---

## Issue 11: Custom Logger Implementation
### Problem
Standard library logging was unstructured and lacked contextual tagging.
### Root Cause
Default `log.Println` implementation.
### Solution Implemented
Migrated to Go 1.21's structured `slog` library.
### Technical Changes
Created a custom `logger` package providing a `Logger` interface. Configured JSON formatting and injected the logger across the router, middleware, and services.
### Outcome
Machine-readable, parseable logs ready for log-aggregation systems like ELK/Datadog.

---

## Issue 12: Database Dependency Injection
### Problem
Global database state (`var DB`) caused hidden dependencies and race conditions during testing.
### Root Cause
Initialization of DB objects directly into the package scope.
### Solution Implemented
Passed database connections down via constructor injection.
### Technical Changes
`repositories.NewTodoRepository(db)` now explicitly requires a database handler at initialization, completely eliminating global state.
### Outcome
Predictable initialization and the ability to inject mock database pools.

---

## Issue 13: Error Handling Refactoring
### Problem
Controllers were polluted with repetitive `c.JSON()` blocks, and generic errors risked leaking internal database schemas to clients.
### Root Cause
No central capture pipeline for application-level failures.
### Solution Implemented
Built the `apperrors` package and a unified `ErrorHandler` middleware.
### Technical Changes
Services now return `apperrors.NewNotFound()`, etc. Controllers push errors via `c.Error(err)`. The `ErrorHandler` middleware interprets these and outputs structured JSON, falling back safely to a generic 500 for unknown panics.
### Outcome
DRY controllers, strictly standardized JSON error schemas, and hardened security.

---

## Issue 14: Graceful Shutdown Support
### Problem
In-flight requests and database connections were aggressively severed upon application exit.
### Root Cause
`router.Run()` blocks synchronously with no signal interception mechanism.
### Solution Implemented
Managed OS signals to allow an orderly teardown phase.
### Technical Changes
Wired `signal.Notify` for `SIGINT`/`SIGTERM` in `main.go`. Configured an `http.Server` with a 10-second `Shutdown(ctx)` timeout. Appended `database.CloseDatabase()` to safely drain MySQL connections.
### Outcome
Zero-downtime deployments and robust protection against data corruption.

---

## Issue 15: Update Password Feature
### Problem
Users were permanently locked into their initial credentials.
### Root Cause
Missing API surface and backend logic.
### Solution Implemented
Created a secure password rotation endpoint.
### Technical Changes
Implemented `PATCH /api/auth/password`. Evaluated `current_password` via bcrypt, ran strict validation on the new password, and executed an update via `UserRepository.UpdatePassword()`.
### Outcome
Full compliance with user lifecycle management expectations.

---

## Issue 16: Reset Password Feature
### Problem
"Reset Password" natively requires integrating an SMTP dependency for email magic-links, drastically bloating scope.
### Root Cause
Feature bloat relative to core value.
### Solution Implemented
Traded a complex unauthenticated email recovery flow for a strict, authenticated "Update Password" feature (Issue 15).
### Technical Changes
Architectural decision to ensure scope remained laser-focused on Todo functionality, opting out of an unauthenticated email recovery pipeline.
### Outcome
High security maintained without relying on external 3rd party email service providers.

---

## Issue 17: Migration from GORM to SQLC
### Problem
GORM utilized heavy reflection at runtime, obscuring the actual SQL executed and hindering performance.
### Root Cause
Use of dynamic, reflection-based ORM frameworks.
### Solution Implemented
Migrated to `sqlc`, generating type-safe Go code directly from raw SQL statements.
### Technical Changes
Wrote explicit `schema.sql` and `queries.sql`. Generated `queries.go` and `row_types.go`. Completely stripped GORM dependencies.
### Outcome
100% type-safe compilation, ultra-high performance leveraging standard `database/sql`, and perfect visibility into execution plans.

---

## Issue 18: DTO Segregation
### Problem
Database entities (`models.Todo`) were passed directly to HTTP clients, leaking column shapes and extraneous fields.
### Root Cause
Tight coupling between the Storage Layer and the Presentation Layer.
### Solution Implemented
Introduced Data Transfer Objects (DTOs) for incoming requests and outgoing responses.
### Technical Changes
Created the `dto` package (`request.go`, `response.go`). Implemented `dto.MapTodo` and `dto.MapUser` to safely transform raw models into defined `TodoResponse` shapes.
### Outcome
Strict JSON API contracts that are completely immune to internal schema mutations.

---

## Issue 19: Request Object Pattern
### Problem
Service methods accepted large lists of primitive arguments (e.g., `title, desc, date, groupID`), causing signature bloat.
### Root Cause
Parameter inflation.
### Solution Implemented
Encapsulated parameters into request struct payloads.
### Technical Changes
Refactored `AuthService` and `TodoService` to accept `dto.*Request` objects natively (e.g., `CreateGroup(ctx, req dto.CreateGroupRequest, userID)`).
### Outcome
Cleaner interfaces, and adding new fields in the future requires exactly zero changes to the method signatures.

---

## Issue 20: Interface-based Architecture
### Problem
Controllers and services were hardcoded against concrete structs.
### Root Cause
Lack of idiomatic Go interfaces.
### Solution Implemented
Decoupled logic by implementing explicit interfaces.
### Technical Changes
Defined `UserRepository`, `TodoRepository`, `AuthService`, and `TodoService` interfaces.
### Outcome
Vastly improved modularity and the immediate capability to inject mock implementations during unit testing.

---

## Issue 21: 404 Middleware
### Problem
Unregistered routes returned Gin's default text `404 page not found`, breaking client JSON parsers.
### Root Cause
Missing `NoRoute` configuration.
### Solution Implemented
Intercepted unmapped routes.
### Technical Changes
Added `router.NoRoute()` in `routes.go` to return standard `{"success": false, "message": "Route not found"}` JSON payload.
### Outcome
Consistent API response structures regardless of client errors.

---

## Issue 22: 405 Middleware
### Problem
Invalid HTTP methods (e.g., sending POST to a GET route) returned default text.
### Root Cause
Gin's method interception was disabled by default.
### Solution Implemented
Enabled interception and normalized the output.
### Technical Changes
Set `router.HandleMethodNotAllowed = true` and implemented `router.NoMethod()`.
### Outcome
Polished REST compliance with predictable JSON payloads.

---

# Architecture Improvements

### Scalability Improvements
By removing the heavy ORM (GORM) in favor of `sqlc`, database execution overhead is virtually eliminated. Combined with the stateless, interface-driven `services` layer, the application can easily be scaled horizontally across multiple instances.

### Maintainability Improvements
Global state (Config, DB, Logger) has been eradicated via Dependency Injection. DTO segregation means the database schema can evolve independently from the API JSON contracts.

### Security Improvements
Implementation of rigorous regex password validation, safe error masking (preventing SQL injection leakage via generic 500s), and the strict elimination of exposed DB IDs on frontend forms.

### Performance Improvements
React `useEffect` rendering loops were severed on the client side, while the Go backend benefits from native `database/sql` execution speeds and zero-reflection querying.

### Testability Improvements
The entire application operates against interfaces. A mock `DBTX` or a mock `TodoRepository` can be passed into the services, allowing for pure business-logic testing without spinning up a live MySQL instance.

---

# API Changes

### New APIs Added

**1. Update Password**
*   **Purpose**: Allows an authenticated user to rotate their password.
*   **Request Structure**:
    ```json
    {
      "current_password": "...",
      "new_password": "..."
    }
    ```
*   **Response Structure**: `200 OK`, `{ "success": true, "message": "Password updated successfully", "data": nil }`

**2. Group Sharing Management**
*   **Purpose**: Manages access control lists for projects.
*   **Request Structure (POST /groups/:id/share)**:
    ```json
    {
      "email": "user@domain.com",
      "permission": "VIEW" 
    }
    ```
*   **Response Structure**: `201 Created`, Returns `GroupShareResponse` DTO mapping Owner and Target User structures safely.

---

# Database Changes

### New Tables
*   `group_shares`:
    *   **Relationships**: Connects `todos` (Groups) to `users` (Collaborators).
    *   **Constraints**: `UNIQUE(group_id, shared_with_user_id)` ensures no duplicate shares. Foreign keys CASCADE on delete.
    *   **Indexes**: Built index on `shared_with_user_id` for fast cross-join lookups.

### Modified Tables
*   `todos`:
    *   Added `parent_todo_id` (nullable int) to establish the Group -> Subtask hierarchy.
    *   Added `due_date` (nullable timestamp) for deadline tracking.

---

# UI Changes

*   **Dashboard Changes**: Implemented conditional rendering to display visual Progress Bars, "Days Remaining" chips, and "Overdue" red-alert indicators. Reduced API network spam.
*   **Group Details Changes**: Renders nested Subtasks effectively. Added an Edit modal for metadata adjustments.
*   **Sharing UI Changes**: Built a dynamic `select` element allowing Owners to toggle Collaborators between `VIEW` and `EDIT` modes instantly.
*   **Profile UI Changes**: Purged exposed internal `ID` fields, elevating the visual hierarchy of the User's name and joining date.

---

# Final Summary

| Issue | Status | Improvement |
| ----- | ------ | ----------- |
| Multiple API calls on Dashboard | Resolved | Performance & stability |
| Current Date Validation Bug | Resolved | UX & Data integrity |
| Missing Edit Group Feature | Resolved | UX capability |
| Missing Collaborator Role Update | Resolved | Access control capability |
| Profile Page displaying User ID | Resolved | Security & UX |
| Group → Subtask Hierarchy | Resolved | Data architecture |
| Group Sharing Implementation | Resolved | Multi-tenant logic |
| Deadlines and Progress Tracking | Resolved | UX & Feature set |
| Secure Password Handling | Resolved | Security |
| Config Dependency Injection | Resolved | Maintainability & Testing |
| Custom Logger Implementation | Resolved | Observability |
| Database Dependency Injection | Resolved | Maintainability & Testing |
| Error Handling Refactoring | Resolved | Security & Code DRYness |
| Graceful Shutdown Support | Resolved | Infrastructure stability |
| Update Password Feature | Resolved | Feature set |
| Reset Password Feature | Re-scoped | Focused architecture |
| Migration from GORM to SQLC | Resolved | Performance & Type safety |
| DTO Segregation | Resolved | Architecture boundary |
| Request Object Pattern | Resolved | Signature cleanliness |
| Interface-based Architecture | Resolved | Testability |
| 404 Middleware | Resolved | API consistency |
| 405 Middleware | Resolved | API consistency |
