# Internal DNS Server Project

This project provides an internal DNS server with a web interface to manage domain-to-IP mappings. It's built with a Go backend, Redis for caching, PostgreSQL for persistence, and a React SPA for the frontend.

## Features

-   **Internal DNS Resolution**: Resolves internal service domains (CNAME and A records).
-   **Web Interface**: A Single Page Application (SPA) for CRUD operations on DNS records.
-   **Authentication & Authorization**: JWT-based authentication with Role-Based Access Control (`user` and `admin` roles).
-   **Performance**: Scalable to handle 10k-100k records, with Redis caching and a Bloom filter for duplicate prevention.
-   **Observability**: Prometheus metrics, health checks, and audit trails.
-   **Security**: Rate limiting, password hashing, and input sanitization.
-   **Containerized**: Fully containerized with Docker and orchestrated with Docker Compose.

## Architecture

-   **Backend**: Go (using Echo framework)
-   **Database**: PostgreSQL
-   **Cache**: Redis (for DNS records and a Bloom filter)
-   **Frontend**: React SPA with Tailwind CSS
-   **DNS Server**: Custom Go server using `miekg/dns`
-   **API Documentation**: Swagger/OpenAPI

## Getting Started

### Prerequisites

-   Docker and Docker Compose
-   Go (1.21+)
-   Node.js and npm (for frontend development)
-   `swag` CLI (`go install github.com/swaggo/swag/cmd/swag@latest`)

### Setup & Running the Application

1.  **Clone the repository:**
    ```sh
    git clone <repository-url>
    cd internal-dns-server
    ```

2.  **Set up environment variables:**
    Copy the example environment file and customize it if needed. The default values are configured to work with the provided Docker Compose setup.
    ```sh
    cp .env.example .env
    ```

3.  **Build and run with Docker Compose:**
    This is the recommended way to run the entire stack (API, DNS server, PostgreSQL, Redis).
    ```sh
    docker-compose up --build
    ```

    -   **API Server**: Available at `http://localhost:8080`
    -   **DNS Server**: Listening on UDP port `5353`
    -   **Swagger Docs**: Available at `http://localhost:8080/swagger/index.html`
    -   **Prometheus Metrics**: Available at `http://localhost:8080/metrics`

### Development

#### Backend

-   **Run tests:**
    ```sh
    make test
    ```
-   **Run benchmarks:**
    ```sh
    make bench
    ```
-   **Generate Swagger docs:**
    After adding/updating `swag` annotations in the code:
    ```sh
    make swag
    ```

#### Frontend

Navigate to the `frontend` directory for frontend-specific commands.

-   **Install dependencies:**
    ```sh
    cd frontend
    npm install
    ```
-   **Run development server:**
    ```sh
    npm start
    ```

## API Endpoints

The API is versioned under `/api/v1`. See the Swagger documentation at `http://localhost:8080/swagger/index.html` for a detailed list of endpoints.

-   `/health`: Health check
-   `/auth/register`: Register a new user
-   `/auth/login`: Log in and receive JWT
-   `/dns-records`: CRUD operations for user's DNS records (requires auth)
-   `/admin/users`: User management (admin only)

## Project Structure

The project follows Clean Architecture principles.

-   `/cmd`: Main entry points for the `api` and `dns` servers.
-   `/internal`: Contains the core application logic.
    -   `/domain`: Core entities and business rules.
    -   `/usecase`: Application-specific business logic interfaces.
    -   `/service`: Implementation of use cases.
    -   `/repository`: Data access layer interfaces.
    -   `/infrastructure`: Implementation of external concerns (database, cache, transport).
-   `/pkg`: Reusable packages (e.g., `bloomfilter`, `token`).
-   `/frontend`: React SPA source code.
-   `/docs`: Generated Swagger documentation.

