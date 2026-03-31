FROM ubuntu:latest

ARG BOT_ARTIFACT_PATH

WORKDIR /app

COPY ${BOT_ARTIFACT_PATH} /app/bot

RUN apt-get update && apt-get install -y ca-certificates
RUN chmod +x /app/bot

CMD ["sh", "-c", "./bot"]