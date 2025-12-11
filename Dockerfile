# Build
FROM alpine

COPY drk /drk
COPY AmazonRootCA1.pem /AmazonRootCA1.pem

RUN chmod +x /drk

ENTRYPOINT [ "/drk" ]
