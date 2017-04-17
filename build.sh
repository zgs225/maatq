#! /bin/bash

git tag $1
git push origin $1
echo $1 > VERSION
