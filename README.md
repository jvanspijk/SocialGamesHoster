# SocialGamesHoster
**N.B. This is a work in progress.**
Project to host social deduction games on your local network.

## User setup
1. Install docker engine or docker desktop.
2. Run the docker-compose-prod.yaml.
3. You might have to change your firewall settings to allow connections to the web app.

(For Windows, go into "Windows Defender with Firewall with Advanced Security" and 
add an inbound TCP rule for port 80 on private networks.)

## Developer setup
1. Run the docker-compose.yaml.

Both the API and web app use hot reloading, so code changes appear instantly for ease of development.
Swagger documentation can be found on http://localhost:8080/swagger/

## Docker compose
Docker compose was used for the following reasons:
- So that users and hosts don't have to fiddle with dependencies.
- Hosts just need to install Docker Desktop/Engine and then copy-paste a single line into their terminal to get everything set up.
- To keep the development environment identical to the host environment and thus to prevent environment specific bugs.

The docker compose consists of four services:
- The API (ASP.NET)
- The Web app (Sveltekit)
- A database (Postgres)
- nginx

## Architecture
The API and Web app are separate projects because Svelte offers a better and more comprehensive developer experience for creating a good looking UI and reactive user experience.

The API is designed to be mostly stateless to keep the logic simple 
and prevent data loss if the network is down.
The API uses Entity Framework to store and retrieve C# objects in a database without having to manually write queries. The database implementation can be swapped easily and an in-memory database is used for testing.
