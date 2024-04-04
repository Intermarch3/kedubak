## How to run this app :
rename the .env.sample file to .env and fill the empty variable "MONGODB_URI"

##### Then just run this few commands :
docker compose build
docker compose up

##### or to run only the api :
docker build -t kedubak .
docker run -p 8080:8080 --env-file .env kedubak