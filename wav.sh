#!/bin/bash

AUDIO_IN=$(pwd)/data/input
AUDIO_OUT=$(pwd)/data/output
MODEL_DIRECTORY=$(pwd)/data/model

echo "[wav.sh] Running spleeter ✨"

docker run \
    -v $AUDIO_IN:/input \
    -v $AUDIO_OUT:/output \
    -v $MODEL_DIRECTORY:/model \
    -e MODEL_PATH=/model \
    researchdeezer/spleeter \
    separate -o /output -i $1 --verbose
