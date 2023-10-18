
## Getting Started






# Speech-to-Text for Oral Argument Audios

Oral Argument Audio to Text
- Speech-to-Text model
    - Would like confidence score for each token
    - Resources
        - [Speech2Text](https://huggingface.co/docs/transformers/model_doc/speech_to_text)
        - [wav2vec](https://pytorch.org/audio/stable/tutorials/speech_recognition_pipeline_tutorial.html)
        - [Otter.ai](https://otter.ai)
- Speaker Diarization feature to tag speaker
    - Need to know Judges voices
    - Resources
        - [Speaker Embedding for Diarization in PyTorch](https://github.com/WiraDKP/pytorch_speaker_embedding_for_diarization)
- Front-End for users to listen and proof-read translation
    - Show the model's confidence score
    - Allow user to change a token

https://github.com/speechbrain/speechbrain


For purely speech-to-text (seq2seq model), OpenAI's [Whisper](https://github.com/openai/whisper) is very good.

[Improving Timestamp Accuracy](vhttps://github.com/openai/whisper/discussions/435)

[Transcription and Diarization](https://github.com/openai/whisper/discussions/264)


We want a UI that allows the user to hear the audio and review the transcription. Need to color each word with confidence score, and allow single word editing of the transcription



[`pyannote.audio`](https://github.com/pyannote/pyannote-audio) is an open-source toolkit written in Python for **speaker diarization**.

Based on [`PyTorch`](https://pytorch.org) machine learning framework, it provides a set of trainable end-to-end neural building blocks that can be combined and jointly optimized to build speaker diarization pipelines.

`pyannote.audio` also comes with pretrained [models](https://huggingface.co/models?other=pyannote-audio-model) and [pipelines](https://huggingface.co/models?other=pyannote-audio-pipeline) covering a wide range of domains for voice activity detection, speaker segmentation, overlapped speech detection, speaker embedding reaching state-of-the-art performance for most of them.