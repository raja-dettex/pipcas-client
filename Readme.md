

# About Pipcas Client

Pipcas client exposes http endpoints to save and retreive files to and from Pipcas Underlying storage for persistence. Pipcas-client also keeps a copy of recently added or fetched file from backend server onto its local file system for a certain period of time. Specify TTL in the environment variables. Least recently used files are evicted according to the ttl specified. 


# Quickstart

## Deploy the binaries on your local machine directly onto your host os.

make sure Go 1.18 is installed and it's path variable is configured

clone the repository
```
    git clone https://github.com/raja-dettex/pipcas-client
```
```
    cd pipcas-lb
```

```
    go mod tidy
```

```
    make build
```

```
    LISTEN_ADDR=<> ASTRA_DB_ID=<> ASTRA_DB_TOKEN=<> ASTRA_DB_KEYSPACE=<> KEYSPACE=<> GATEWAY_HOST=<> GATEWAY_PORT=<> TTL=<> make run
```

## Deploy onto a docker cotainer

after cloning the repository as mentioined previously,

Build a Docker image from the Docker file

```
    docker build -t <your image name> .
```

Start the docker container from the built image

```
    docker run -d -p <port to access contiainer process>:<port according to listen address> -e  LISTEN_ADDR=<> ASTRA_DB_ID=<> ASTRA_DB_TOKEN=<> ASTRA_DB_KEYSPACE=<> KEYSPACE=<> GATEWAY_HOST=<> GATEWAY_PORT=<> TTL=<> <your image name>
```

## latest release: 
    pipcas-lb:1.0
