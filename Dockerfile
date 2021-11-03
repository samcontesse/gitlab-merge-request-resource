FROM concourse/buildroot:git
COPY check /opt/resource/
COPY in /opt/resource/
COPY out /opt/resource/
