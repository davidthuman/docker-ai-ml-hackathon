# Docker AI/ML Hackathon

The [Docker AI/ML Hackathon](https://docker.devpost.com/) wants users to build or design a hack "that use[s] Docker products to help both beginners and advanced users get inspired, get started, and be productive within this exciting new frontier."

## Requirements

### WHAT TO BUILD

Create a working software application, non-code proof of concept, or design with a strong concept validation that uses AI/ML in conjunction with Docker technologies.

*We favor submissions that are as close to a real-world implementation as possible, but accept submissions that are not fully functional if they show a strong proof of concept.

**Examples:**
- AI/ML projects or models built using Docker technology and distributed through DockerHub
- AI/ML integrations into Docker products which improves the developer experience
- Enhancement or extensions of Docker products which makes developer working with AI/ML more productive

### WHAT TO SUBMIT

- Project built with AI/ML and any Docker technology
- URL to your app, non-code proof of concept, or design 
- Text description explaining the features and functionality of your Project currently or in future
- 3-5 minute demo video

## Goal of this Project

The goal is to create an web-application for users to upload an audio file, have it be transcribed with an AI model, and allow the user to then edit and save the transcription. Once saved, the user will have the option to fine-tune their own model with their edited transcriptions.

### Front-end

We are hoping to use [HTMX](https://htmx.org/), a tool-kit for HTML that extends the hypertext's capabilities without using Javascript.

### Back-end

We are hoping to use [Go](https://go.dev/) to handle any HTTP requests, communicate to the database, and call on any AI models.

### Database

We are hoping to use [SQLite](https://www.sqlite.org/index.html) to hold user information and transcription data.

### AI Model

We are hoping to use [Python](https://www.python.org/) with [Pytorch](https://pytorch.org/) to call OpenAI's [Whisper](https://openai.com/research/whisper) model.

### Docker

We are hoping to user [Docker](https://www.docker.com/) to build images of call-able models and for fine-tuning.

