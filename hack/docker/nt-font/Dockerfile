FROM ghcr.io/ysicing/debian

RUN set -x \
    && apt update \
    && apt install --no-install-recommends --no-install-suggests -y cabextract xfonts-utils fontconfig ttf-mscorefonts-installer \
    && rm -rf /var/lib/apt/lists/*

RUN mkfontscale && mkfontdir && fc-cache