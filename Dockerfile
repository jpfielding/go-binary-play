FROM rockylinux:8.5 as base

RUN set -eux -o pipefail && \
    dnf install -y curl dnf-utils dnf-plugins-core epel-release git jq ncurses sudo which rsync && \
    dnf config-manager --set-enabled powertools && \ 
    dnf install -y cmake gcc gcc-c++ glib2-devel

# docker cli
RUN set -eux -o pipefail && \
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo && \
    yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# pretty bash
RUN set -eux -o pipefail && \
    curl -L https://gist.githubusercontent.com/jpfielding/1c28345ea88decf10f5f62ff74a659fb/raw/709a4d4612927921f6e140c70dab4ca60dd8b7ce/.bashrc | grep -v '#!/bin/bash' >> /root/.bashrc 

RUN set -eux -o pipefail && \
    dnf install -y ca-certificates

# envs
ENV INSTALL_PATH "/usr/local"
ENV PATH ${HOME}/bin:${PATH}

# go
ENV GO_VERSION "1.21.0"
ENV PATH ${INSTALL_PATH}/go/bin:${GOPATH}/bin:${PATH}
RUN set -eux && \
    export GO_ARCH="$(arch | sed 's/aarch64/arm64/' | sed 's/x86_64/amd64/')" && \
    curl -o /tmp/go.tar.gz -L https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz && \
    tar xzf /tmp/go.tar.gz -C ${INSTALL_PATH} && \
    rm /tmp/go.tar.gz


