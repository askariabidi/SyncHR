# SyncHR

A distributed, web-based Human Resource Management System built with Go and React, developed as a final project for *Distributed Programming for Web, IoT, and Mobile Systems*.

SyncHR lets employees check in/out, apply for and track leave, and view payslips, while HR managers review and approve leave requests, browse attendance day-by-day, manage the employee directory, and push instant notifications to the whole workforce.

## Tech Stack

**Backend**
- Go, [gorilla/mux](https://github.com/gorilla/mux) for REST routing
- [gorilla/websocket](https://github.com/gorilla/websocket) for real-time push
- PostgreSQL via [lib/pq](https://github.com/lib/pq)
- JWT authentication via [golang-jwt/jwt](https://github.com/golang-jwt/jwt)

**Frontend**
- React 19 + React Router
- Axios for REST calls, the native `WebSocket` API for real-time updates

## Distributed Systems Concepts

- **Client-server architecture** — a stateless Go REST API consumed by an independent React SPA.
- **Real-time synchronization over WebSocket** — a single hub goroutine (`backend/middleware.go`) owns the registry of connected clients and coordinates registration, disconnection, and message dispatch purely through channels (`register`, `unregister`, `direct`) rather than shared-memory locking — the classic Go "actor" pattern. Every connected browser tab gets notifications pushed to it the instant they happen (leave submitted/approved/rejected, HR broadcast announcements), with no polling.
- **JWT-based stateless authentication with RBAC** — every protected request carries a signed JWT; role (`employee` / `hr_manager`) is checked per-endpoint on the backend and per-route on the frontend.
- **Asynchronous, concurrent request handling** — each WebSocket connection is served by its own goroutine; notification delivery (DB persistence + live push) happens without blocking the HTTP request that triggered it.

## Features

- **Auth** — register/login, JWT sessions, role-based routing (employee vs HR).
- **Attendance** — check in/out with a live session timer; employees browse their own attendance by date; HR browses everyone's attendance by date, with month navigation and day tabs.
- **Leave management** — employees apply for leave and track status; HR reviews, approves, or rejects requests with notes; leave balances are deducted automatically on approval.
- **Payslips** — employees view monthly payslips with a full earnings/deductions breakdown.
- **Employee directory** — HR can search and browse all employees.
- **Real-time notifications** — leave apply/approve/reject and HR-authored announcements are persisted to the database and pushed live over WebSocket to anyone online; a bell icon with unread count is available on both dashboards.

## Project Structure

```
backend/    Go REST API + WebSocket server
  main.go              route registration, server startup
  middleware.go        JWT auth middleware + WebSocket hub
  auth.go, attendance.go, leave.go, payslip.go, holiday.go, notification.go
                        one handler file per domain
  models.go             shared request/response/entity structs
  database.go           Postgres connection + notifications table bootstrap

frontend/   React SPA (Vite)
  src/pages/             one component per screen (Login, dashboards, leave, payslip)
  src/context/           AuthContext, NotificationContext (app-wide WebSocket connection)
  src/components/        shared components (ProtectedRoute, NotificationBell)
  src/hooks/             useDateNavigator (date-picker/month-nav/day-tabs logic)
  src/utils/             date/time formatting helpers
  src/services/api.js    Axios client + all REST calls

db/schema.sql   full Postgres schema (tables, indexes)
```

## Getting Started

### Prerequisites

- Go (see `backend/go.mod` for the toolchain version)
- Node.js 18+
- PostgreSQL running locally (or reachable) with a database created for this project

### 1. Database

Apply the schema to your Postgres database:

```bash
psql -U <your_user> -d <your_db> -f db/schema.sql
```

### 2. Backend

```bash
cd backend
cp .env.example .env   # then fill in your real DB credentials and a JWT secret
go run .
```

The API starts on `http://localhost:8080` (override with `PORT` in `.env`).

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

The app starts on `http://localhost:5173` and talks to the backend at the hardcoded base URL in `src/services/api.js`.

### Creating Accounts

`db/schema.sql` defines the tables only - it doesn't seed any users. Register an account of each role via the API:

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"employee@example.com","password":"password123","first_name":"John","last_name":"Doe","role":"employee","department":"Engineering"}'

curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"hr@example.com","password":"password123","first_name":"Jane","last_name":"Smith","role":"hr_manager","department":"HR"}'
```

`role` must be either `employee` or `hr_manager`.

## API Overview

All routes are prefixed with `/api`. Protected routes require `Authorization: Bearer <token>`.

| Area | Endpoints |
|---|---|
| Auth | `POST /auth/login`, `POST /auth/register`, `GET/PUT /users/profile`, `GET /users/employees` (HR) |
| Attendance | `POST /attendance/checkin`, `POST /attendance/checkout`, `GET /attendance/history`, `GET /attendance/records?date=` (HR) |
| Leave | `POST /leave/apply`, `GET /leave/balance`, `GET /leave/requests`, `PUT /leave/approve/{id}`, `PUT /leave/reject/{id}` (HR) |
| Payslip | `GET /payslip/list`, `GET /payslip/{id}` |
| Holidays | `GET /holidays` |
| Notifications | `GET /notifications`, `PUT /notifications/{id}/read`, `PUT /notifications/read-all`, `POST /notifications/broadcast` (HR) |
| Real-time | `GET /ws/notifications?token=` — WebSocket upgrade (public route; the JWT travels as a query param since browsers can't set custom headers on the handshake) |
