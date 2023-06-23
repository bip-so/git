#!/bin/bash
# change this file permission to exec `sudo chmod 755 swagger.sh`

if ! [ -x "$(command -v swag)" ]; then
    echo "swag is not present. Installing it..."
    go get github.com/swaggo/swag/cmd/swag
fi

export PATH=$(go env GOPATH)/bin:$PATH
swag init
