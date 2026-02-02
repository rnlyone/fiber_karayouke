# Fiberavel

Fiberavel is a Go web application skeleton powered by Fiber and structured with Laravel-inspired conventions (MVC, service layers, artisan-style tooling). It is fully API focused—no Blade-like views ship by default—so you can grow REST services quickly while keeping the directory layout familiar to PHP artisans.

## Features

- MVC pattern: Organize your code into Models, Views, and Controllers for a structured and maintainable application.
- Go Fiber: Utilize the fast and efficient Fiber web framework for building high-performance web applications.
- Service Layer: Implement a Service layer to encapsulate business logic and keep your controllers lean.
- Flexible Routing: Define routes and handle HTTP requests easily with the powerful routing capabilities of Go Fiber.
- Database Integration: Connect to your preferred database system using GORM, an excellent ORM library for Go.

## Project Structure

The project follows a folder structure similar to Laravel, providing a clear separation of concerns and promoting modular development. Here's a brief overview:

- `app/`: Contains the application-specific code, including models, views, controllers, and services.
- `config/`: Stores configuration files for the application, such as database settings, middleware configurations, etc.
- `database/`:Not Available now but can be customized.
- `public/`: Contains publicly accessible assets such as CSS, JavaScript, and static files.
- `routes/`: Defines the application routes and maps them to the appropriate controllers and actions.
- `main.go`: The entry point of the application where the server starts and routes are registered.

## Getting Started

To get started with the Go Fiber MVC App:

1. Clone the repository.
2. Install the necessary dependencies using `go get` or any package manager you prefer.
3. Configure the application settings in the `config/` folder, such as database connection details.
4. Run database migrations to set up the required database schema.
5. Start the application using `go run main.go` or by building and running the binary.
6. Access the application in your web browser at `http://localhost:3000` or the configured port.

### Environment variables

Sensitive settings such as database credentials now live in a `.env` file that is loaded through `github.com/joho/godotenv`.

1. Copy `.env.example` to `.env`.
2. Set `IS_PRODUCTION=true` when deploying; keeping it `false` allows blank `DB_PASSWORD`/`OAUTH_DB_PASSWORD` values for local setups.
3. Fill in the values for both the primary application database (`DB_*`) and the OAuth database (`OAUTH_DB_*`).
4. Keep `.env` out of version control (it is already ignored via `.gitignore`).

### Artisan-style helpers

Fiberavel ships with a lightweight artisan CLI so your terminal muscle-memory still works:

- `go run . artisan migrate` — run GORM `AutoMigrate` across the registered models list.
- `go run . artisan make model User` — scaffold a new model file (e.g., `app/models/user_model.go`).
- `go run . artisan make controller User` — create `app/controllers/user_controller.go` with starter handler methods.
- `go run . artisan make repository Client` — create `app/repositories/client_repository.go` with constructor stub.

You can also use the colon form (`go run . artisan make:model User`) just like in Laravel. Feel free to extend the `app/artisan` package with additional commands (seeders, jobs, etc.) as your project grows.

### Release artifacts

The `release/` directory is intentionally ignored in this repository so that production bundles can live in a separate repo. To fetch or update the external release repo, run:

```
bash scripts/setup_release_repo.sh
```

By default this clones `https://github.com/rnlyone/karayouke_release.git` into `release/`. Pass a custom path/URL if you need a different mirror.

### Credits

Fiberavel stands on the shoulders of the original [GoFiberMVC](https://github.com/samrat415/GoFiberMVC) project. Massive thanks to the maintainers for the initial inspiration.

Feel free to explore the code, modify it according to your needs, and build upon this foundation to create powerful web applications using Go Fiber.

