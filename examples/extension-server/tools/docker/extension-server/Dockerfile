FROM gcr.io/distroless/static:nonroot@sha256:e9ac71e2b8e279a8372741b7a0293afda17650d926900233ec3a7b2b7c22a246
COPY ./bin/extension-server /usr/local/bin/extension-server

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/extension-server"]
