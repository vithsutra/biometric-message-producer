#!/bin/bash

# ========= CONFIG =========
HOST="biometric.mqtt.broker1.vithsutra.com"
PORT="1883"
TOPIC="vs23cg003/process/attendance/message"
# ==========================

echo "Publishing messages to $HOST:$PORT on topic $TOPIC"

for i in $(seq 1 100); do
    # Decide if we want to use a duplicate mid
    if (( i % 10 == 0 )); then
        # Every 10th message reuse an old mid
        MID="DUPLICATE-MID-1234"
    else
        MID="MID-$(uuidgen)"
    fi

    JSON=$(cat <<EOF
{
    "mid":"$MID",
    "sid":7,
    "index":$i,
    "tmstmp":"2006-01-02T15:04:05"
}
EOF
)

    mosquitto_pub -h "$HOST" -p "$PORT" -t "$TOPIC" -m "$JSON"
    echo "Published message $i with mid: $MID"
    sleep 0.1
done

echo "Done publishing 50 messages."

