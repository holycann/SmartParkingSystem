# Smart Parking System

A robust, fault-tolerant, and highly concurrent real-time smart parking system built with Golang, React, Tailwind CSS, and PostgreSQL.

## System Architecture

This smart parking system is designed to handle real-time tracking and management of parking spaces across multiple lots and garages in a busy downtown area. The architecture follows a microservices approach with the following components:

- **Backend (Golang)**: Handles business logic, API endpoints, WebSocket connections, and database interactions
- **Frontend (React + Tailwind CSS)**: Provides a responsive and intuitive user interface for viewing and reserving parking spots, utilizing Tailwind CSS for modern, utility-first styling
- **Database (PostgreSQL)**: Stores information about parking lots, spaces, users, and reservations
- **WebSockets**: Enables real-time communication between the server and clients

## Key Features

1. Real-time tracking and management of parking space availability
2. User-friendly interface for viewing and reserving parking spots
3. Automatic updates when vehicles enter or exit parking spaces
4. Notifications about available spaces, time limits, and parking fees
5. Fault tolerance and high availability through concurrency control mechanisms

## Project Structure

```
SmartParkingSystem/
├── backend/             # Golang backend
│   ├── api/             # API handlers and routes
│   ├── models/          # Data models
│   ├── services/        # Business logic
│   ├── utils/           # Utility functions
│   ├── middleware/      # Middleware components
│   ├── config/          # Configuration files
│   └── tests/           # Unit and integration tests
├── frontend/            # React frontend
│   ├── public/          # Static files
│   └── src/             # Source code
│       ├── components/  # Reusable UI components
│       ├── pages/       # Page components
│       ├── hooks/       # Custom React hooks
│       ├── services/    # API services
│       ├── utils/       # Utility functions
│       ├── assets/      # Images, fonts, etc.
│       ├── context/     # React context providers
│       └── styles/      # Tailwind configuration and global styles
├── database/            # Database migration scripts
└── docs/                # Documentation
```

## Getting Started

### Prerequisites

- Go (version 1.21+)
- Node.js (version 18+)
- PostgreSQL (version 14+)

### Installation

1. Clone the repository
2. Set up the backend:
   ```bash
   cd SmartParkingSystem/backend
   go mod init github.com/holycann/smartparkingsystem
   go mod tidy
   ```
3. Set up the frontend:
   ```bash
   cd SmartParkingSystem/frontend
   npm install
   ```
4. Set up the database:
   ```bash
   # Create PostgreSQL database
   createdb smart_parking_system
   ```

### Running the Application

1. Start the backend:
   ```bash
   cd SmartParkingSystem/backend
   go run main.go
   ```
2. Start the frontend:
   ```bash
   cd SmartParkingSystem/frontend
   npm start
   ```

## Concurrency and Fault Tolerance

This system implements several concurrency patterns and fault tolerance mechanisms:

- **Buffers**: For managing parking space availability updates
- **Synchronization**: To handle concurrent reservation requests
- **Semaphores**: To control access to shared resources
- **Fault Tolerance**: Through retry mechanisms, circuit breakers, and graceful degradation

## License

This project is licensed under the MIT License - see the LICENSE file for details.
