#!/bin/bash -x

# creates the necessary docker images to run testrunner.sh locally

docker build --tag="cjminercn/cppjit-testrunner" docker-cppjit
docker build --tag="cjminercn/python-testrunner" docker-python
docker build --tag="cjminercn/go-testrunner" docker-go
