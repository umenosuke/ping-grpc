FROM centos:7 as base

RUN yum update -y
RUN yum install -y wget unzip git

ARG GO_VERSION='1.14.1'
ARG GO_BINARY='go'"${GO_VERSION}"'.linux-amd64.tar.gz'
RUN wget https://dl.google.com/go/"${GO_BINARY}" -O /tmp/"${GO_BINARY}"
RUN tar -C /usr/local -xzf /tmp/"${GO_BINARY}"
ENV GOROOT=/usr/local/go
RUN echo 'export GOROOT="'"${GOROOT}"'"' >> /etc/profile.d/myGo.sh
ENV GOPATH=/usr/local/gopath
RUN mkdir -p ${GOPATH}
RUN echo 'export GOPATH="'"${GOPATH}"'"' >> /etc/profile.d/myGo.sh
RUN echo 'export PATH=${PATH}:${GOROOT}/bin:${GOPATH}/bin' >> /etc/profile.d/myGo.sh
ENV PATH=${PATH}:${GOROOT}/bin:${GOPATH}/bin

ARG PROTOC_VERSION="3.11.4"
ARG PROTOC_BINARY='protoc-'"${PROTOC_VERSION}"'-linux-x86_64.zip'
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v"${PROTOC_VERSION}"/"${PROTOC_BINARY}" -O /tmp/"${PROTOC_BINARY}"
RUN unzip -d /usr/local/protoc /tmp/"${PROTOC_BINARY}"
RUN chmod -R +x /usr/local/protoc
RUN echo 'export PATH=${PATH}:/usr/local/protoc/bin' >> /etc/profile.d/myProtoc.sh
ENV PATH=${PATH}:/usr/local/protoc/bin

ENV TZ=Asia/Tokyo


FROM base as go

ARG PROTOC_GEN_GO_VERSION="v1.3.4"
RUN go get -d -u github.com/golang/protobuf/protoc-gen-go
RUN cd "$(go env GOPATH)"/src/github.com/golang/protobuf && git checkout "${PROTOC_GEN_GO_VERSION}"
RUN go install github.com/golang/protobuf/protoc-gen-go


FROM base as tstypes

ARG PROTOC_GEN_TSTYPES_VERSION="0.4.0"
RUN go get github.com/tmc/grpcutil/protoc-gen-tstypes
RUN cd "$(go env GOPATH)"/src/github.com/tmc/grpcutil && git checkout "${PROTOC_GEN_TSTYPES_VERSION}"
RUN go install github.com/tmc/grpcutil/protoc-gen-tstypes
