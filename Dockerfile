FROM golang as build

WORKDIR /sources/

COPY . .

ENV GOPROXY=https://goproxy.cn GO111MODULE=on
RUN go build -o /xunlei-from-syno ./

FROM ubuntu

COPY --from=build /xunlei-from-syno /xunlei-from-syno
COPY host /var/packages/pan-xunlei-com/host
COPY target /var/packages/pan-xunlei-com/target

VOLUME [ "/var/packages/pan-xunlei-com/shares", "/downloads" ]

CMD [ "/xunlei-from-syno", "run", "--port=2345", "--download-dir=/downloads" ]
