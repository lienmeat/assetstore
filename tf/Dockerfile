FROM alpine:3.3

RUN apk --update upgrade \
  && apk add ca-certificates \
  && apk add curl \
  && apk add unzip \
  && apk add python \
  && update-ca-certificates \
  && rm -rf /var/cache/apk/*

#AWS S3 Install
RUN curl "https://s3.amazonaws.com/aws-cli/awscli-bundle.zip" \
    -o "awscli-bundle.zip" \
    && unzip awscli-bundle.zip \
    && awscli-bundle/install -i /usr/local/aws -b /usr/local/bin/aws \
    && rm awscli-bundle.zip \
    && rm -rf awscli-bundle

RUN mkdir -p /app
COPY app /app

EXPOSE 8080

ENTRYPOINT "/app/main"