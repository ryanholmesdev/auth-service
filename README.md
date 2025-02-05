# Auth Service

## Overview

The Auth Service is a microservice for managing OAuth tokens and authentication for various music providers (e.g., Spotify, Tidal). It supports token storage in Redis and provides endpoints for login, callback handling, and token retrieval.

## Features
* 	OAuth login and callback handling for music providers.
* 	Token storage using Redis.
* 	Session management via cookies.

## User Flow

The flow for authentication follows these steps:
1.	Login (GetAuthProviderLogin): The user is redirected to the OAuth provider’s login page. 
A redirect_uri is provided to guide the user back to the front-end after successful authentication.


2.	Callback (GetAuthProviderCallback): The OAuth provider redirects the user back to your back-end with an authorization code.
	The back-end exchanges the code for an access token and stores it in Redis. A session ID is set in a cookie, and the user is redirected back to the front-end.


3. Retrieve Token (GetAuthProviderToken):The front-end can call this endpoint to retrieve the access token for the user’s session.

## Prerequisites

To set up the development environment, you’ll need:
1. **Go** (version 1.20 or higher):
2. **Docker**: Install Docker
3. **Redis**: Installed locally or via Docker.

## Setting Up the Environment

### 1. Clone the Repository

`git clone https://github.com/your-username/auth-service.git`

`cd auth-service`

## 2. Install Dependencies:

`go mod tidy`

## 3. Set Up Redis

You can run Redis using Docker:

`docker run --name redis -p 6379:6379 -d redis:7`

Verify Redis is Running:

`docker ps`

`docker logs redis`

## 4. Configure Environment Variables

### **Environment Configuration**
The service uses `.env` files for environment-specific configuration. By default, it dynamically loads `.env.<APP_ENV>` based on the `APP_ENV` environment variable.

#### **Default Behavior**
- If `APP_ENV` is not set, it defaults to `local` and loads `.env.local`.

#### **Environment Variables**
Below is the list of required environment variables:

| Variable              | Description                               | Example Value                   |
|-----------------------|-------------------------------------------|----------------------------------|
| `SPOTIFY_CLIENT_ID`   | Spotify client ID                        | `your-spotify-client-id`        |
| `SPOTIFY_CLIENT_SECRET` | Spotify client secret                  | `your-spotify-client-secret`    |
| `SPOTIFY_REDIRECT_URL` | Spotify OAuth redirect URL              | `http://localhost:8080/auth/spotify/callback` |
| `REDIS_ADDR`          | Redis server address                     | `localhost:6379`                |

#### **Environment File Structure**
You can create the following `.env` files for different environments:

1. **`.env.local`** - Used for running the application locally on your machine.
    ```dotenv
    SPOTIFY_CLIENT_ID=local-client-id
    SPOTIFY_CLIENT_SECRET=local-client-secret
    SPOTIFY_REDIRECT_URL=http://localhost:8080/auth/spotify/callback
    REDIS_ADDR=localhost:6379
    ```

2. **`.env.dev`** - Used for the dev enviroment deployments
    ```dotenv
    SPOTIFY_CLIENT_ID=dev-client-id
    SPOTIFY_CLIENT_SECRET=dev-client-secret
    SPOTIFY_REDIRECT_URL=http://dev.yourdomain.com/auth/spotify/callback
    REDIS_ADDR=dev-redis-server:6379
    ```

3 **`.env.ci`** - Used for the ci intergration tests

4 **`.env.prod`** - Used for the prod enviroment 
    ```dotenv
    SPOTIFY_CLIENT_ID=prod-client-id
    SPOTIFY_CLIENT_SECRET=prod-client-secret
    SPOTIFY_REDIRECT_URL=https://yourdomain.com/auth/spotify/callback
    REDIS_ADDR=prod-redis-server:6379
    ```

`REDIS_ADDR=localhost:6379`

# Running the Application

Set the desired APP_ENV and start the application:

Windows (PowerShell): `$env:APP_ENV = "local"`

Linux/Mac: export APP_ENV=local"

Start the Application

   `go run main.go`

Access the Endpoints

Swagger UI: http://localhost:8080/swagger/

## Generating Endpoints and Editing the OpenAPI Spec

1. Open the openapi.yaml file and make changes to your API definitions.
2. Run the following command to generate or update the handlers and types:
`oapi-codegen -generate chi-server,types,spec -o generated/server.go -package generated openapi.yaml`
3.  Implement New Endpoints
    •	Update the generated ServerInterface in your handlers package to implement the new endpoints.

4. Generate Clients - oapi-codegen -generate types,client -package authclient -o ./auth-service-client/client.gen.go ./openapi.yaml

example:

`func (s *Server) RefreshAuthToken(w http.ResponseWriter, r *http.Request, provider string) {
   // Logic for refreshing the token
}`
## Testing

`go test`

## Creating Builds

You can dynamically specify the environment with Docker Compose to read from the appropriate .env:

Windows (PowerShell):
$env:APP_ENV = "dev"
docker-compose up --build

Linux/Mac:
export APP_ENV=dev
docker-compose up --build


## License

This project is licensed under the MIT License. See the LICENSE file for details.

todo: add license file