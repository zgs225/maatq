#! /bin/bash

APP=mataq

# Build docker
docker build -t $APP .
docker tag $APP $APP:latest

if [ "$#" -eq 1 ]; then
    VERSION=$1

    docker tag $APP $APP:$VERSION
    docker push $APP:$VERSION

    # Tag github
    git tag $VERSION
    git push origin $VERSION

    echo $VERSION > VERSION
fi
