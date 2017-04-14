#! /bin/bash

ORG="youplus"
APP="mataq"
REPO="docker.youplus.cc"

docker build -t "$ORG/$APP" .
docker tag "$ORG/$APP" "$REPO/$APP:latest"
docker push "$REPO/$APP:latest"

if [ "$#" -eq 1 ]; then
    VERSION=$1
    docker tag "$ORG/$APP" "$REPO/$APP:$VERSION"
    docker push "$REPO/$APP:$VERSION"
    echo $VERSION > VERSION
fi
