# TradeService

TradeService follows the Restful api specification and provides trade services for miners and validators, which include:

1. Receive the trades from all the miners.
2. Process validator queries to check trades & prices.

You can view the TradeService as a buffer between miners and validators.

Right now the service is run by OpenScope team, we plan to require all the validator to run their own trade service very soon.

## Run the service

You can try to run the trade service follow the below steps now.

## Prerequisite

### Install Dependencies docker and docker compose

    1. Set up Docker's apt repository.

    ```bash

    # Add Docker's official GPG key:
    sudo apt-get update
    sudo apt-get install ca-certificates curl
    sudo install -m 0755 -d /etc/apt/keyrings
    sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    sudo chmod a+r /etc/apt/keyrings/docker.asc

    # Add the repository to Apt sources:
    echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update

    ```

    2. Install the Docker packages.

    To install the latest version, run:

    ```bash

    sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    ```

### Build from source code

    ```bash

    make build

    ```

### Start TradeService

    ```bash

    docker compose up -d

    ```

### Start TradeService

    ```bash

    docker compose down

    ```