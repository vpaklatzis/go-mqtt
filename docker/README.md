# GO-MQTT

## Prerequisites

* Go version 18. Check with `go version` command.
* Docker and docker-compose. Check with `docker -v` and `docker-compose -v` commands.
* Bash

## Project Structure

* `binds` folder: Holds the mosquitto and rabbit-mq configuration files that are passed in the containers, respectivelly.

* `consumer` project: The mqtt consumer is developed in go and connects to the mqtt broker host running on localhost and port 1883, via tcp. It accepts and prints to the terminal mqtt the messages that are being produced and passed through the broker via the topic/temps of the broker.

* `producer` project: The mqtt producer is developed in go and connects to the mqtt broker host running on localhost and port 1883, via tcp. It produces mqtt messages, prints them to the terminal and sends them directly to the topic/temps of the broker. A delay of 1 second is set, to limit the rate of messages being produced. Otherwise, the produces sends and absurd amount of messages.

* `build-linux.sh` file: Compiles the go projects present in the root directory to linux binaries.

* `docker-compose` folder: Contains all the docker-compose files.

    * `emqx.yaml` file: docker-compose file, contains the necessary configuration to run the EMQX docker image inside a docker container.

    * `hivemq.yaml` file: docker-compose file, contains the necessary configuration to run the HiveMQ docker image inside a docker container.

    * `mosquitto.yaml` file: docker-compose file, contains the necessary configuration to run the Mosquitto docker image inside a docker container.

    * `rabbitmq.yaml` file: docker-compose file, contains the necessary configuration to run the RabbitMQ docker image inside a docker container.

    * `vernemq.yaml` file: docker-compose file, contains the necessary configuration to run the VerneMQ docker image inside a docker container.

    * `services.yaml` file: docker-compose file, contains the necessary configuration to run the producer and consumer images inside docker containers.

The brokers by default accept mqtt connections and messages to the port 1883 via tcp.

## Initials steps

* Open the terminal and change directory into the docker folder. 

* You can start any of the above broker containers, using docker-compose. E.g `docker-compose -f emqx.yaml up`.

* In another terminal window, change directory into the consumer project. Run the `go run main.go` command.

* In another terminal window, change directory into the producer project.Run the `go run main.go` command.

You should be able to observe in the terminals the messages that are being produced, sent to the broker and being consumed.
