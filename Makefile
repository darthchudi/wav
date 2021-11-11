.PHONY: split
split:
	 @docker run \
        -v $AUDIO_IN:/input \
        -v $AUDIO_OUT:/output \
        -v $MODEL_DIRECTORY:/model \
        -e MODEL_PATH=/model \
        researchdeezer/spleeter \
        separate -o /output -i /input/temp/$1 --verbose
