import os
from pydub import AudioSegment
from pyannote.audio import Pipeline
import torch
import re
import whisper
import json
from datetime import timedelta

print("Setting up files")

TMP_FOLDER = "tmp"  # temporary folder for work
WORK_FOLDER = "work"

if not os.path.exists(TMP_FOLDER):
    os.mkdir(TMP_FOLDER)
if not os.path.exists(WORK_FOLDER):
    os.mkdir(WORK_FOLDER)

FILE = list(filter(lambda x: "input" in x, os.listdir("./tmp")))[0]

FILE_PATH = os.path.join(TMP_FOLDER, FILE)

print("Create prep audio file")

# Create silent audio clip
spacermilli = 2000
spacer = AudioSegment.silent(duration=spacermilli)

# Load audio file
file_extension = FILE.split(".")[-1]  # get audio file extension
audio = AudioSegment.from_file(FILE_PATH, file_extension)
# Append silent clip
audio = spacer.append(audio, crossfade=0)
# Save modified audio
PREP_FILE = os.path.join(WORK_FOLDER, 'input_prep.wav')
audio.export(PREP_FILE, format='wav')

print("Loading PyAnnote Diarization model")

# Hugging Face Access Token
access_token = "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"

# Download Hugging Face model using access token
# Will download the model or use a cache
pipeline = Pipeline.from_pretrained('pyannote/speaker-diarization-3.0', use_auth_token = access_token, cache_dir = './.cache/pyannote')

# Define machine device and set model's device

# Get device
device  = "cpu"
if torch.cuda.is_available():
    device = "cuda"
if torch.backends.mps.is_available():
    device = "mps"
    
device = torch.device(device)
pipeline.to(device)

print("Process audio file for diarization")

DIARIZATION_FILE = os.path.join(WORK_FOLDER, "diarization.txt")
DEMO_FILE = {'uri': 'blabla', 'audio': PREP_FILE}

# Run pipeline
diarizations = pipeline(DEMO_FILE)

# Save diarization times
with open(DIARIZATION_FILE, "w") as text_file:
    text_file.write(str(diarizations))

def millisec(timeStr):
    spl = timeStr.split(":")
    s = (int)((int(spl[0]) * 60 * 60 + int(spl[1]) * 60 + float(spl[2]) )* 1000)
    return s

# Load diarization split
if diarizations is None:
    diarizations = open(DIARIZATION_FILE).read().splitlines()
else:
    diarizations = str(diarizations).split("\n")

print("Create speaker audio files")

groups = []
group = []
lastend = 0

for diarizarion in diarizations:  # for each diarization split
    if group and (group[0].split()[-1] != diarizarion.split()[-1]):  # if the same speaker
        groups.append(group)
        group = []

    group.append(diarizarion)  # Append the diarization information

    end = re.findall('[0-9]+:[0-9]+:[0-9]+\.[0-9]+', string=diarizarion)[1]  # Regex to find the ending time-string
    end = millisec(end)  # convert to mili-seconds
    if (lastend > end):  # if segment engulfed by a previous segment
        groups.append(group)
        group = []
    else:
        lastend = end

if group:  # append final temporary grouping
    groups.append(group)

audio = AudioSegment.from_wav(os.path.join(WORK_FOLDER, "input_prep.wav"))
gidx = -1
for group in groups:
    start = re.findall('[0-9]+:[0-9]+:[0-9]+\.[0-9]+', string=group[0])[0]
    end = re.findall('[0-9]+:[0-9]+:[0-9]+\.[0-9]+', string=group[-1])[1]
    start = millisec(start) #- spacermilli
    end = millisec(end)  #- spacermilli
    gidx += 1
    audio[start:end].export(os.path.join(WORK_FOLDER, str(gidx) + '.wav'), format='wav')

del   DEMO_FILE, pipeline, spacer,  audio, diarizations

print("Load OpenAI Whisper model")

# Get device
device  = "cpu"
if torch.cuda.is_available():
    device = "cuda"
if torch.backends.mps.is_available():
    device = "mps"

model = whisper.load_model(name = 'medium.en', download_root='./.cache/whisper')  # load Whisper model

print("Process audio file for speech-to-text")

transcript = []

for i, group in enumerate(groups):

    audiof = os.path.join(WORK_FOLDER, str(i) + '.wav')  # audio file path
    result = model.transcribe(audio=audiof, language='en', word_timestamps=True)  # transcribe audio file

    # Get time shift
    shift = re.findall('[0-9]+:[0-9]+:[0-9]+\.[0-9]+', string=group[0])[0]  # get starting time for speaker cluster
    shift = millisec(shift) - spacermilli  # the start time in the original video
    shift = max(shift, 0)  # time shift in miliseconds

    # Lambda function to apply time shift
    apply_shift = lambda time: (shift + (time * 1000.0)) / 1000.0

    segments = result["segments"]
    speaker = group[0].split()[-1]  # first section in speaker cluster, get speaker name

    if not segments:
        continue

    for segment in segments:

        # Update times for segment
        segment['start'] = apply_shift(segment['start'])
        segment['end'] = apply_shift(segment['end'])

        for i, word in enumerate(segment['words']):
            if word == "":
                continue
            # Update times for word
            word['start'] = apply_shift(word['start'])
            word['end'] = apply_shift(word['end'])

    result['segments'] = segments
    transcript.append({"speaker": speaker, "result": result})

print("Save transcript")

with open(os.path.join(TMP_FOLDER, 'transcript'+'.json'), "w") as outfile:  # write result
    json.dump(transcript, outfile, indent=4)