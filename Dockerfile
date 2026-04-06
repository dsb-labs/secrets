FROM gcr.io/distroless/static

ARG TARGETPLATFORM

COPY $TARGETPLATFORM/secrets /usr/bin/secrets
COPY LICENSE.md /usr/bin/LICENSE.md
COPY README.md /usr/bin/README.md
COPY licenses /usr/bin/licenses

# Default behaviour with no arguments is to just run the secrets server on port 53.
ENTRYPOINT ["/usr/bin/secrets"]
CMD ["serve"]
