# Ticket App

A modern, scalable ticketing system built with Go, following Clean Architecture principles for maintainability, testability, and extensibility.

---

## 📝 Assignment: Event Ticket Booking System

**Objective**  
Build a RESTful API that allows users to book tickets for events. The system must handle concurrent bookings for a limited number of tickets — ensuring data consistency, safe transactions, and proper race condition handling.

---

### 🎯 Functional Requirements

#### 1. Events
Support creating, updating, and deleting events.  
Each event should include:
- `id`, `name`, `description`, `date_time`
- `total_tickets`: total number of tickets available for sale
- `ticket_price`: price per ticket

#### 2. Users
Each user has:
- `id`, `name`, `email`

> Authentication can be simplified or mocked. Users can be pre-registered or created as part of test data.

#### 3. Ticket Booking
- Users can book a specified number of tickets for an event.
- The system must ensure:
  - No overbooking beyond `total_tickets`
  - No overselling, even under high concurrency
- Booking `status` should be one of:
  - `PENDING`, `CONFIRMED`, `CANCELLED`

#### 4. Payment (Simulated)
- After successful booking, simulate a payment process
- Use an async job system (e.g., **RabbitMQ**, **Redis Queue**, or **cron + DB polling**) for payment handling
- If payment is not completed within **15 minutes**, automatically:
  - Cancel the booking (`CANCELLED`)
  - Release the reserved tickets

#### 5. Statistics
Provide statistics per event:
- Total tickets sold
- Estimated revenue (from `CONFIRMED` bookings)

---

### ⚙️ Technical Requirements

- Use **database transactions** to safely handle concurrent bookings.
- Implement **row-level locking** or **atomic updates** to prevent overselling.
  - > Explain the approach used clearly in documentation or code comments.
- Optimize **indexes** for booking-related queries.
- APIs must:
  - Follow **standard HTTP** practices
  - Include **input validation**
- Write **at least 3 unit tests** for core booking logic.

---

### 📬 Submission

- Upload to **GitHub**/**GitLab** and share the repository link.
- Ensure the application **runs out of the box** using instructions in `README.md`.

---

## Table of Contents

- [Ticket App](#ticket-app)
  - [📝 Assignment: Event Ticket Booking System](#-assignment-event-ticket-booking-system)
    - [🎯 Functional Requirements](#-functional-requirements)
      - [1. Events](#1-events)
      - [2. Users](#2-users)
      - [3. Ticket Booking](#3-ticket-booking)
      - [4. Payment (Simulated)](#4-payment-simulated)
      - [5. Statistics](#5-statistics)
    - [⚙️ Technical Requirements](#️-technical-requirements)
    - [📬 Submission](#-submission)
  - [Table of Contents](#table-of-contents)
  - [Overview](#overview)
  - [Features](#features)
  - [Architecture](#architecture)
  - [Implementation Details](#implementation-details)
    - [Booking Logic \& Concurrency Handling](#booking-logic--concurrency-handling)
    - [Payment Simulation \& Asynchronous Processing](#payment-simulation--asynchronous-processing)
    - [Booking Status Lifecycle](#booking-status-lifecycle)
    - [Event Statistics](#event-statistics)
    - [Indexes \& Performance](#indexes--performance)
    - [Unit Testing](#unit-testing)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)

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

## Implementation Details

### Booking Logic & Concurrency Handling
- All ticket bookings are processed within a database transaction to ensure atomicity and consistency.
- Row-level locking (e.g., `SELECT ... FOR UPDATE`) or atomic updates are used to prevent overselling tickets under high concurrency. When a user books tickets, the system locks the event row, checks available tickets, and decrements the count only if enough tickets remain.
- This approach guarantees that no more tickets are sold than the event's `total_tickets`, even with simultaneous booking requests.

### Payment Simulation & Asynchronous Processing
- After a booking is created (status `PENDING`), a simulated payment process is triggered asynchronously using a job queue (e.g., Redis Queue, RabbitMQ, or cron-based DB polling).
- If payment is completed successfully, the booking status is updated to `CONFIRMED`.
- If payment is not completed within 15 minutes, a background worker automatically cancels the booking (`CANCELLED`) and releases the reserved tickets back to the pool.

### Booking Status Lifecycle
- Bookings transition through the following statuses:
  - `PENDING`: Booking created, awaiting payment.
  - `CONFIRMED`: Payment successful, tickets are officially reserved.
  - `CANCELLED`: Payment failed or timed out, tickets are released.

### Event Statistics
- For each event, the system provides:
  - **Total tickets sold**: Sum of tickets from all `CONFIRMED` bookings.
  - **Estimated revenue**: Calculated from the ticket price and number of `CONFIRMED` bookings.

### Indexes & Performance
- Database indexes are created on booking and event tables to optimize queries related to ticket availability, user bookings, and event statistics.
- This ensures efficient lookups and reporting, even as data volume grows.

### Unit Testing
- At least three unit tests are implemented for the core booking logic, covering scenarios such as successful booking, overbooking prevention, and booking cancellation due to payment timeout.

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
