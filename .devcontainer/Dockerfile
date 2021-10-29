# See here for image contents: https://github.com/microsoft/vscode-dev-containers/tree/v0.163.1/containers/go/.devcontainer/base.Dockerfile
ARG VARIANT="1.17"
FROM mcr.microsoft.com/vscode/devcontainers/go:0-${VARIANT}

# [Optional] Install a version of Node.js using nvm for front end dev
ARG INSTALL_NODE="false"
ARG NODE_VERSION="lts/*"
RUN if [ "${INSTALL_NODE}" = "true" ]; then su vscode -c "source /usr/local/share/nvm/nvm.sh && nvm install ${NODE_VERSION} 2>&1"; fi

ENV GOPROXY="https://goproxy.cn,direct"

# [Optional] Uncomment this section to install additional OS packages.
RUN set -e \
    && sed -i "s/deb.debian.org/mirrors.aliyun.com/g" /etc/apt/sources.list \
    && sed -i "s/security.debian.org/mirrors.aliyun.com/g" /etc/apt/sources.list \
    && apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends protobuf-compiler

# [Optional] Uncomment the next line to use go get to install anything else you need
# RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26 \
#     && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1 \
#     && go get -u github.com/cweill/gotests/gotests

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1