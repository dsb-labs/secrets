FROM gcr.io/distroless/static

ARG TARGETPLATFORM

COPY $TARGETPLATFORM/keeper /usr/bin/keeper
COPY LICENSE.md /usr/bin/LICENSE.md
COPY README.md /usr/bin/README.md
COPY licenses /usr/bin/licenses

# Default behaviour with no arguments is to just run the keeper server on port 53.
ENTRYPOINT ["/usr/bin/keeper"]
CMD ["serve"]
