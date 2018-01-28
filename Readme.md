# Publish-subscribe system

## Dependencies:
- Go
- Node
- Docker
- docker-compose
- [dep](https://github.com/golang/dep)

## Task:
- create a publish-subscribe system with three different parts:
    - tracking service -> service that receives a request for publishing data,
    - publish service -> service that receives messages and distributes them to others,
    - subscribe service -> service that subscribes on publish service and receives messages from it.

## Technologies used:
- for *tracking* and *subscribe* service, [Golang](https://golang.org/) was used. Tracking service uses MongoDB for storing data because of its ease of use.
- for communication between services (*publish* service), WebSockets were used. They were used because of its asynchronous, event-driven architecture, which is great for a task of publishing and subscribing between services. The service was written in JavaScript, running on Node.js runtime.

## How to run
NOTE: the whole setup was tested and works on Linux machine. If you are using MacOS and it is not working, please contact me to fix things.

NOTE2: Please make sure that ports `8000`, `8080` and `27017` are free so that containers can easily expose ports.

- first run `make devbox/run`, which will build needed docker images and run 3 docker containers for you. On port `:8080` you should see `POST` endpoint exported.
- if you want to include some basic data, run `make demo-data` inside `tracker` folder. This will generate 16 accounts, with ids from 5937e2d316ca1b6d4066aa20 up to 5937e2d316ca1b6d4066aa2f. First 8 account will have `isActive` set to true. 
- to run the client for subscribing run `make run/aggregator` or `make run/printer`. To add filtering by ID, run `make run/aggregator/:ACC_ID` or `make run/printer/:ACC_ID`

## Tests
To run tests, run `make qa` in the root folder. This should run all tests for you. Please ensure that your devbox is running.

NOTE: because of many services running, there might be the case, where tests are failing. Please rerun tests if this happens.

## Other
Please run `make help` in root folder to get a list of all possible commands available.

If you want to test *subscribe* service resilience to error, please shutdown devbox with `make devbox/stop` not directly by `docker-compose restart` because of strange behavior of services when this is used.

There are some `sudo` commands in makefiles etc. This is due to Linux environment.

If you have any question, contact me.