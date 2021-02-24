#!/bin/bash

set -e

RANDOM=$$

function get_task {
    word=$(shuf -n 1 /usr/share/dict/words)
    duration=10000

    jq -n -r --arg task "$word" --arg duration $duration '{($task): $duration}'
}

j=$(get_task)
for i in $(seq $(( $1 - 1 ))); do
    t=$(get_task)
    j=$(echo $j $t | jq -s 'add | map_values(tonumber)')
done

echo $j

curl -s -i -d "$j" http://localhost:8080/
