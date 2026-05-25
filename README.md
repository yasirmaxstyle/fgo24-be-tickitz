<div align="center">
  <h1>🍿 Noir Backend API (Movie Ticketing)</h1>
  <p>A modern backend REST API for a movie ticketing application written in Go, adhering to the SOLID principles and clean Architecture. </p>

  [![Go Version](https://img.shields.io/github/go-mod/go-version/yasirmaxstyle/fgo24-be-tickitz)](https://github.com/yasirmaxstyle/fgo24-be-tickitz)
  [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

## Overview

The API provides core functionalities to manage theaters, users, movies, genres, cast assignments, transaction tracking and JWT Role-Based authentication. It is built natively on the powerful **Gin Web Framework** leveraging **PostgreSQL** through **pgx**.

### Key Features
- **MVC & Repository Pattern**: Clean separation of concerns routing through Controllers, Services, and DB-agnostic Repositories.
- **Dynamic File Processing**: Natively process multipart form data such as Movie posters and Backdrops uploads.
- **Relational Integrity**: Complete constraints mapping relationships for complex `many-to-many` tables (e.g. Movies to Genres, Movies to Cast).
- **JWT & Role Middleware**: Built in HTTP guards for authentication routines & protecting Administrative boundaries.
- **Redis Integration**: High speed caching support ready.
- **Swagger Documentation**: Live OpenAPI tracking for the endpoints.

---

## Tech Stack

- **Go 1.20+**
- **Gin-gonic/gin** - HTTP Web Framework
- **JackC/pgx** - Pure Go PostgreSQL driver and toolkit
- **Redis** - Key-Value store primarily for in-memory caching
- **Docker Compose** - Immediate local provisioning

---

## Getting Started

Follow these steps to run the application locally on your machine.

### 1. Requirements
Ensure you have the following installed:
- [Go](https://go.dev/doc/install) (latest)
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose

### 2. Clone the repository
```sh
git clone https://github.com/yasirmaxstyle/fgo24-be-tickitz.git
cd fgo24-be-tickitz
```

### 3. Setup Configuration
Copy the provided `.env.example` mapping to establish your core environment variables.
```bash
cp .env.example .env
```
Ensure you fill out the DB and Redis credentials in the `.env` file prior to launching the containers.

### 4. Running With Docker (Recommended)
You can instantaneously build and attach the backend network along with a live PostgreSQL and Redis server by running:
```sh
docker compose up --build
```
> **Note**: This will automatically spin up PostgreSQL on port `5432`, Redis on `6379`, and the API on your configured port.

### 5. Running Manually (Hot Reloading)

If you strictly want to run the Go application natively while retaining Hot-Reload:

**Install `gow`**:
```sh
go install github.com/mitranim/gow@latest
export GOPATH="$HOME/go"
export PATH="$GOPATH/bin:$PATH"
```
**Start the local Postgres & Redis services**:
```sh
docker compose up -d postgres redis
```
**Initialize standard go binaries**:
```sh
go mod download
gow run main.go
```

---

## Key API Endpoints 

| Route | Method | Access | Description |
| --- | --- | --- | --- |
| `/auth/register` | `POST` | Public | Register standard application user |
| `/auth/login` | `POST` | Public | Returns secure JWT payload |
| `/movie/:id` | `GET` | Public | Get aggregate movie data |
| `/admin/movie` | `POST` | Admin | Upload a movie entry with its relationships |
| `/admin/movie/:id` | `PATCH` | Admin | Update existing movie columns/images |
| `/admin/movie/:id` | `DELETE`| Admin | Delete a movie and cascade its genres & cast |

*For full API interaction payloads, view the Swagger route upon initializing the server via `/swagger/index.html`.*

---

## Contributing

Contributions, issues, and feature requests are always welcome! Feel free to check the [issues page](https://github.com/yasirmaxstyle/fgo24-be-tickitz/issues) to take part in this project.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## Contact
- **Instagram:** [@yasirmaxstyle](https://www.instagram.com/yasirmaxstyle/)
- **LinkedIn:** [Muhamad Yasir](https://www.linkedin.com/in/muhamad-yasir-806230117/)

## License
This project is [MIT licensed](https://opensource.org/licenses/MIT).