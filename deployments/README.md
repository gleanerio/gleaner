# Deployments

## Compose

```
export DATAVOL=/home/fils/Data/P418/dataVolumes
```

## Swarm

```
docker stack deploy --compose-file=hatest.yml gleaner

docker-compose -f gleaner-compose.yml down
docker-compose -f gleaner-compose.yml up -d
docker-compose -f hatest.yml up -d
docker-compose -f hatest.yml down

docker rm $(docker ps -a -f status=exited -q)

docker stack deploy --compose-file=hatest.yml gleaner
docker stack ls
docker stack ps gleaner
docker stack services gleaner
docker stack rm gleaner
docker stop $(docker ps -a -q)
docker rm $(docker ps -a -q)
docker ps
```

```
docker-compose up -d
docker-compose scale web=3
docker-compose ps you can see there are 4 containers run
```
