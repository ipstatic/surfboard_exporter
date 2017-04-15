#!/bin/bash
mkdir build
for build in $(find .build -type f); do
  IFS='/' read -ra path <<< "$build"
  arch=${path[1]}
  mv "${build}" "build/surfboard-exporter-${TRAVIS_TAG}.${arch}"
done
