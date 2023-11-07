# syntax=docker/dockerfile:1

FROM python:3.10-slim

RUN apt-get clean
RUN apt-get update && apt-get install -y apt-transport-https
RUN apt-get install -y git
RUN apt-get install -y ffmpeg libsndfile1-dev

WORKDIR /app

COPY . .

RUN pip install pydub
RUN pip install torch==2.0.1 torchaudio==2.0.2
RUN pip install --upgrade --user hmmlearn
RUN pip install pyannote.audio==3.0
RUN pip install git+https://github.com/openai/whisper.git

RUN python download.py

CMD ["python", "transcribe.py"]



