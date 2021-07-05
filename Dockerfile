FROM centurylink/ca-certs
WORKDIR /app
COPY ./wg /app
COPY ./migrations /app/migrations

CMD ["/app/wg"]