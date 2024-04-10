#!/usr/bin/env bash

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <input_file>"
    exit 1
fi

input_file=$1

temp_file=$(mktemp)

sed -n '
/^# / {
    s/^# \(.*\)/+++\ntitle = "\1"\n+++\n/
    p
    d
}
p
' "$input_file" > "$temp_file"

mv "$temp_file" "$input_file"
