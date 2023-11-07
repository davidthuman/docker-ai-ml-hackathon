from pyannote.audio import Pipeline
import whisper
import torch

# Get device
device  = "cpu"
if torch.cuda.is_available():
    device = "cuda"
if torch.backends.mps.is_available():
    device = "mps"

# Hugging Face Access Token
access_token = "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"

# Download PyAnnote model
print("PyAnnote")
Pipeline.from_pretrained('pyannote/speaker-diarization-3.0', use_auth_token = access_token, cache_dir = './.cache/pyannote')

# Download Whisper medium.en model
print("Whisper")
whisper.load_model(name = 'medium.en', device = device, download_root='./.cache/whisper')  # load Whisper model