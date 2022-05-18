FROM alpine:edge

COPY ./package/build/tprelay /bin/tprelay

EXPOSE 8090/tcp

ENV BIND_ADDRESS=":8090"
ENV LOG_LEVEL="info"

CMD ["/bin/tprelay"]