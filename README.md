# DadJokes
Simple service to handle storing all the cringy dad jokes and access through a RESTful API

- Utlizes gin-gonic HTTP framework for the API
- Connects to MongoDb for local storage
- makes HTTP request to external API for random jokes for inspiration

## To Start

    $ docker compose -f docker-compose.yml up --build
Or

    $ ./scripts/start.sh

## To Use
Through Postman (or any other API platform)

    GET http://localhost:8082/pings 
    POST http://localhost:8082/joke
    GET http://localhost:8082/joke 
    GET http://localhost:8082/jokes/all
    GET http://localhost:8082/joke/random  


