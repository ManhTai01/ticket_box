# Ticket App

A modern, scalable ticketing system built with Go, following Clean Architecture principles for maintainability, testability, and extensibility.

---

## Table of Contents

- [Ticket App](#ticket-app)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Architecture](#architecture)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
  - [Configuration](#configuration)
  - [Running the Application](#running-the-application)
    - [Using Docker Compose](#using-docker-compose)
    - [Local Development](#local-development)
  - [Testing](#testing)
  - [Project Structure](#project-structure)
  - [Contributing](#contributing)
  - [License](#license)

---

## Overview

**Ticket App** is a professional event ticketing platform designed to manage events, bookings, and payments. The application is built in Go, leveraging Clean Architecture to ensure a clear separation of concerns, ease of testing, and adaptability to future requirements.

---

## Features

- User authentication and authorization
- Event creation, listing, and management
- Ticket booking and order management
- Payment processing and status tracking
- Health checks and monitoring endpoints
- RESTful API with robust validation and error handling
- Dockerized for easy deployment

---

## Architecture

Ticket App implements the Clean Architecture pattern, which emphasizes:

- Independence from frameworks and external agencies
- Testability of business logic
- Decoupling of UI, database, and core logic

**Layers:**

- **Domain:** Business models and interfaces
- **Repository:** Data access logic
- **Service/Usecase:** Business rules and application logic
- **Delivery:** HTTP handlers and middleware

![Clean Architecture Diagram](clean-arch.png)

---

## Getting Started

### Prerequisites

- Go 1.20+
- Docker & Docker Compose
- (Optional) Make

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/ticket_app.git
   cd ticket_app
   ```

2. **Copy environment variables:**
   ```bash
   cp example.env .env
   ```

3. **(Optional) Run database migrations:**
   ```bash
   # Example for PostgreSQL
   docker-compose up -d db
   # Run your migration tool here
   ```

---

## Configuration

All configuration is managed via environment variables. See `example.env` for reference.

---

## Running the Application

### Using Docker Compose

```bash
make up
```

The application will be available at `http://localhost:9090`.

### Local Development

```bash
go mod tidy
make run
```

---

## Testing

Run all tests with:

```bash
make tests
```

---

## Project Structure

```
app/                # Application entrypoint
auth/               # Authentication logic
booking/            # Booking domain logic
domain/             # Core business models and interfaces
event/              # Event domain logic
health/             # Health check services
internal/           # Internal packages (db, repository, rest, etc.)
migrations/         # Database migration scripts
payment/            # Payment domain logic
```

---

## Contributing

Contributions are welcome! Please open issues or submit pull requests for any improvements or bug fixes.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

If you have any questions or need support, please open an issue in the repository.

---

Let me know if you want to add a section for API documentation, deployment to cloud, or anything else!
