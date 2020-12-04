#!/bin/bash

go mod vendor
git add .
git commit -m "$1"
git push origin $2
