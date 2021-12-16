FROM ubuntu

COPY target /var/packages/pan-xunlei-com/target
COPY xunlei-linux /xunlei-linux
COPY host  /var/packages/pan-xunlei-com/host

VOLUME [ "/var/packages/pan-xunlei-com/shares" ]

CMD [ "/xunlei-linux" ]