package bob

// TODO: make this a valid compose file without port clashes.
// If you want to simulate port clashes add a test which assures
// that some ports are blocked.
var dockercompose = []byte(`version: '3'
services:
  adminer:
    image: adminer
    restart: always
    ports:
      - "8080:8080"
      #- 50001:50001
      #- 5555:5555/udp
      #- 9090-9091:8080-8081
    depends_on:
      - mysql
  mysql:
    image: mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: pass
    ports:
      - "3306:3306"
      #- 9090-9091:8080-8081 # weird case
      #- 6379:6379 # conflict with local env and mongo
  mongo:
    image: mongo
    restart: always
    environment:
      MONGO_INITDB_ROOT_USERNAME: user
      MONGO_INITDB_ROOT_PASSWORD: pass
    ports:
      - "27017:27017"
      #- 9090-9091:8080-8081 # weird case
      #- 6379:6379 # conflict with local env and mysql
      #- 5555:5558/udp # different container port, but host collides
`)

var dockercomposewhoami = []byte(`version: '3'
services:
  whoami:
    image: containous/whoami
`)
