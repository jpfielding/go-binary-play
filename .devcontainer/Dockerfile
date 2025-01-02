FROM rockylinux:8.5 as base

ARG REPO_NAME=go-play
ARG USERNAME=gopher
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN set -eux -o pipefail && \
    dnf install -y curl dnf-utils dnf-plugins-core epel-release git jq ncurses sudo which rsync && \
    dnf config-manager --set-enabled powertools && \ 
    dnf install -y ca-certificates cmake gcc gcc-c++ glib2-devel

# docker cli
RUN set -eux -o pipefail && \
    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo && \
    yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# envs
ENV INSTALL_PATH "/usr/local"
ENV PATH ${HOME}/bin:${PATH}

# Create the user
RUN groupadd --gid ${USER_GID} ${USERNAME} && \
  useradd --uid ${USER_UID} --gid ${USER_GID} -m ${USERNAME} -s /bin/zsh && \
  echo ${USERNAME} ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/${USERNAME} && \
  chmod 0440 /etc/sudoers.d/${USERNAME}

########### NON ROOT FROM HERE DOWN ###########
FROM base as user

USER ${USERNAME}

RUN set -eux && \
    sudo mkdir -p /workdir/${REPO_NAME} && \
    sudo chown ${USERNAME}:${USERNAME} /workdir

# envs
ENV HOME "/home/${USERNAME}"
ENV SDK_HOME ${HOME}/sdk
ENV PATH ${HOME}/bin:${PATH}

# go
ENV GO_VERSION "1.23.3"
ENV GOPATH ${SDK_HOME}/go
ENV PATH ${GOPATH}/bin:${HOME}/sdk/go:${PATH}
RUN set -eux && \
    export GO_ARCH="$(arch | sed 's/aarch64/arm64/' | sed 's/x86_64/amd64/')" && \
    cd /tmp && mkdir -p ${SDK_HOME} && \
    curl -o go.tar.gz -L https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz && \
    tar xzf go.tar.gz && mv go ${SDK_HOME} && rm /tmp/go.tar.gz && \
    cd ${SDK_HOME} && mv go go${GO_VERSION} && ln -s go${GO_VERSION} go
   
# pretty bash
COPY --chown=${USERNAME}:${USERNAME} .bash_pretty ${HOME}/.bash_pretty
RUN set -eux && \
    echo "source ${HOME}/.bash_pretty" >> ${HOME}/.bashrc

