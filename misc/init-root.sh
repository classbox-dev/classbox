#!/usr/bin/dumb-init /bin/sh

echo "execute \"$@\" as root"
exec "$@"
