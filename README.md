# Kedubak

## Description
This project is an API built in Go using the Fiber framework for creating RESTful endpoints, and it utilizes MongoDB Atlas as the database.  
The API is designed to be used with a front-end application, which can be find in a docker container.

## Getting Started

### Prerequisites
- Docker installed on your machine
- MongoDB Atlas database with User and Post collection 

### Installation
1. Clone this repository.
2. Copy the `env.sample` file and rename it to `.env`. Fill in the required environment variables as specified in the file.
3. Refer to the `subject.md` file for detailed information about the project requirements and specifications.

### Running the Application
commands written in the run_app.md file.
#### With Docker Compose (Front-end + Back-end)
```bash
docker compose up --build
```
#### only the back-end
```bash
docker build -t kedubak .
docker run -p 8080:8080 --env-file .env kedubak
```
