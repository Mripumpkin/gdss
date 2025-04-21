FROM alpine:3.18

WORKDIR /app


COPY bin/gdss /app/gdss

RUN chmod +x /app/gdss

EXPOSE 7799

ENTRYPOINT ["/app/gdss"]
