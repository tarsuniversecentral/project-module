```markdown
# Project Name

A Go (Golang) web application that allows users to create projects, add team members, upload multiple PDFs and images, and retrieve project details. This application uses Gorilla Mux for routing and MySQL for database management.


## Features

- **Project Management:** Create new projects and update project details.
- **Team Management:** Add team members, update team member roles, and retrieve team member lists.
- **File Uploads:** Upload and manage multiple PDFs and images for each project.
- **RESTful API:** Built using Gorilla Mux to handle routing.
- **Database:** Utilizes MySQL with migration support.

## Project Structure



    ├── cmd                      # Entry point for the application
    │   └── main.go
    ├── config                   # Application configuration settings
    │   └── config.go
    ├── go.mod
    ├── go.sum
    ├── images                   # Stores uploaded image files 
    │   ├── uuid.png
    │   ├── uuid.png
    ├── internal                 # Application logic 
    │   ├── api                  # API endpoints initialization 
    │   │   └── api.go
    │   ├── dto                  # Data objects 
    │   │   ├── file.go
    │   │   └── project.go
    │   ├── handlers             # HTTP handlers 
    │   │   └── project.go
    │   ├── models               # Database Layer
    │   │   └── project.go
    │   ├── router               # Route definitions using Gorilla Mux
    │   │   └── router.go
    │   └── services             # Business logic of handlers
    │       ├── file.go
    │       └── project.go
    ├── pdfs                      # Stores uploaded PDF files 
    │   ├── uuid.pdf
    │   ├── uuid.pdf
    └── pkg
        ├── database        
        │   ├── database.go        # Database connection initialization
        │   └── migration
        │       ├── migration.go   # Migration logic  
        │       └── migrations
        │           ├── 0001_create_table_project_up.sql
        │           ├── 0002_create_table_team_members_up.sql
        │           ├── 0003_create_table_project_pitch_decks_up.sql
        │           ├── 0004_create_table_project_images_up.sql
        │           └── 0005_add_counts_and_verified.sql
        └── utils
            ├── errors.go
            └── file.go

```

1. **Clone the repository:**

   ```bash
   git clone https://github.com/tarsuniversecentral/project-module.git
   cd project-module
   ```

2. **Install dependencies:**

   Ensure you have [Go](https://golang.org/dl/) installed (version 1.16 or later recommended).

   ```bash
   go mod download
   ```

3. **Set up the MySQL Database:**

   - Create a new MySQL database.
   - Update the database connection settings in the configuration file (`config/config.go`) or via environment variables as required by your application.


## Configuration

Update the configuration settings (such as database credentials, server port, etc.) in the corresponding environment variables. Example configuration variables might include:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `SERVER_PORT`

## Running the Project

After setting up the configuration and database, you can run the project using:

```bash
go run cmd/main.go
```

This will start the server, and you should see output indicating that the server is running on the specified port.

## Usage

- **API Endpoints:**  
  The application exposes RESTful endpoints for creating projects, managing team members, and uploading files.  
  Refer to the API documentation (if available) or review the handlers in `internal/handlers` for endpoint details.

- **Uploading Files:**  
  The `images` and `pdfs` directories are used to store uploaded image and PDF documents respectively. Ensure that these directories have the appropriate write permissions.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
```