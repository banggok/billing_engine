## Technologies Used
- **Golang** (1.23)
- **Gin** (Web framework)
- **GORM** (ORM for PostgreSQL)
- **PostgreSQL** (Database)
- **Ginkgo** and **Gomega** (Testing frameworks)
- **Testify** (Mocking framework)
- **Sonarqube** (Code Quality Analysis)
- **Soda CLI** (Database Migration)

## Prerequisites
Before you start, ensure that you have the following installed:
- **Golang** (1.23 or above)
- **PostgreSQL** (version 12 or higher)
- **Git**
- **Sonarqube**
- **Soda CLI**

## Setup Instructions

### Clone the Repository
First, clone the repository to your local machine:
```bash
https://github.com/banggok/billing_engine.git
cd billing_engine
```

### Run Migrations
To set up the database schema, run the SQL migration file:

```bash
soda migrate
```

This will create the necessary table in your PostgreSQL database.

## Running the Application

### Running in Development
To run the application in development mode, execute the following command:

```bash
make run
```

The server will start on `http://localhost:8080`.

## Running Tests

### End-to-End Tests
This project uses Ginkgo for end-to-end testing. To run the tests:

```bash
make test
```

This will run all tests across your project, providing verbose output.

### Running with Coverage
To run tests with coverage and generate a coverage report, use:

```bash
make coverage
```

You can open `coverage.html` in a browser to see detailed test coverage results.

### Running Sonar Scanner
To run tests with coverage and generate a coverage report, use:

```bash
make sonar
```

## Project Structure

```bash
/project-root
│
├── /api
│   ├── /handler        # Contains API handlers for processing HTTP requests
│   ├── /middleware     # Contains middleware to handle request
│   ├── /routes         # Defines the routes for the application
│
├── /cmd
│   ├── /api            # Application entry point, main.go for starting the server
│
├── /internal
│   ├── /entity         # Domain entities (Customer, Loan, Payment)
│   ├── /model          # GORM models for database interaction
│   ├── /repository     # Database interaction logic (CRUD operations)
│   ├── /usecase        # Business logic related to handling loans, payments, etc.
│
├── /pkg
│   ├── db.go           # Database connection logic
│
├── /tests              
│   ├── /e2e            # Contains end-to-end test scenarios
│
├── .env                # Environment variables for the application
├── .env.test           # Environment variables specific to the testing environment
├── go.mod              # Go module file
├── go.sum              # Go module dependencies
└── README.md           # Project documentation
```


## Contribution Guidelines

If you'd like to contribute, feel free to fork the repository and submit a pull request. Make sure to:
- Write tests for new features and bug fixes.
- Follow Go conventions and run `go fmt` to format your code.
- Include detailed commit messages.
