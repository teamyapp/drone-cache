#!/bin/bash

VERSION=$1

docker build -t "drone-cache:$VERSION" .
docker tag "drone-cache:$VERSION" "ghcr.io/teamyapp/drone-cache:$VERSION"
docker push "ghcr.io/teamyapp/drone-cache:$VERSION"