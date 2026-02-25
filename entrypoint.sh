#!/bin/bash

export LINES=40
export COLUMNS=120
export TERM=xterm-256color

echo "=== Starting tron IDE ==="
echo "TERM=$TERM LINES=$LINES COLUMNS=$COLUMNS"

script -q -c "timeout 3 ./tron 2>&1" /dev/null 2>&1 | cat -v

echo ""
echo "=== Program exited with code: $? ==="
