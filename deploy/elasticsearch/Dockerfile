FROM curlimages/curl:latest
WORKDIR /app

COPY ./indexes ./indexes
COPY ./init.sh ./init.sh

ENTRYPOINT ["./init.sh"]