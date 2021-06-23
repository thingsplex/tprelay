FROM alpine:edge

COPY ./package/build/tprelay /bin/tprelay

EXPOSE 8090/tcp

ENV BIND_ADDRESS=":8090"

CMD ["/bin/tprelay"]