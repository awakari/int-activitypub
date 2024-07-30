#!/bin/bash

COVERAGE=$(cat cover.tmp)
THRESHOLD=34
if [[ ${COVERAGE} -lt ${THRESHOLD} ]]; \
  then \
    echo "FAILED: test coverage ${COVERAGE}% < ${THRESHOLD}%"; \
    exit 1; \
  else \
    echo "PASSED: test coverage ${COVERAGE} >= ${THRESHOLD}%"; \
fi
