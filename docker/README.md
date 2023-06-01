# GO-MQTT

## Prerequisites

* Go version 18. Check with `go version` command.
* Docker and docker-compose. Check with `docker -v` and `docker-compose -v` commands.
* Bash

## Project Structure

* `api` folder: The api project spins up an http server using Gin Framework, on port 8080. It receives data from the suscriber and persists it to MongoDB.

* `ble` folder: The ble project scans for BLE signals and parses the data received from the signal.

* `binds` folder: Holds the mosquitto and RabbitMQ configuration files that are passed in the containers, respectivelly.

* `k6` folder: Holds the api performance testing files written in javascript.

* `subscriber` project: The mqtt subscriber is developed in go and connects to the mqtt broker host running on localhost and port 1883, via tcp. It accepts and prints to the terminal mqtt the messages that are being published and passed through the broker via the topic/temps of the broker.

* `publisher` project: The mqtt publisher is developed in go and connects to the mqtt broker host running on localhost and port 1883, via tcp. It publishes mqtt messages, prints them to the terminal and sends them directly to the /sensor/temp topic of the broker. A delay of 1 second is set, to limit the rate of messages being published.

* `build-linux.sh` file: Compiles the go projects present in the root directory to linux binaries.

* `docker-compose` folder: Contains all the docker-compose files.

    * `emqx.yaml` file: docker-compose file, contains the necessary configuration to run the EMQX docker image inside a docker container.

    * `hivemq.yaml` file: docker-compose file, contains the necessary configuration to run the HiveMQ docker image inside a docker container.

    * `mosquitto.yaml` file: docker-compose file, contains the necessary configuration to run the Mosquitto docker image inside a docker container.

    * `rabbitmq.yaml` file: docker-compose file, contains the necessary configuration to run the RabbitMQ docker image inside a docker container.

    * `vernemq.yaml` file: docker-compose file, contains the necessary configuration to run the VerneMQ docker image inside a docker container.

    * `services.yaml` file: docker-compose file, contains the necessary configuration to run the publisher and subscriber images inside docker containers.

    * `mongo.yaml` file: docker-compose file, contains the necessary configuration to run MongoDB docker image inside a docker container.

The brokers by default accept mqtt connections and messages to the port 1883 via tcp.

## Initials steps

* Open the terminal and change directory into the docker folder. 

* You can start any of the above broker containers, using docker-compose. E.g `docker-compose -f mosquitto.yaml up`.

* In another terminal window, change directory into the subscriber project. Run the `go run main.go` command.

* In another terminal window, change directory into the publisher project.Run the `go run main.go` command.

You should be able to observe in the terminals the messages that are being published, sent to the broker and being consumed by the subscriber.
