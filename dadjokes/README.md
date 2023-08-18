# DadJokes
Simple service to handle storing all the cringy dad jokes and access through a RESTful API

- Utlizes gin-gonic HTTP framework for the API
- Connects to MongoDb for local storage

## To Start

    $ docker compose -f docker-compose.yml up --build
Or
    $ ./scripts/start.sh

Through Postman 

GET http://localhost:5001/pings 
POST http://localhost:5001/joke
GET http://localhost:5001/joke 


