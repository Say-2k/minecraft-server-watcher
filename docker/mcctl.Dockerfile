FROM eclipse-temurin:17-jre

WORKDIR /minecraft

ARG ARTIFACT_PATH


RUN apt-get update && \
apt-get install -y wget && \
apt-get clean && \
rm -rf /var/lib/apt/lists/*

# Создаем пользователя
RUN groupadd -r minecraft && \
useradd -r -g minecraft minecraft

# Меняем владельца рабочей директории
RUN chown -R minecraft:minecraft /minecraft

# Переключаемся на пользователя minecraft
USER minecraft

# Объявляем том
VOLUME /minecraft

COPY ${ARTIFACT_PATH} /minecraft/ 

# Команда для запуска сервера
CMD ["./run.sh"]
