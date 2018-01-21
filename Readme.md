# Publish-subscribe system

## Dependencies:
- Go
- Node
- Docker
- docker-compose

## Task:
- create publish-subscribe system with three different parts:
    - tracking service -> service that receives request for publishing data
    - publish service -> service that receives messages and distributes them to others
    - subscribe service -> service that subscribes on publish service and receives messages from it.

## Technologies used:
- for *tracking* and *subscribe* service, Go lang was used. Tracking service also uses MongoDB for it ease of use.
- for communication between services (*publish* service) WebSockets were used. They were used because of its asynchronous, event driven arhitecture, which is great for a task of publishing and subscribing between services.

## How to run
NOTE: whole setup was tested and works on Linux machine. If you are using MacOS and it is not working, please contact me to fix things.

- first you run `make devbox/run`, which will build needed docker images and run 3 docker containers for you. On port `:8080` you should see `POST` endpoint exported.
- if you want to include some basic data, run `make demo-date` inside `tracker` folder. This will generate 16 accounts, with ids from 5937e2d316ca1b6d4066aa20 up to 5937e2d316ca1b6d4066aa2f. First 8 account will have `isActive` set to true. 
- to run client for subscribe run `make run/aggregator` or `make run/printer`. To add filtering by ID, run `make run/aggregator/:ACC_ID` or `make run/printer/:ACC_ID`

If you have any question, contact me.