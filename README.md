# SocialGamesHoster
Project to host social games on your local wifi network.

## Setup
You might have to change your firewall settings to allow connections to the web app.
For Windows, go into "Windows Defender with Firewall with Advanced Security" and add an inbound TCP rule for port 8081 on private networks.

## Architecture
TODO:
- Clean architecture could work with 4 layers. (https://www.youtube.com/watch?v=1OLSE6tX71Y)
- You can have a separate project just for the program.cs start up.
- OR monolithic seems ok as the API is currently not very complicated.
	- We should however put the data access stuff in a folder
https://www.youtube.com/watch?v=pYl_jnqlXu8 - Minimal API looks better.
https://www.youtube.com/watch?v=a1ye9eGTB98 - For result type, but maybe exception middleware is better

## The following is no longer true
The API uses an N-tier architecture with the following layers:
- API 
	- contains the routes and controllers, and DTO models 
	- Depends on the application layer
	- Depends on the infrastructure layer with the sole purpose of initializing the DbContext 
Application 
	- Contains the services and business logic
	- Is responsible for converting DTO objects to Domain objects
	- Depends on the API for the DTO objects (?).
	- Depends on Domain for the models
	- Depends on infrastructure for the repositories
- Domain
	- contains the models
- Infrastructure
	- is responsible for database communication
